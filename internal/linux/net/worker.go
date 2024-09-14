// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package net

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
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

	netConnWorkerID = "network_connection_sensors"
)

type connectionsWorker struct {
	bus    *dbusx.Bus
	logger *slog.Logger
	list   map[string]*connection
	mu     sync.Mutex
}

func (w *connectionsWorker) track(conn *connection) {
	w.mu.Lock()
	w.list[conn.name] = conn
	w.mu.Unlock()
}

func (w *connectionsWorker) untrack(id string) {
	w.mu.Lock()
	delete(w.list, id)
	w.mu.Unlock()
}

func (w *connectionsWorker) isTracked(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, found := w.list[id]; found {
		return true
	}

	return false
}

func (w *connectionsWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return nil, linux.ErrUnimplemented
}

//nolint:mnd
//revive:disable:function-length
func (w *connectionsWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)
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
		w.logger.Debug("Watching for network connections.")

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
				w.logger.Debug("Could not monitor connection.", slog.Any("error", err))
			}
		}

		w.logger.Debug("Stopped watching network connections.")
	}()

	go func() {
		defer connCancel()
		<-ctx.Done()
		w.logger.Debug("Stopped events.")
	}()

	// monitor all current active connections
	connectionlist, err := dbusx.NewProperty[[]dbus.ObjectPath](w.bus, dBusNMPath, dBusNMObj, dBusNMObj+".ActiveConnections").Get()
	if err != nil {
		w.logger.Debug("Error getting active connections from D-Bus", slog.Any("error", err))
	} else {
		for _, path := range connectionlist {
			if err := w.handleConnection(connCtx, path, sensorCh); err != nil {
				w.logger.Debug("Could not monitor connection.", slog.Any("error", err))
			}
		}
	}

	return sensorCh, nil
}

func (w *connectionsWorker) handleConnection(ctx context.Context, path dbus.ObjectPath, sensorCh chan sensor.Details) error {
	conn, err := newConnection(w.bus, path)
	if err != nil {
		return fmt.Errorf("could not create connection: %w", err)
	}
	// Ignore loopback or already tracked connections.
	if conn.name == "lo" || w.isTracked(conn.name) {
		slog.Debug("Ignoring connection.", slog.String("connection", conn.name))

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

func NewConnectionWorker(ctx context.Context) (*linux.SensorWorker, error) {
	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return nil, linux.ErrNoSystemBus
	}

	return &linux.SensorWorker{
			Value: &connectionsWorker{
				bus:    bus,
				logger: slog.With(slog.String("worker", netConnWorkerID)),
				list:   make(map[string]*connection),
			},
			WorkerID: netConnWorkerID,
		},
		nil
}
