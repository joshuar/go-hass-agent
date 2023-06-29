// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

func (agent *Agent) setupDevice(ctx context.Context) hass.DeviceInfo {
	return linux.NewDevice(ctx, Name, Version)
}
