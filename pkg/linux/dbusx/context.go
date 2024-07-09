// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"context"
	"log/slog"
	"sync"
)

type dBusAPI struct {
	dbus map[dbusType]*bus
	mu   sync.Mutex
}

func newDBusAPI(ctx context.Context, logger *slog.Logger) *dBusAPI {
	api := &dBusAPI{
		dbus: make(map[dbusType]*bus),
		mu:   sync.Mutex{},
	}

	api.mu.Lock()
	for _, b := range []dbusType{SessionBus, SystemBus} {
		bus, err := newBus(ctx, b, logger)
		if err != nil {
			slog.Warn("Could not connect to D-Bus.", "bus", b.String(), "error", err.Error())
		} else {
			api.dbus[b] = bus
		}
	}
	api.mu.Unlock()

	return api
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// linuxCtxKey is the key for dbusAPI values in Contexts. It is unexported;
// clients use Setup and getBus instead of using this key directly.
var linuxCtxKey key

// Setup returns a new Context that contains the D-Bus API.
func Setup(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, linuxCtxKey, newDBusAPI(ctx, logger))
}

// getBus retrieves the D-Bus API object from the context.
//
//revive:disable:indent-error-flow
func getBus(ctx context.Context, busType dbusType) (*bus, bool) {
	bus, ok := ctx.Value(linuxCtxKey).(*dBusAPI)
	if !ok {
		return nil, false
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus, busExists := bus.dbus[busType]; !busExists {
		return nil, false
	} else {
		return bus, true
	}
}
