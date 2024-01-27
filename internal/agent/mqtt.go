// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	mqtthass "github.com/joshuar/go-hass-anything/v3/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v3/pkg/mqtt"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
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
