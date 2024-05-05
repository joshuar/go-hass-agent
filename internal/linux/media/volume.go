// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"encoding/json"
	"strconv"

	"mrogalski.eu/go/pulseaudio"

	"github.com/davecgh/go-spew/spew"
	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type audioDevice struct {
	pulseAudio *pulseaudio.Client
	msgCh      chan *mqttapi.Msg
	muteEntity *mqtthass.SwitchEntity
	volEntity  *mqtthass.NumberEntity[int]
}

func VolumeControl(ctx context.Context, msgCh chan *mqttapi.Msg) (*mqtthass.NumberEntity[int], *mqtthass.SwitchEntity) {
	device := linux.MQTTDevice()

	client, err := pulseaudio.NewClient()
	if err != nil {
		log.Warn().Err(err).Msg("Unable to connect to Pulseaudio. Volume control will be unavailable.")
		return nil, nil
	}

	audioDev := &audioDevice{
		pulseAudio: client,
		msgCh:      msgCh,
	}

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
			WithIcon("mdi:volume-mute").
			WithCommandCallback(audioDev.muteCommandCallback).
			WithStateCallback(audioDev.muteStateCallback).
			WithValueTemplate("{{ value }}"),
		true)
	spew.Dump(audioDev.muteEntity)

	if _, err := audioDev.getVolume(); err != nil {
		log.Warn().Err(err).Msg("Could not get volume.")
	}
	go func() {
		audioDev.publishVolume()
		audioDev.publishMute()
	}()

	go func() {
		events, err := client.Updates()
		if err != nil {
			log.Warn().Err(err).Msg("Cannot monitor Pulseaudio.")
			return
		}
		log.Debug().Msg("Monitoring pulseaudio for events.")
		for {
			select {
			case <-events:
				audioDev.publishVolume()
				audioDev.publishMute()
			case <-ctx.Done():
				return
			}
		}
	}()
	return audioDev.volEntity, audioDev.muteEntity
}

func (d *audioDevice) getVolume() (int, error) {
	v, err := d.pulseAudio.Volume()
	if err != nil {
		return 0, err
	}
	return int(v * 100), nil
}

func (d *audioDevice) setVolume(v int) error {
	newVol := float32(v) / 100
	return d.pulseAudio.SetVolume(newVol)
}

func (d *audioDevice) publishVolume() {
	msg, err := d.volEntity.MarshalState()
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve current volume.")
		return
	}
	d.msgCh <- msg
}

func (d *audioDevice) volStateCallback(_ ...any) (json.RawMessage, error) {
	vol, err := d.getVolume()
	if err != nil {
		return json.RawMessage(`{ "value": 0 }`), err
	}
	return json.RawMessage(`{ "value": ` + strconv.Itoa(vol) + ` }`), nil
}

func (d *audioDevice) volCommandCallback(p *paho.Publish) {
	if newValue, err := strconv.Atoi(string(p.Payload)); err != nil {
		log.Warn().Err(err).Msg("Could not parse new volume level.")
	} else {
		log.Trace().Int("volume", newValue).Msg("Received volume change from Home Assistant.")
		if err := d.setVolume(newValue); err != nil {
			log.Warn().Err(err).Msg("Could not set volume level.")
			return
		}
		go func() {
			d.publishVolume()
		}()
	}
}

func (d *audioDevice) setMute(v bool) {
	var err error
	switch v {
	case true:
		err = d.pulseAudio.SetMute(true)
	case false:
		err = d.pulseAudio.SetMute(false)
	}
	if err != nil {
		log.Warn().Err(err).Msg("Could not set mute state.")
	}
}

func (d *audioDevice) publishMute() {
	msg, err := d.muteEntity.MarshalState()
	if err != nil {
		log.Warn().Msg("Could not retrieve mute state.")
	} else {
		d.msgCh <- msg
	}
}

func (d *audioDevice) muteStateCallback(_ ...any) (json.RawMessage, error) {
	muteState, err := d.pulseAudio.Mute()
	if err != nil {
		return json.RawMessage(`OFF`), err
	}
	switch muteState {
	case true:
		return json.RawMessage(`ON`), nil
	default:
		return json.RawMessage(`OFF`), nil
	}
}

func (d *audioDevice) muteCommandCallback(p *paho.Publish) {
	spew.Dump(p)
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
