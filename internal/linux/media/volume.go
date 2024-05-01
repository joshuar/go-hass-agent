// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"mrogalski.eu/go/pulseaudio"

	mqtthass "github.com/joshuar/go-hass-anything/v7/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v7/pkg/mqtt"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/linux"
)

type audioDevice struct {
	pulseAudio *pulseaudio.Client
	topics     *mqtthass.Topics
	msgCh      chan mqttapi.Msg
	volume     float64
}

func VolumeControl(ctx context.Context, msgCh chan mqttapi.Msg) []*mqtthass.EntityConfig {
	var entities []*mqtthass.EntityConfig
	client, err := pulseaudio.NewClient()
	if err != nil {
		log.Warn().Err(err).Msg("Unable to connect to Pulseaudio. Volume control will be unavailable.")
		return nil
	}

	audioDev := &audioDevice{
		pulseAudio: client,
		msgCh:      msgCh,
	}

	volCtrl := linux.NewSlider("volume", 1, 0, 100).
		WithIcon("mdi:knob").
		WithCommandCallback(audioDev.parseVolume).
		WithValueTemplate("{{ value_json.value }}")
	audioDev.topics = volCtrl.GetTopics()
	entities = append(entities, volCtrl)

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
			entities = nil
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
	return entities
}

func (d *audioDevice) getVolume() (bool, error) {
	newVol, err := d.pulseAudio.Volume()
	if err != nil {
		return false, err
	}
	if newVol != float32(d.volume) {
		d.volume = float64(newVol * 100)
		return true, nil
	}
	return false, nil
}

func (d *audioDevice) setVolume(v float64) error {
	if err := d.pulseAudio.SetVolume(float32(v)); err != nil {
		return err
	}
	d.volume = v
	return nil
}

func (d *audioDevice) publishVolume() {
	msg := mqttapi.NewMsg(d.topics.State, []byte(`{ "value": `+strconv.FormatFloat(d.volume, 'f', -1, 64)+` }`))
	d.msgCh <- *msg
}

func (d *audioDevice) parseVolume(_ MQTT.Client, msg MQTT.Message) {
	if newValue, err := strconv.ParseFloat(string(msg.Payload()), 64); err != nil {
		log.Warn().Err(err).Msg("Could not parse new volume level.")
	} else {
		log.Trace().Float64("volume", newValue).Msg("Received volume change from Home Assistant.")
		if err := d.setVolume(newValue / 100); err != nil {
			log.Warn().Err(err).Msg("Could not set volume level.")
			return
		}
		go func() {
			d.publishVolume()
		}()
	}
}
