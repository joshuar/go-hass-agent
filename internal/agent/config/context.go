// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"context"
)

type key int

var cfgKey key

func EmbedInContext(ctx context.Context, c Config) context.Context {
	return context.WithValue(ctx, cfgKey, c)
}

func FetchFromContext(ctx context.Context) Config {
	c, ok := ctx.Value(cfgKey).(Config)
	if !ok {
		return nil
	}
	return c
}
