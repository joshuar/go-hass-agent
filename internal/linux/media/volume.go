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
	slogctx "github.com/veqryn/slog-context"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

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

// VolumeWorker is a struct containing the data for providing audio state
// tracking and control.
type VolumeWorker struct {
	pulseAudio    *pulseaudiox.PulseAudioClient
	MsgCh         chan mqttapi.Msg
	MuteControl   *mqtthass.SwitchEntity
	VolumeControl *mqtthass.NumberEntity[int]
	*preferences.CommonWorkerPrefs
}

// entity is a convienience interface to treat all entities the same.
type entity interface {
	MarshalState(args ...any) (*mqttapi.Msg, error)
}

func (d *VolumeWorker) PreferencesID() string {
	return audioControlPreferencesID
}

func (d *VolumeWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

// volStateCallback is executed when the volume is read on MQTT.
func (d *VolumeWorker) volStateCallback(_ ...any) (json.RawMessage, error) {
	vol, err := d.pulseAudio.GetVolume()
	if err != nil {
		return json.RawMessage(`{ "value": 0 }`), err
	}

	slog.Debug("Publishing volume change.", slog.Int("volume", int(vol)))

	return json.RawMessage(`{ "value": ` + strconv.FormatFloat(vol, 'f', 0, 64) + ` }`), nil
}

// volCommandCallback is called when the volume is changed on MQTT.
func (d *VolumeWorker) volCommandCallback(p *paho.Publish) {
	if newValue, err := strconv.Atoi(string(p.Payload)); err != nil {
		slog.Debug("Could not parse new volume level.", slog.Any("error", err))
	} else {
		slog.Debug("Received volume change from Home Assistant.", slog.Int("volume", newValue))

		if err := d.pulseAudio.SetVolume(float64(newValue)); err != nil {
			slog.Error("Could not set volume level.", slog.Any("error", err))
			return
		}

		go func() {
			if err := publishAudioState(d.MsgCh, d.VolumeControl); err != nil {
				slog.Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
			}
		}()
	}
}

// muteStateCallback is executed when the mute state is read on MQTT.
func (d *VolumeWorker) muteStateCallback(_ ...any) (json.RawMessage, error) {
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
func (d *VolumeWorker) muteCommandCallback(p *paho.Publish) {
	var err error

	state := string(p.Payload)
	switch state {
	case "ON":
		err = d.pulseAudio.SetMute(true)
	case "OFF":
		err = d.pulseAudio.SetMute(false)
	}

	if err != nil {
		slog.Error("Could not set mute state.", slog.Any("error", err))

		return
	}

	go func() {
		if err := publishAudioState(d.MsgCh, d.MuteControl); err != nil {
			slog.Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
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
func newAudioControl(ctx context.Context) (*VolumeWorker, error) {
	client, err := pulseaudiox.NewPulseClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Pulseaudio: %w", err)
	}

	audioDev := &VolumeWorker{
		pulseAudio: client,
		MsgCh:      make(chan mqttapi.Msg),
	}

	return audioDev, nil
}

//nolint:nilnil
func NewVolumeWorker(ctx context.Context, device *mqtthass.Device) (*VolumeWorker, error) {
	worker, err := newAudioControl(ctx)
	if err != nil {
		return nil, errors.Join(ErrInitAudioWorker, err)
	}

	worker.CommonWorkerPrefs, err = preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitAudioWorker, err)
	}

	if worker.IsDisabled() {
		return nil, nil
	}

	worker.VolumeControl = mqtthass.NewNumberEntity[int]().
		WithMin(minVolpc).
		WithMax(maxVolpc).
		WithStep(volStepPc).
		WithMode(mqtthass.NumberSlider).
		WithDetails(
			mqtthass.App(preferences.AppName+"_"+device.Name),
			mqtthass.Name("Volume"),
			mqtthass.ID(device.Name+"_volume"),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon(volIcon),
		).
		WithState(
			mqtthass.StateCallback(worker.volStateCallback),
			mqtthass.ValueTemplate("{{ value_json.value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(worker.volCommandCallback),
		)

	worker.MuteControl = mqtthass.NewSwitchEntity().
		OptimisticMode().
		WithDetails(
			mqtthass.App(preferences.AppName+"_"+device.Name),
			mqtthass.Name("Mute"),
			mqtthass.ID(device.Name+"_mute"),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon(muteIcon),
		).
		WithState(
			mqtthass.StateCallback(worker.muteStateCallback),
			mqtthass.ValueTemplate("{{ value }}"),
		).
		WithCommand(
			mqtthass.CommandCallback(worker.muteCommandCallback),
		)

	update := func() { // Pulseaudio changed state. Get the new state.
		// Publish and update mute state if it changed.
		if err := publishAudioState(worker.MsgCh, worker.MuteControl); err != nil {
			slogctx.FromCtx(ctx).Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
		}

		// Publish and update volume if it changed.
		if err := publishAudioState(worker.MsgCh, worker.VolumeControl); err != nil {
			slogctx.FromCtx(ctx).Error("Failed to publish mute state to MQTT.", slog.Any("error", err))
		}
	}

	// Process Pulseaudio state updates as they are received.
	go func() {
		slogctx.FromCtx(ctx).Debug("Monitoring for events.")
		update()

		for {
			select {
			case <-ctx.Done():
				slogctx.FromCtx(ctx).Debug("Closing connection.")

				return
			case <-worker.pulseAudio.EventCh:
				if ctx.Err() == nil {
					update()
				}
			}
		}
	}()

	return worker, nil
}
