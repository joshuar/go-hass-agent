// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/preferences"
	pulseaudiox "github.com/joshuar/go-hass-agent/pkg/linux/pulseaudio"
)

type audioDevice struct {
	pulseAudio *pulseaudiox.PulseAudioClient
	msgCh      chan *mqttapi.Msg
	muteEntity *mqtthass.SwitchEntity
	volEntity  *mqtthass.NumberEntity[int]
	logger     *slog.Logger
}

//nolint:exhaustruct,mnd,lll
func VolumeControl(ctx context.Context, msgCh chan *mqttapi.Msg, parentLogger *slog.Logger, device *mqtthass.Device) (*mqtthass.NumberEntity[int], *mqtthass.SwitchEntity) {
	logger := parentLogger.With(slog.String("controller", "volume"))

	client, err := pulseaudiox.NewPulseClient(ctx)
	if err != nil {
		logger.Warn("Unable to connect to Pulseaudio. Volume control will be unavailable.", "error", err.Error())

		return nil, nil
	}

	audioDev := &audioDevice{
		pulseAudio: client,
		msgCh:      msgCh,
		logger:     logger,
	}

	audioDev.logger.Debug("Connected.")

	audioDev.volEntity = mqtthass.AsNumber(
		mqtthass.NewEntity(preferences.AppName, "Volume", device.Name+"_volume").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(device).
			WithIcon("mdi:knob").
			WithCommandCallback(audioDev.volCommandCallback).
			WithStateCallback(audioDev.volStateCallback).
			WithValueTemplate("{{ value_json.value }}"),
		1, 0, 100, mqtthass.NumberSlider)

	audioDev.muteEntity = mqtthass.AsSwitch(
		mqtthass.NewEntity(preferences.AppName, "Mute", device.Name+"_mute").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(device).
			WithIcon("mdi:volume-mute").
			WithCommandCallback(audioDev.muteCommandCallback).
			WithStateCallback(audioDev.muteStateCallback).
			WithValueTemplate("{{ value }}"),
		true)

	go func() {
		audioDev.logger.Debug("Monitoring for events.")
		audioDev.publishVolume()
		audioDev.publishMute()

		for {
			select {
			case <-ctx.Done():
				audioDev.logger.Debug("Closing connection.")

				return
			case <-client.EventCh:
				repl, err := client.GetState()
				if err != nil {
					audioDev.logger.Debug("Failed to retrieve pulseaudio state.", "error", err.Error())

					continue
				}

				volPct := pulseaudiox.ParseVolume(repl)

				switch {
				case repl.Mute != client.Mute:
					audioDev.publishMute()
					audioDev.pulseAudio.Mute = repl.Mute
				case volPct != client.Vol:
					audioDev.pulseAudio.Vol = volPct
					audioDev.publishVolume()
				}
			}
		}
	}()

	return audioDev.volEntity, audioDev.muteEntity
}

func (d *audioDevice) publishVolume() {
	msg, err := d.volEntity.MarshalState()
	if err != nil {
		d.logger.Error("Could not retrieve current volume.", "error", err.Error())

		return
	}
	d.msgCh <- msg
}

func (d *audioDevice) volStateCallback(_ ...any) (json.RawMessage, error) {
	vol, err := d.pulseAudio.GetVolume()
	if err != nil {
		return json.RawMessage(`{ "value": 0 }`), err
	}

	d.logger.Debug("Publishing volume change.", "volume", int(vol))

	return json.RawMessage(`{ "value": ` + strconv.FormatFloat(vol, 'f', 0, 64) + ` }`), nil
}

func (d *audioDevice) volCommandCallback(p *paho.Publish) {
	if newValue, err := strconv.Atoi(string(p.Payload)); err != nil {
		d.logger.Debug("Could not parse new volume level.", "error", err.Error())
	} else {
		d.logger.Debug("Received volume change from Home Assistant.", "volume", newValue)

		if err := d.pulseAudio.SetVolume(float64(newValue)); err != nil {
			d.logger.Error("Could not set volume level.", "error", err.Error())

			return
		}

		go func() {
			d.publishVolume()
		}()
	}
}

func (d *audioDevice) setMute(muteVal bool) {
	var err error

	switch muteVal {
	case true:
		err = d.pulseAudio.SetMute(true)
	case false:
		err = d.pulseAudio.SetMute(false)
	}

	if err != nil {
		d.logger.Error("Could not set mute state.", "error", err.Error())
	}
}

func (d *audioDevice) publishMute() {
	msg, err := d.muteEntity.MarshalState()
	if err != nil {
		d.logger.Error("Could retrieve mute state.", "error", err.Error())
	} else {
		d.msgCh <- msg
	}
}

func (d *audioDevice) muteStateCallback(_ ...any) (json.RawMessage, error) {
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

func (d *audioDevice) muteCommandCallback(p *paho.Publish) {
	state := string(p.Payload)
	switch state {
	case "ON":
		d.setMute(true)
	case "OFF":
		d.setMute(false)
	}

	go func() {
		d.publishMute()
	}()
}
