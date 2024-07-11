// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"context"
	"log/slog"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/logging"
)

type DBusAPI struct {
	dbus map[dbusType]*Bus
	mu   sync.Mutex
}

func NewDBusAPI(ctx context.Context, logger *slog.Logger) *DBusAPI {
	api := &DBusAPI{
		dbus: make(map[dbusType]*Bus),
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

func (a *DBusAPI) GetBus(ctx context.Context, busType dbusType) (*Bus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var (
		bus    *Bus
		exists bool
	)

	if bus, exists = a.dbus[busType]; !exists {
		return newBus(ctx, busType, logging.FromContext(ctx))
	}

	return bus, nil
}
