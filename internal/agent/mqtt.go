// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	mqttapi "github.com/joshuar/go-hass-anything/v9/pkg/mqtt"
)

type mqttDevice interface {
	Subscriptions() []*mqttapi.Subscription
	Configs() []*mqttapi.Msg
	Msgs() chan *mqttapi.Msg
	Setup(ctx context.Context) error
}
