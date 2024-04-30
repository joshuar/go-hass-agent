// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system

import (
	"context"
	"encoding/json"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/godbus/dbus/v5"
	mqttapi "github.com/joshuar/go-hass-anything/v7/pkg/mqtt"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dbusControlTopic = "gohassagent/dbus"
)

type dbusControlMsg struct {
	Bus            string          `json:"bus"`
	Destination    string          `json:"destination"`
	Path           dbus.ObjectPath `json:"path"`
	Method         string          `json:"method"`
	Args           []any           `json:"args"`
	UseSessionPath bool            `json:"useSessionPath"`
}

func NewDBusControlSubscription(ctx context.Context) *mqttapi.Subscription {
	return &mqttapi.Subscription{
		Callback: func(_ MQTT.Client, msg MQTT.Message) {
			var dbusMsg dbusControlMsg

			if err := json.Unmarshal(msg.Payload(), &dbusMsg); err != nil {
				log.Warn().Err(err).Msg("could not unmarshal dbus MQTT message")
				return
			}

			if dbusMsg.UseSessionPath {
				dbusMsg.Path = dbusx.GetSessionPath(ctx)
			}

			dbusType, ok := dbusx.DbusTypeMap[dbusMsg.Bus]
			if !ok {
				log.Warn().Msg("unsupported dbus type")
				return
			}

			log.Info().
				Str("bus", dbusMsg.Bus).
				Str("destination", dbusMsg.Destination).
				Str("path", string(dbusMsg.Path)).
				Str("method", dbusMsg.Method).
				Msg("Dispatching D-Bus MQTT message")

			err := dbusx.NewBusRequest(ctx, dbusType).
				Path(dbusMsg.Path).
				Destination(dbusMsg.Destination).
				Call(dbusMsg.Method, dbusMsg.Args...)
			if err != nil {
				log.Warn().Err(err).
					Str("bus", dbusMsg.Bus).
					Str("destination", dbusMsg.Destination).
					Str("path", string(dbusMsg.Path)).
					Str("method", dbusMsg.Method).
					Msg("Error dispatching D-Bus command.")
			}
		},
		Topic: dbusControlTopic,
	}
}
