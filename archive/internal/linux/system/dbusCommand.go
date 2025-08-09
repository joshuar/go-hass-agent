// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/eclipse/paho.golang/paho"
	slogctx "github.com/veqryn/slog-context"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dbusCmdPreferencesID = controlsPrefPrefix + "dbus_commands"
)

var ErrInitDBusCommands = errors.New("could not init D-Bus commands worker")

type dbusCommandMsg struct {
	Bus            string `json:"bus"`
	Destination    string `json:"destination"`
	Path           string `json:"path"`
	Method         string `json:"method"`
	Args           []any  `json:"args"`
	UseSessionPath bool   `json:"use_session_path"`
}

type dbusCmdWorker struct{}

func (w *dbusCmdWorker) PreferencesID() string {
	return dbusCmdPreferencesID
}

func (w *dbusCmdWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewDBusCommandSubscription(ctx context.Context, device *mqtthass.Device) (*mqttapi.Subscription, error) {
	worker := &dbusCmdWorker{}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitDBusCommands, err)
	}

	//nolint:nilnil
	if prefs.IsDisabled() {
		return nil, nil
	}

	systemBus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitDBusCommands, linux.ErrNoSystemBus)
	}

	sessionBus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitDBusCommands, linux.ErrNoSessionBus)
	}

	busMap := map[string]*dbusx.Bus{"session": sessionBus, "system": systemBus}

	return &mqttapi.Subscription{
			Callback: func(packet *paho.Publish) {
				var (
					dbusMsg dbusCommandMsg
					err     error
				)

				logger := slogctx.FromCtx(ctx)

				// Unmarshal the request.
				if err = json.Unmarshal(packet.Payload, &dbusMsg); err != nil {
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
					logger.Warn("Error dispatching D-Bus command.", slog.Any("error", err))
				}
			},
			Topic: "gohassagent/" + device.Name + "/dbuscommand",
		},
		nil
}
