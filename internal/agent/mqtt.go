// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	mqtthass "github.com/joshuar/go-hass-anything/v7/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v7/pkg/mqtt"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type mqttObj struct {
	entities []*mqtthass.EntityConfig
}

func (o *mqttObj) Name() string {
	return preferences.AppName
}

func (o *mqttObj) Configuration() []*mqttapi.Msg {
	var msgs []*mqttapi.Msg
	for _, c := range o.entities {
		if msg, err := mqtthass.MarshalConfig(c); err != nil {
			log.Error().Err(err).Msg("Failed to marshal payload for entity.")
		} else {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

func (o *mqttObj) Subscriptions() []*mqttapi.Subscription {
	var subs []*mqttapi.Subscription
	for _, v := range o.entities {
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

func (o *mqttObj) States() []*mqttapi.Msg {
	return nil
}
