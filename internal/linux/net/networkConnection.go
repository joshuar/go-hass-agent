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
	"slices"
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
	bus  *dbusx.Bus
	list []dbus.ObjectPath
	mu   sync.Mutex
}

func (w *connectionsWorker) track(path dbus.ObjectPath) {
	w.mu.Lock()
	w.list = append(w.list, path)
	w.mu.Unlock()
}

func (w *connectionsWorker) untrack(path dbus.ObjectPath) {
	w.mu.Lock()
	w.list = slices.DeleteFunc(w.list, func(p dbus.ObjectPath) bool {
		return path == p
	})
	w.mu.Unlock()
}

func (w *connectionsWorker) isTracked(path dbus.ObjectPath) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	return slices.Contains(w.list, path)
}

func (w *connectionsWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return nil, linux.ErrUnimplemented
}

//nolint:cyclop,mnd
//revive:disable:function-length
func (w *connectionsWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)
	logger := slog.With(slog.String("worker", netConnWorkerID))

	connectionlist, err := dbusx.NewProperty[[]dbus.ObjectPath](w.bus, dBusNMPath, dBusNMObj, dBusNMObj+".ActiveConnections").Get()
	if err != nil {
		logger.Debug("Error getting active connections from D-Bus", slog.Any("error", err))
	}

	w.list = connectionlist

	handleConn := func(path dbus.ObjectPath) {
		var conn *connection

		conn, err = newConnection(w.bus, path)
		if err != nil {
			logger.Debug("Unable to monitor connection.", slog.Any("error", err))

			return
		}
		// Ignore loopback.
		if conn.name == "lo" {
			return
		}
		// Start monitoring the connection. Pass any sensor updates from the
		// connection through the sensor channel.
		go func() {
			for s := range conn.monitor(ctx, w.bus) {
				sensorCh <- s
			}

			w.untrack(path)
		}()
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPathNamespace(dbusNMActiveConnPath),
		dbusx.MatchInterface(dbusNMActiveConnIntr),
		dbusx.MatchMembers("StateChanged"),
	).Start(ctx, w.bus)
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("failed to create network connections D-Bus watch: %w", err)
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-triggerCh:
				connectionPath := dbus.ObjectPath(event.Path)
				// If this connection is in the process of deactivating, don't
				// start tracking it.
				if state, stateChange := event.Content[0].(uint32); stateChange {
					if state > 2 {
						continue
					}
				}
				// Track all activating/new connections.
				if !w.isTracked(connectionPath) {
					w.track(connectionPath)
					handleConn(connectionPath)
				}
			}
		}
	}()

	// monitor all current active connections
	for _, path := range w.list {
		handleConn(path)
	}

	return sensorCh, nil
}

func NewConnectionWorker(ctx context.Context, api *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	bus, err := api.GetBus(ctx, dbusx.SystemBus)
	if err != nil {
		return nil, fmt.Errorf("unable to monitor for network connections: %w", err)
	}

	return &linux.SensorWorker{
			Value: &connectionsWorker{
				bus: bus,
			},
			WorkerID: netConnWorkerID,
		},
		nil
}
