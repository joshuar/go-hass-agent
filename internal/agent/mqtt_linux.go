// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	mqtthass "github.com/joshuar/go-hass-anything/v7/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/linux/power"
)

func newMQTTObject(ctx context.Context) *mqttObj {
	var entities []*mqtthass.EntityConfig

	entities = append(entities, power.NewScreenLockControl(ctx))
	entities = append(entities, power.NewPowerControl(ctx)...)

	return &mqttObj{
		entities: entities,
	}
}
