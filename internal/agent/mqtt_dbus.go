// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"encoding/json"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"

	mqttapi "github.com/joshuar/go-hass-anything/v5/pkg/mqtt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	dbus_mqtt_topic = "gohassagent/dbus"
)

type DbusMessage struct {
	Bus            string          `json:"bus"`
	Destination    string          `json:"destination"`
	Path           dbus.ObjectPath `json:"path"`
	UseSessionPath bool            `json:"useSessionPath"`
	Method         string          `json:"method"`
	Args           []any           `json:"args"`
}

func newDbusSubscription(ctx context.Context) *mqttapi.Subscription {
	return &mqttapi.Subscription{
		Callback: func(c MQTT.Client, m MQTT.Message) {
			var dbusMsg DbusMessage

			if err := json.Unmarshal(m.Payload(), &dbusMsg); err != nil {
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

			log.Info().Str("bus", dbusMsg.Bus).Str("destination", dbusMsg.Destination).Str("path", string(dbusMsg.Path)).Str("method", dbusMsg.Method).Msg("dispatching dbus MQTT message")

			dbusx.NewBusRequest(ctx, dbusType).
				Path(dbus.ObjectPath(dbusMsg.Path)).
				Destination(dbusMsg.Destination).
				Call(dbusMsg.Method, dbusMsg.Args...)
		},
		Topic: dbus_mqtt_topic,
	}
}
