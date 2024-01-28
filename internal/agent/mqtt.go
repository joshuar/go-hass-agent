// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	mqtthass "github.com/joshuar/go-hass-anything/v3/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v3/pkg/mqtt"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
)

type mqttDevice struct {
	configs map[string]*mqtthass.EntityConfig
}

func (d *mqttDevice) Name() string {
	return config.AppName
}

func (d *mqttDevice) Configuration() []*mqttapi.Msg {
	var msgs []*mqttapi.Msg

	for id, c := range d.configs {
		if msg, err := mqtthass.MarshalConfig(c); err != nil {
			log.Error().Err(err).Msgf("Failed to marshal payload for %s.", id)
		} else {
			msgs = append(msgs, msg)
		}
	}

	return msgs
}

func (d *mqttDevice) Subscriptions() []*mqttapi.Subscription {
	var subs []*mqttapi.Subscription
	for _, v := range d.configs {
		if v.CommandCallback != nil {
			if sub, err := mqtthass.MarshalSubscription(v); err != nil {
				log.Error().Err(err).Str("entity", v.Entity.Name).
					Msg("Error adding subscription.")
			} else {
				subs = append(subs, sub)
			}
		}
	}
	return subs
}

func (d *mqttDevice) States() []*mqttapi.Msg {
	return nil
}
