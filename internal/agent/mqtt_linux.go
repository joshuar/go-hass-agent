// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	mqtthass "github.com/joshuar/go-hass-anything/v7/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v7/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/linux/media"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
)

// newMQTTObject creates an MQTT object for the agent to use for this operating
// system (Linux).
func newMQTTObject(ctx context.Context) *mqttObj {
	var entities []*mqtthass.EntityConfig
	var subscriptions []*mqttapi.Subscription

	msgCh := make(chan mqttapi.Msg)

	// Add screensaver/screenlock control.
	entities = append(entities, power.NewScreenLockControl(ctx))
	// Add power controls (poweroff, reboot, suspend, etc.).
	entities = append(entities, power.NewPowerControl(ctx)...)
	// Add volume control
	entities = append(entities, media.VolumeControl(ctx, msgCh)...)

	// Add subscription for issuing D-Bus commands to the Linux device.
	subscriptions = append(subscriptions, system.NewDBusCommandSubscription(ctx))

	return &mqttObj{
		entities:      entities,
		subscriptions: subscriptions,
		msgCh:         msgCh,
	}
}
