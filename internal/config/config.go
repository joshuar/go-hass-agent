// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"context"
	"errors"
)

type Config interface {
	Get(string) (interface{}, error)
	Set(string, interface{}) error
	Validate() error
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// configKey is the key for agent.AppConfig values in Contexts. It is
// unexported; clients use config.NewContext and config.FromContext
// instead of using this key directly.
var configKey key

// StoreInContext returns a new Context that stores the Config, c.
func StoreInContext(ctx context.Context, c Config) context.Context {
	return context.WithValue(ctx, configKey, c)
}

// FetchFromContext returns the Config value stored in ctx, if any.
func FetchFromContext(ctx context.Context) (Config, error) {
	if c, ok := ctx.Value(configKey).(Config); !ok {
		return nil, errors.New("no API in context")
	} else {
		return c, nil
	}
}

// FetchPropertyFromContext is a helper function to retrieve a specific config
// property from the Config store in ctx, rather than retrieving the entire
// Config object.
func FetchPropertyFromContext(ctx context.Context, property string) (interface{}, error) {
	config, err := FetchFromContext(ctx)
	if err != nil {
		return nil, err
	} else {
		value, err := config.Get(property)
		if err != nil {
			return nil, err
		} else {
			return value, nil
		}
	}
}
