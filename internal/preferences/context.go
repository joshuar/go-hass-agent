// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"context"
)

type key int

var cfgKey key

// EmbedInContext will store the config in the given context.
func EmbedInContext(ctx context.Context, p *Preferences) context.Context {
	return context.WithValue(ctx, cfgKey, *p)
}

// FetchFromContext will attempt to fetch the config from the given context.
func FetchFromContext(ctx context.Context) Preferences {
	c, ok := ctx.Value(cfgKey).(Preferences)
	if !ok {
		return *defaultPreferences()
	}

	return c
}
