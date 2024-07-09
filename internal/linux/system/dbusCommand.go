// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/eclipse/paho.golang/paho"
	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dbusCommandTopic = "gohassagent/dbus"
)

type dbusCommandMsg struct {
	Bus            string `json:"bus"`
	Destination    string `json:"destination"`
	Path           string `json:"path"`
	Method         string `json:"method"`
	Args           []any  `json:"args"`
	UseSessionPath bool   `json:"use_session_path"`
}

func NewDBusCommandSubscription(ctx context.Context) *mqttapi.Subscription {
	return &mqttapi.Subscription{
		Callback: func(p *paho.Publish) {
			var dbusMsg dbusCommandMsg

			if err := json.Unmarshal(p.Payload, &dbusMsg); err != nil {
				logging.FromContext(ctx).Warn("Could not unmarshal D-Bus MQTT message.", "error", err.Error())

				return
			}

			if dbusMsg.UseSessionPath {
				var err error

				dbusMsg.Path, err = dbusx.GetSessionPath(ctx)
				if err != nil {
					logging.FromContext(ctx).Warn("Could not determine session path.", "error", err.Error())

					return
				}
			}

			dbusType, ok := dbusx.DbusTypeMap[dbusMsg.Bus]
			if !ok {
				logging.FromContext(ctx).Warn("Unsupported D-Bus type.")

				return
			}

			logging.FromContext(ctx).With(
				slog.String("bus", dbusMsg.Bus),
				slog.String("destination", dbusMsg.Destination),
				slog.String("path", dbusMsg.Path),
				slog.String("method", dbusMsg.Method),
			).Info("Dispatching D-Bus command to MQTT.")

			err := dbusx.Call(ctx, dbusType, dbusMsg.Path, dbusMsg.Destination, dbusMsg.Method, dbusMsg.Args...)
			if err != nil {
				logging.FromContext(ctx).With(
					slog.String("bus", dbusMsg.Bus),
					slog.String("destination", dbusMsg.Destination),
					slog.String("path", dbusMsg.Path),
					slog.String("method", dbusMsg.Method),
					slog.Any("error", err),
				).Warn("Error dispatching D-Bus command.")
			}
		},
		Topic: dbusCommandTopic,
	}
}
