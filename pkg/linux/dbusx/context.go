// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

type dBusAPI struct {
	dbus map[dbusType]*Bus
	mu   sync.Mutex
}

func NewDBusAPI(ctx context.Context) *dBusAPI {
	a := &dBusAPI{}
	a.dbus = make(map[dbusType]*Bus)
	a.mu.Lock()
	for _, b := range []dbusType{SessionBus, SystemBus} {
		bus, err := NewBus(ctx, b)
		if err != nil {
			log.Warn().Err(err).Msg("Could not connect to D-Bus.")
		} else {
			a.dbus[b] = bus
		}
	}
	a.mu.Unlock()
	return a
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// linuxCtxKey is the key for dbusAPI values in Contexts. It is unexported;
// clients use Setup and getBus instead of using this key directly.
var linuxCtxKey key

// Setup returns a new Context that contains the D-Bus API.
func Setup(ctx context.Context) context.Context {
	return context.WithValue(ctx, linuxCtxKey, NewDBusAPI(ctx))
}

// getBus retrieves the D-Bus API object from the context.
func getBus(ctx context.Context, e dbusType) (*Bus, bool) {
	b, ok := ctx.Value(linuxCtxKey).(*dBusAPI)
	if !ok {
		return nil, false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.dbus[e], true
}
