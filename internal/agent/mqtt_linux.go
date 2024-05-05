// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/linux/media"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
)

type linuxMQTTDevice struct {
	msgs     chan *mqttapi.Msg
	sensors  []*mqtthass.SensorEntity
	buttons  []*mqtthass.ButtonEntity
	numbers  []*mqtthass.NumberEntity[int]
	switches []*mqtthass.SwitchEntity
	controls []*mqttapi.Subscription
}

func (d *linuxMQTTDevice) Subscriptions() []*mqttapi.Subscription {
	var subs []*mqttapi.Subscription

	for _, button := range d.buttons {
		if sub, err := button.MarshalSubscription(); err != nil {
			log.Warn().Err(err).Str("entity", button.Name).Msg("Could not create subscription.")
		} else {
			subs = append(subs, sub)
		}
	}
	for _, number := range d.numbers {
		if sub, err := number.MarshalSubscription(); err != nil {
			log.Warn().Err(err).Str("entity", number.Name).Msg("Could not create subscription.")
		} else {
			subs = append(subs, sub)
		}
	}
	for _, sw := range d.switches {
		if sub, err := sw.MarshalSubscription(); err != nil {
			log.Warn().Err(err).Str("entity", sw.Name).Msg("Could not create subscription.")
		} else {
			subs = append(subs, sub)
		}
	}

	subs = append(subs, d.controls...)

	return subs
}

func (d *linuxMQTTDevice) Configs() []*mqttapi.Msg {
	var configs []*mqttapi.Msg

	for _, sensor := range d.sensors {
		if sub, err := sensor.MarshalConfig(); err != nil {
			log.Warn().Err(err).Str("entity", sensor.Name).Msg("Could not create subscription.")
		} else {
			configs = append(configs, sub)
		}
	}
	for _, button := range d.buttons {
		if sub, err := button.MarshalConfig(); err != nil {
			log.Warn().Err(err).Str("entity", button.Name).Msg("Could not create subscription.")
		} else {
			configs = append(configs, sub)
		}
	}
	for _, number := range d.numbers {
		if sub, err := number.MarshalConfig(); err != nil {
			log.Warn().Err(err).Str("entity", number.Name).Msg("Could not create subscription.")
		} else {
			configs = append(configs, sub)
		}
	}
	for _, sw := range d.switches {
		if sub, err := sw.MarshalConfig(); err != nil {
			log.Warn().Err(err).Str("entity", sw.Name).Msg("Could not create subscription.")
		} else {
			configs = append(configs, sub)
		}
	}

	return configs
}

func (d *linuxMQTTDevice) Msgs() chan *mqttapi.Msg {
	return d.msgs
}

func (d *linuxMQTTDevice) Setup(_ context.Context) error {
	return nil
}

func newMQTTDevice(ctx context.Context) *linuxMQTTDevice {
	dev := &linuxMQTTDevice{
		msgs: make(chan *mqttapi.Msg),
	}

	// Add the power controls (suspend, resume, poweroff, etc.).
	dev.buttons = append(dev.buttons, power.NewPowerControl(ctx)...)
	// Add the screen lock controls.
	dev.buttons = append(dev.buttons, power.NewScreenLockControl(ctx))
	// Add the volume controls.
	volEntity, muteEntity := media.VolumeControl(ctx, dev.Msgs())
	dev.numbers = append(dev.numbers, volEntity)
	dev.switches = append(dev.switches, muteEntity)
	// Add the D-Bus command action.
	dev.controls = append(dev.controls, system.NewDBusCommandSubscription(ctx))

	return dev
}
