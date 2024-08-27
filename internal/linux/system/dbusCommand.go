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
	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

type dbusCommandMsg struct {
	Bus            string `json:"bus"`
	Destination    string `json:"destination"`
	Path           string `json:"path"`
	Method         string `json:"method"`
	Args           []any  `json:"args"`
	UseSessionPath bool   `json:"use_session_path"`
}

//nolint:lll
func NewDBusCommandSubscription(ctx context.Context, api *dbusx.DBusAPI, parentLogger *slog.Logger, device *mqtthass.Device) *mqttapi.Subscription {
	logger := parentLogger.With(slog.String("controller", "dbus_command"))

	sessionBus, err := api.GetBus(ctx, dbusx.SessionBus)
	if err != nil {
		logger.Warn("Cannot connect to session bus.", slog.Any("error", err))

		return nil
	}

	systemBus, err := api.GetBus(ctx, dbusx.SessionBus)
	if err != nil {
		logger.Warn("Cannot connect to system bus.", slog.Any("error", err))

		return nil
	}

	busMap := map[string]*dbusx.Bus{"session": sessionBus, "system": systemBus}

	return &mqttapi.Subscription{
		Callback: func(p *paho.Publish) {
			var dbusMsg dbusCommandMsg

			// Unmarshal the request.
			if err = json.Unmarshal(p.Payload, &dbusMsg); err != nil {
				logger.Error("Could not unmarshal D-Bus MQTT message.", slog.Any("error", err))

				return
			}
			// Check which bus type was requested.
			bus, ok := busMap[dbusMsg.Bus]
			if !ok {
				logger.Error("Unsupported D-Bus type.")

				return
			}
			// Fetch the session path if requested.
			if dbusMsg.UseSessionPath {
				dbusMsg.Path, err = busMap["session"].GetSessionPath()
				if err != nil {
					logger.Error("Could not determine session path.", slog.Any("error", err))

					return
				}
			}

			logger.With(
				slog.String("bus", dbusMsg.Bus),
				slog.String("destination", dbusMsg.Destination),
				slog.String("path", dbusMsg.Path),
				slog.String("method", dbusMsg.Method),
			).Info("Dispatching D-Bus command.")

			// Call the method.
			err = dbusx.NewMethod(bus, dbusMsg.Destination, dbusMsg.Path, dbusMsg.Method).Call(ctx, dbusMsg.Args...)
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
		Topic: "gohassagent/" + device.Name + "/dbuscommand",
	}
}
