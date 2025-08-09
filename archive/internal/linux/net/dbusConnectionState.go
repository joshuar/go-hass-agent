// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package net

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	dBusNMPath           = "/org/freedesktop/NetworkManager"
	dBusNMObj            = "org.freedesktop.NetworkManager"
	dbusNMActiveConnPath = dBusNMPath + "/ActiveConnection"
	dbusNMActiveConnIntr = dBusNMObj + ".Connection.Active"

	connStateChangedSignal = "StateChanged"
	ipv4ConfigPropName     = "Ip4Config"
	ipv6ConfigPropName     = "Ip6Config"
	statePropName          = "State"
	activeConnectionsProp  = "ActivatingConnection"

	netConnWorkerID   = "network_connection_sensors"
	netConnWorkerDesc = "NetworkManager connection status"
	netConnPrefID     = prefPrefix + "connections"
)

var _ workers.EntityWorker = (*ConnectionsWorker)(nil)

var ErrInitConnStateWorker = errors.New("could not init network connection state worker")

type ConnectionsWorker struct {
	bus   *dbusx.Bus
	list  map[string]*connection
	prefs *WorkerPrefs
	mu    sync.Mutex
	*models.WorkerMetadata
}

func (w *ConnectionsWorker) track(conn *connection) {
	w.mu.Lock()
	w.list[conn.name] = conn
	w.mu.Unlock()
}

func (w *ConnectionsWorker) untrack(id string) {
	w.mu.Lock()
	delete(w.list, id)
	w.mu.Unlock()
}

func (w *ConnectionsWorker) isTracked(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, found := w.list[id]; found {
		return true
	}

	return false
}

//nolint:mnd
//revive:disable:function-length
func (w *ConnectionsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)
	connCtx, connCancel := context.WithCancel(ctx)

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPathNamespace(dbusNMActiveConnPath),
		dbusx.MatchInterface(dbusNMActiveConnIntr),
		dbusx.MatchMembers("StateChanged"),
	).Start(connCtx, w.bus)
	if err != nil {
		close(sensorCh)
		connCancel()

		return sensorCh, fmt.Errorf("failed to create network connections D-Bus watch: %w", err)
	}

	go func() {
		defer close(sensorCh)
		slogctx.FromCtx(ctx).Debug("Watching for network connections.")

		for event := range triggerCh {
			connectionPath := dbus.ObjectPath(event.Path)
			// If this connection is in the process of deactivating, don't
			// start tracking it.
			if state, stateChange := event.Content[0].(uint32); stateChange {
				if state > 2 {
					continue
				}
			}
			// Track all activating/new connections.
			if err = w.handleConnection(connCtx, connectionPath, sensorCh); err != nil {
				slogctx.FromCtx(ctx).Debug("Could not monitor connection.", slog.Any("error", err))
			}
		}

		slogctx.FromCtx(ctx).Debug("Stopped watching network connections.")
	}()

	go func() {
		defer connCancel()
		<-ctx.Done()
		slogctx.FromCtx(ctx).Debug("Stopped events.")
	}()

	// monitor all current active connections
	connectionlist, err := dbusx.NewProperty[[]dbus.ObjectPath](w.bus, dBusNMPath, dBusNMObj, dBusNMObj+".ActiveConnections").Get()
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Error getting active connections from D-Bus", slog.Any("error", err))
	} else {
		for _, path := range connectionlist {
			if err := w.handleConnection(connCtx, path, sensorCh); err != nil {
				slogctx.FromCtx(ctx).Debug("Could not monitor connection.", slog.Any("error", err))
			}
		}
	}

	return sensorCh, nil
}

func (w *ConnectionsWorker) PreferencesID() string {
	return netConnPrefID
}

func (w *ConnectionsWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{
		IgnoredDevices: defaultIgnoredDevices,
	}
}

func (w *ConnectionsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *ConnectionsWorker) handleConnection(ctx context.Context, path dbus.ObjectPath, sensorCh chan models.Entity) error {
	conn, err := newConnection(w.bus, path)
	if err != nil {
		return fmt.Errorf("could not create connection: %w", err)
	}
	// Ignore loopback or already tracked connections.
	if conn.name == "lo" || w.isTracked(conn.name) {
		slog.Debug("Ignoring connection.", slog.String("connection", conn.name))

		return nil
	}
	// Ignore user-defined devices.
	if slices.ContainsFunc(w.prefs.IgnoredDevices, func(e string) bool {
		return strings.HasPrefix(conn.name, e)
	}) {
		return nil
	}

	// Start monitoring the connection. Pass any sensor updates from the
	// connection through the sensor channel.
	go func() {
		w.track(conn)

		for s := range conn.monitor(ctx, w.bus) {
			sensorCh <- s
		}

		w.untrack(conn.name)
	}()

	return nil
}

func NewConnectionWorker(ctx context.Context) (workers.EntityWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitConnStateWorker, linux.ErrNoSystemBus)
	}

	worker := &ConnectionsWorker{
		WorkerMetadata: models.SetWorkerMetadata(netConnWorkerID, netConnWorkerDesc),
		bus:            bus,
		list:           make(map[string]*connection),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return worker, errors.Join(ErrInitConnStateWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
