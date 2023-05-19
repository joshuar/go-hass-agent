// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"errors"
)

type API interface {
	SensorWorkers() []func(context.Context, chan interface{})
	EndPoint(string) interface{}
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// configKey is the key for API values in Contexts. It is unexported; clients
// use device.StoreAPIInContext and device.FetchAPIFromContext instead of using
// this key directly.
var configKey key

// StoreAPIInContext returns a new Context that embeds an API.
func StoreAPIInContext(ctx context.Context, a API) context.Context {
	return context.WithValue(ctx, configKey, a)
}

// FetchAPIFromContext returns the API stored in ctx, or an error if there is
// none
func FetchAPIFromContext(ctx context.Context) (API, error) {
	if c, ok := ctx.Value(configKey).(API); !ok {
		return nil, errors.New("no API in context")
	} else {
		return c, nil
	}
}

func GetAPIEndpoint[T any](api API, endpoint string) T {
	if e := api.EndPoint(endpoint); e != nil {
		return e.(T)
	} else {
		return *new(T)
	}
}
