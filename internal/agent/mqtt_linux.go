// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	mqtthass "github.com/joshuar/go-hass-anything/v3/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dbusDest           = "org.freedesktop.login1"
	dbusLockMethod     = dbusDest + ".Session.Lock"
	dbusUnlockMethod   = dbusDest + ".Session.UnLock"
	dbusRebootMethod   = dbusDest + ".Manager.Reboot"
	dbusPowerOffMethod = dbusDest + ".Manager.PowerOff"
)

func newMQTTObject(ctx context.Context) *mqttObj {
	appName := "go_hass_agent"

	baseEntity := func(entityID string) *mqtthass.EntityConfig {
		return mqtthass.NewEntityByID(entityID, appName).
			AsButton().
			WithDefaultOriginInfo().
			WithDeviceInfo(mqttDevice())
	}

	dbusCall := func(ctx context.Context, path dbus.ObjectPath, dest, method string, args ...any) error {
		return dbusx.NewBusRequest(ctx, dbusx.SystemBus).
			Path(path).
			Destination(dest).
			Call(method, args...)
	}

	sessionPath := dbusx.GetSessionPath(ctx)
	entities := make(map[string]*mqtthass.EntityConfig)
	entities["lock_session"] = baseEntity("lock_session").
		WithIcon("mdi:eye-lock").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := dbusCall(ctx, sessionPath, dbusDest, dbusLockMethod)
			if err != nil {
				log.Warn().Err(err).Msg("Could not lock session.")
			}
		})
	entities["unlock_session"] = baseEntity("unlock_session").
		WithIcon("mdi:eye-lock-open").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := dbusCall(ctx, sessionPath, dbusDest, dbusUnlockMethod)
			if err != nil {
				log.Warn().Err(err).Msg("Could not unlock session.")
			}
		})
	entities["reboot"] = baseEntity("reboot").
		WithIcon("mdi:restart").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := dbusCall(ctx, dbus.ObjectPath("/org/freedesktop/login1"), dbusDest, dbusRebootMethod, true)
			if err != nil {
				log.Warn().Err(err).Msg("Could not reboot session.")
			}
		})
	entities["poweroff"] = baseEntity("poweroff").
		WithIcon("mdi:power").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := dbusCall(ctx, dbus.ObjectPath("/org/freedesktop/login1"), dbusDest, dbusPowerOffMethod, true)
			if err != nil {
				log.Warn().Err(err).Msg("Could not power off session.")
			}
		})
	return &mqttObj{
		entities: entities,
	}
}

func mqttDevice() *mqtthass.Device {
	dev := linux.NewDevice(config.AppName, config.AppVersion)
	return &mqtthass.Device{
		Name:         dev.DeviceName(),
		URL:          config.AppURL,
		SWVersion:    dev.OsVersion(),
		Manufacturer: dev.Manufacturer(),
		Model:        dev.Model(),
		Identifiers:  []string{dev.DeviceID()},
	}
}
