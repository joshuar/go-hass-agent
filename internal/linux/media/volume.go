// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/eclipse/paho.golang/paho"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/pulseaudiox"
)

const (
	muteIcon = "mdi:volume-mute"
	volIcon  = "mdi:knob"

	minVolpc  = 0
	maxVolpc  = 100
	volStepPc = 1

	audioControlPreferencesID = preferences.ControlsPrefPrefix + "media" + preferences.PathDelim + "audio"
)

var ErrInitAudioWorker = errors.New("could not init audio worker")

// audioControl is a struct containing the data for providing audio state
// tracking and control.
type audioControl struct {
	pulseAudio *pulseaudiox.PulseAudioClient
	msgCh      chan mqttapi.Msg
	muteEntity *mqtthass.SwitchEntity
	volEntity  *mqtthass.NumberEntity[int]
	logger     *slog.Logger
	prefs      *preferences.CommonWorkerPrefs
}

// entity is a convienience interface to treat all entities the same.
type entity interface {
	MarshalState(args ...any) (*mqttapi.Msg, error)
}

func (d *audioControl) PreferencesID() string {
	return audioControlPreferencesID
}

func (d *audioControl) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
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

func publishAudioState(msgCh chan mqttapi.Msg, entity entity) error {
	msg, err := entity.MarshalState()
	if err != nil {
		return fmt.Errorf("could not marshal entity state: %w", err)
	}
	msgCh <- *msg

	return nil
}

// newAudioControl will establish a connection to PulseAudio and return a
// audioControl object for tracking and controlling audio state.
func newAudioControl(ctx context.Context, msgCh chan mqttapi.Msg) (*audioControl, error) {
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

//nolint:lll
func VolumeControl(ctx context.Context, msgCh chan mqttapi.Msg, device *mqtthass.Device) (*mqtthass.NumberEntity[int], *mqtthass.SwitchEntity, error) {
	control, err := newAudioControl(ctx, msgCh)
	if err != nil {
		return nil, nil, errors.Join(ErrInitAudioWorker, err)
	}

	control.prefs, err = preferences.LoadWorker(control)
	if err != nil {
		return nil, nil, errors.Join(ErrInitAudioWorker, err)
	}

	if control.prefs.IsDisabled() {
		return nil, nil, nil
	}

	control.volEntity = mqtthass.NewNumberEntity[int]().
		WithMin(minVolpc).
		WithMax(maxVolpc).
		WithStep(volStepPc).
		WithMode(mqtthass.NumberSlider).
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Volume"),
			mqtthass.ID(device.Name+"_volume"),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon(volIcon),
		).
		WithState(
			mqtthass.StateCallback(control.volStateCallback),
			mqtthass.ValueTemplate("{{ value_json.value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(control.volCommandCallback),
		)

	control.muteEntity = mqtthass.NewSwitchEntity().
		OptimisticMode().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Mute"),
			mqtthass.ID(device.Name+"_mute"),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon(muteIcon),
		).
		WithState(
			mqtthass.StateCallback(control.muteStateCallback),
			mqtthass.ValueTemplate("{{ value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(control.muteCommandCallback),
		)

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

	return control.volEntity, control.muteEntity, nil
}
