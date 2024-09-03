// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/eclipse/paho.golang/paho"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/pulseaudiox"
)

const (
	muteIcon = "mdi:volume-mute"
	volIcon  = "mdi:knob"

	minVolpc  = 0
	maxVolpc  = 100
	volStepPc = 1
)

// audioControl is a struct containing the data for providing audio state
// tracking and control.
type audioControl struct {
	pulseAudio *pulseaudiox.PulseAudioClient
	msgCh      chan *mqttapi.Msg
	muteEntity *mqtthass.SwitchEntity
	volEntity  *mqtthass.NumberEntity[int]
	logger     *slog.Logger
}

// entity is a convienience interface to treat all entities the same.
type entity interface {
	MarshalState(args ...any) (*mqttapi.Msg, error)
}

//nolint:lll
func VolumeControl(ctx context.Context, msgCh chan *mqttapi.Msg, device *mqtthass.Device) (*mqtthass.NumberEntity[int], *mqtthass.SwitchEntity) {
	control, err := newAudioControl(ctx, msgCh)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not configure Pulseaudio. Volume control will not be available.", slog.Any("error", err))

		return nil, nil
	}

	control.volEntity = mqtthass.AsNumber(
		mqtthass.NewEntity(preferences.AppName, "Volume", device.Name+"_volume").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(device).
			WithIcon(volIcon).
			WithCommandCallback(control.volCommandCallback).
			WithStateCallback(control.volStateCallback).
			WithValueTemplate("{{ value_json.value }}"),
		volStepPc, minVolpc, maxVolpc, mqtthass.NumberSlider)

	control.muteEntity = mqtthass.AsSwitch(
		mqtthass.NewEntity(preferences.AppName, "Mute", device.Name+"_mute").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(device).
			WithIcon(muteIcon).
			WithCommandCallback(control.muteCommandCallback).
			WithStateCallback(control.muteStateCallback).
			WithValueTemplate("{{ value }}"),
		false).AsTypeSwitch()

	update := func() { // Pulseaudio changed state. Get the new state.
		// Publish and update mute state if it changed.
		if err := publishAudioState(msgCh, control.muteEntity); err != nil {
			control.logger.Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
		}

		// Publish and update volume if it changed.
		if err := publishAudioState(msgCh, control.volEntity); err != nil {
			control.logger.Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
		}
	}

	// Process Pulseaudio state updates as they are received.
	go func() {
		control.logger.Debug("Monitoring for events.")
		update()

		for {
			select {
			case <-ctx.Done():
				control.logger.Debug("Closing connection.")

				return
			case <-control.pulseAudio.EventCh:
				update()
			}
		}
	}()

	return control.volEntity, control.muteEntity
}

// newAudioControl will establish a connection to PulseAudio and return a
// audioControl object for tracking and controlling audio state.
func newAudioControl(ctx context.Context, msgCh chan *mqttapi.Msg) (*audioControl, error) {
	client, err := pulseaudiox.NewPulseClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Pulseaudio: %w", err)
	}

	audioDev := &audioControl{
		pulseAudio: client,
		msgCh:      msgCh,
		logger:     logging.FromContext(ctx).WithGroup("volume_controller"),
	}

	return audioDev, nil
}

// volStateCallback is executed when the volume is read on MQTT.
func (d *audioControl) volStateCallback(_ ...any) (json.RawMessage, error) {
	vol, err := d.pulseAudio.GetVolume()
	if err != nil {
		return json.RawMessage(`{ "value": 0 }`), err
	}

	d.logger.Debug("Publishing volume change.", slog.Int("volume", int(vol)))

	return json.RawMessage(`{ "value": ` + strconv.FormatFloat(vol, 'f', 0, 64) + ` }`), nil
}

// volCommandCallback is called when the volume is changed on MQTT.
func (d *audioControl) volCommandCallback(p *paho.Publish) {
	if newValue, err := strconv.Atoi(string(p.Payload)); err != nil {
		d.logger.Debug("Could not parse new volume level.", slog.Any("error", err))
	} else {
		d.logger.Debug("Received volume change from Home Assistant.", slog.Int("volume", newValue))

		if err := d.pulseAudio.SetVolume(float64(newValue)); err != nil {
			d.logger.Error("Could not set volume level.", slog.Any("error", err))

			return
		}

		go func() {
			if err := publishAudioState(d.msgCh, d.volEntity); err != nil {
				d.logger.Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
			}
		}()
	}
}

// muteStateCallback is executed when the mute state is read on MQTT.
func (d *audioControl) muteStateCallback(_ ...any) (json.RawMessage, error) {
	muteVal, err := d.pulseAudio.GetMute()
	if err != nil {
		return json.RawMessage(`OFF`), err
	}

	switch muteVal {
	case true:
		return json.RawMessage(`ON`), nil
	default:
		return json.RawMessage(`OFF`), nil
	}
}

// muteCommandCallback is executed when the mute state is changed on MQTT.
func (d *audioControl) muteCommandCallback(p *paho.Publish) {
	var err error

	state := string(p.Payload)
	switch state {
	case "ON":
		err = d.pulseAudio.SetMute(true)
	case "OFF":
		err = d.pulseAudio.SetMute(false)
	}

	if err != nil {
		d.logger.Error("Could not set mute state.", slog.Any("error", err))

		return
	}

	go func() {
		if err := publishAudioState(d.msgCh, d.muteEntity); err != nil {
			d.logger.Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
		}
	}()
}

func publishAudioState(msgCh chan *mqttapi.Msg, entity entity) error {
	msg, err := entity.MarshalState()
	if err != nil {
		return fmt.Errorf("could not marshal entity state: %w", err)
	}
	msgCh <- msg

	return nil
}
