// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"

	mqtthass "github.com/joshuar/go-hass-anything/v3/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

func (a *Agent) newMQTTDevice(ctx context.Context) *mqttDevice {
	cfg := config.FetchFromContext(ctx)

	var deviceName string
	cfg.Get(config.PrefDeviceName, &deviceName)
	deviceInfo := &mqtthass.Device{
		Name:         deviceName,
		URL:          "https://github.com/joshuar/go-hass-agent",
		SWVersion:    config.AppVersion,
		Manufacturer: "go-hass-agent",
		Model:        a.AppID(),
		Identifiers:  []string{"go-hass-agent01"},
	}

	sessionPath := dbusx.GetSessionPath(ctx)
	configs := make(map[string]*mqtthass.EntityConfig)
	configs["lock_session"] = mqtthass.NewEntityByID("lock_session", "go_hass_agent").
		AsButton().
		WithDefaultOriginInfo().
		WithDeviceInfo(deviceInfo).
		WithIcon("mdi:eye-lock").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
				Path(sessionPath).
				Destination("org.freedesktop.login1").
				Call("org.freedesktop.login1.Session.Lock")
			if err != nil {
				log.Warn().Err(err).Msg("Could not lock session.")
			}
		})
	configs["unlock_session"] = mqtthass.NewEntityByID("unlock_session", "go_hass_agent").
		AsButton().
		WithDefaultOriginInfo().
		WithDeviceInfo(deviceInfo).
		WithIcon("mdi:eye-lock-open").
		WithCommandCallback(func(_ MQTT.Client, _ MQTT.Message) {
			err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
				Path(sessionPath).
				Destination("org.freedesktop.login1").
				Call("org.freedesktop.login1.Session.UnLock")
			if err != nil {
				log.Warn().Err(err).Msg("Could not unlock session.")
			}
		})
	return &mqttDevice{
		configs: configs,
	}
}
