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

func NewDBusCommandSubscription(ctx context.Context, api *dbusx.DBusAPI, parentLogger *slog.Logger) *mqttapi.Subscription {
	logger := parentLogger.With(slog.String("controller", "dbus_command"))

	sessionBus, err := api.GetBus(ctx, dbusx.SessionBus)
	if err != nil {
		logger.Warn("Cannot create D-Bus command listener.", "error", err.Error())

		return nil
	}

	systemBus, err := api.GetBus(ctx, dbusx.SessionBus)
	if err != nil {
		logger.Warn("Cannot create D-Bus command listener.", "error", err.Error())

		return nil
	}

	return &mqttapi.Subscription{
		Callback: func(p *paho.Publish) {
			var dbusMsg dbusCommandMsg

			if err := json.Unmarshal(p.Payload, &dbusMsg); err != nil {
				logger.Warn("Could not unmarshal D-Bus MQTT message.", "error", err.Error())

				return
			}

			if dbusMsg.UseSessionPath {
				var err error

				dbusMsg.Path, err = sessionBus.GetSessionPath(ctx)
				if err != nil {
					logger.Warn("Could not determine session path.", "error", err.Error())

					return
				}
			}

			dbusType, ok := dbusx.DbusTypeMap[dbusMsg.Bus]
			if !ok {
				logger.Warn("Unsupported D-Bus type.")

				return
			}
			logger.With(
				slog.String("bus", dbusMsg.Bus),
				slog.String("destination", dbusMsg.Destination),
				slog.String("path", dbusMsg.Path),
				slog.String("method", dbusMsg.Method),
			).Info("Dispatching D-Bus command to MQTT.")

			var bus *dbusx.Bus
			switch dbusType {
			case dbusx.SessionBus:
				bus = sessionBus
			case dbusx.SystemBus:
				bus = systemBus
			}

			err := bus.Call(ctx, dbusMsg.Path, dbusMsg.Destination, dbusMsg.Method, dbusMsg.Args...)
			if err != nil {
				logger.With(
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
