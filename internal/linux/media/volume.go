// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"strconv"

	"mrogalski.eu/go/pulseaudio"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type audioDevice struct {
	pulseAudio *pulseaudio.Client
	stateTopic string
	msgCh      chan *mqttapi.Msg
	volume     int
}

func VolumeControl(ctx context.Context, msgCh chan *mqttapi.Msg) *mqtthass.NumberEntity[int] {
	device := linux.MQTTDevice()

	client, err := pulseaudio.NewClient()
	if err != nil {
		log.Warn().Err(err).Msg("Unable to connect to Pulseaudio. Volume control will be unavailable.")
		return nil
	}

	audioDev := &audioDevice{
		pulseAudio: client,
		msgCh:      msgCh,
	}

	volCtrl := mqtthass.AsNumber(
		mqtthass.NewEntity(preferences.AppName, "Volume", device.Name+"_volume").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(device).
			WithIcon("mdi:knob").
			WithCommandCallback(audioDev.parseVolume).
			WithValueTemplate("{{ value_json.value }}"),
		1, 0, 100, mqtthass.NumberSlider)
	audioDev.stateTopic = volCtrl.StateTopic
	// muteCtl := linux.NewToggle("volume_mute").
	// 	WithIcon("mdi:volume-mute").
	// 	WithCommandCallback(audioDev.parseVolume).
	// 	WithValueTemplate("{{ value_json.value }}")
	// audioDev.topics = volCtrl.GetTopics()
	// entities = append(entities, volCtrl)

	if _, err := audioDev.getVolume(); err != nil {
		log.Warn().Err(err).Msg("Could not get volume.")
	}
	go func() {
		audioDev.publishVolume()
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
				changed, err := audioDev.getVolume()
				if err != nil {
					log.Warn().Err(err).Msg("Could not get volume.")
				}
				if changed {
					audioDev.publishVolume()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return volCtrl
}

func (d *audioDevice) getVolume() (bool, error) {
	v, err := d.pulseAudio.Volume()
	newVol := int(v * 100)
	if err != nil {
		return false, err
	}
	if newVol != d.volume {
		d.volume = newVol
		return true, nil
	}
	return false, nil
}

func (d *audioDevice) setVolume(v int) error {
	newVol := float32(v) / 100
	if err := d.pulseAudio.SetVolume(newVol); err != nil {
		return err
	}
	d.volume = v
	return nil
}

func (d *audioDevice) publishVolume() {
	msg := mqttapi.NewMsg(d.stateTopic, []byte(`{ "value": `+strconv.Itoa(d.volume)+` }`))
	d.msgCh <- msg
}

func (d *audioDevice) parseVolume(p *paho.Publish) {
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
