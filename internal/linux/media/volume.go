// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	pulseaudiox "github.com/joshuar/go-hass-agent/pkg/linux/pulseaudio"
)

type audioDevice struct {
	pulseAudio *pulseaudiox.PulseAudioClient
	msgCh      chan *mqttapi.Msg
	muteEntity *mqtthass.SwitchEntity
	volEntity  *mqtthass.NumberEntity[int]
}

//nolint:exhaustruct,mnd
func VolumeControl(ctx context.Context, msgCh chan *mqttapi.Msg) (*mqtthass.NumberEntity[int], *mqtthass.SwitchEntity) {
	deviceInfo := device.MQTTDeviceInfo(ctx)

	client, err := pulseaudiox.NewPulseClient(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to connect to Pulseaudio. Volume control will be unavailable.")

		return nil, nil
	}

	log.Debug().Msg("Connected to pulseaudio.")

	audioDev := &audioDevice{
		pulseAudio: client,
		msgCh:      msgCh,
	}

	audioDev.volEntity = mqtthass.AsNumber(
		mqtthass.NewEntity(preferences.AppName, "Volume", deviceInfo.Name+"_volume").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(deviceInfo).
			WithIcon("mdi:knob").
			WithCommandCallback(audioDev.volCommandCallback).
			WithStateCallback(audioDev.volStateCallback).
			WithValueTemplate("{{ value_json.value }}"),
		1, 0, 100, mqtthass.NumberSlider)

	audioDev.muteEntity = mqtthass.AsSwitch(
		mqtthass.NewEntity(preferences.AppName, "Mute", deviceInfo.Name+"_mute").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(deviceInfo).
			WithIcon("mdi:volume-mute").
			WithCommandCallback(audioDev.muteCommandCallback).
			WithStateCallback(audioDev.muteStateCallback).
			WithValueTemplate("{{ value }}"),
		true)

	go func() {
		log.Debug().Msg("Monitoring pulseaudio for events.")
		audioDev.publishVolume()
		audioDev.publishMute()

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Closing pulseaudio connection.")

				return
			case <-client.EventCh:
				repl, err := client.GetState()
				if err != nil {
					log.Debug().Err(err).Msg("Failed to parse pulseaudio state.")

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
		log.Debug().Err(err).Msg("Could not retrieve current volume.")

		return
	}
	d.msgCh <- msg
}

func (d *audioDevice) volStateCallback(_ ...any) (json.RawMessage, error) {
	vol, err := d.pulseAudio.GetVolume()
	if err != nil {
		return json.RawMessage(`{ "value": 0 }`), err
	}

	log.Trace().Int("volume", int(vol)).Msg("Publishing volume change.")

	return json.RawMessage(`{ "value": ` + strconv.FormatFloat(vol, 'f', 0, 64) + ` }`), nil
}

func (d *audioDevice) volCommandCallback(p *paho.Publish) {
	if newValue, err := strconv.Atoi(string(p.Payload)); err != nil {
		log.Debug().Err(err).Msg("Could not parse new volume level.")
	} else {
		log.Trace().Int("volume", newValue).Msg("Received volume change from Home Assistant.")

		if err := d.pulseAudio.SetVolume(float64(newValue)); err != nil {
			log.Debug().Err(err).Msg("Could not set volume level.")

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
		log.Debug().Err(err).Msg("Could not set mute state.")
	}
}

func (d *audioDevice) publishMute() {
	msg, err := d.muteEntity.MarshalState()
	if err != nil {
		log.Debug().Msg("Could not retrieve mute state.")
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
