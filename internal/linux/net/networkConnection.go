// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package net

import (
	"context"
	"slices"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

//go:generate stringer -type=connState -output networkConnectionStates.go -linecomment
const (
	connUnknown      connState = iota // Unknown
	connActivating                    // Activating
	connOnline                        // Online
	connDeactivating                  // Deactivating
	connOffline                       // Offline

	iconUnknown      // mdi:help-network
	iconActivating   // mdi:plus-network
	iconOnline       // mdi:network
	iconDeactivating // mdi:network-minus
	iconOffline      // mdi:network-off

	dBusNMPath           = "/org/freedesktop/NetworkManager"
	dBusNMObj            = "org.freedesktop.NetworkManager"
	dbusNMActiveConnPath = dBusNMPath + "/ActiveConnection"
	dbusNMActiveConnIntr = dBusNMObj + ".Connection.Active"

	connStateChangedSignal = "StateChanged"
	ipv4ConfigProp         = "Ip4Config"
	ipv6ConfigProp         = "Ip6Config"
	activeConnectionsProp  = "ActivatingConnection"
)

type connState uint32

type connection struct {
	attrs *connectionAttributes
	name  string
	path  dbus.ObjectPath
	linux.Sensor
	state connState
	mu    sync.Mutex
}

type connectionAttributes struct {
	ConnectionType string `json:"connection_type,omitempty"`
	Ipv4           string `json:"ipv4_address,omitempty"`
	Ipv6           string `json:"ipv6_address,omitempty"`
	DataSource     string `json:"data_source"`
	IPv4Mask       int    `json:"ipv4_mask,omitempty"`
	IPv6Mask       int    `json:"ipv6_mask,omitempty"`
}

func (c *connection) Name() string {
	return c.name + " Connection State"
}

func (c *connection) ID() string {
	return strcase.ToSnake(c.name) + "_connection_state"
}

//nolint:mnd
func (c *connection) Icon() string {
	c.mu.Lock()
	i := c.state + 5
	c.mu.Unlock()

	return i.String()
}

func (c *connection) Attributes() any {
	return c.attrs
}

func (c *connection) State() any {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.state.String()
}

func (c *connection) updateState(state connState) {
	// Update connection state (if it changed).
	if c.state != state {
		c.state = state
	}
}

//nolint:mnd
func (c *connection) updateAddrs(addr address) {
	switch addr.class {
	case 4:
		if c.attrs.Ipv4 != addr.address {
			c.attrs.Ipv4 = addr.address
			c.attrs.IPv4Mask = addr.mask
		}
	case 6:
		if c.attrs.Ipv6 != addr.address {
			c.attrs.Ipv6 = addr.address
			c.attrs.IPv6Mask = addr.mask
		}
	}
}

//nolint:exhaustruct,mnd
func newConnection(ctx context.Context, path dbus.ObjectPath) *connection {
	newConnection := &connection{
		path: path,
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorConnectionState,
			IsDiagnostic:    true,
		},
		attrs: &connectionAttributes{
			DataSource: linux.DataSrcDbus,
		},
	}

	// fetch properties for the connection
	var err error

	newConnection.name, err = dbusx.GetProp[string](ctx, dbusx.SystemBus, string(path), dBusNMObj, dbusNMActiveConnIntr+".Id")
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve connection ID.")
	}

	newConnection.state, err = dbusx.GetProp[connState](ctx, dbusx.SystemBus, string(path), dBusNMObj, dbusNMActiveConnIntr+".State")
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve connection state.")
	}

	newConnection.attrs.ConnectionType, err = dbusx.GetProp[string](ctx,
		dbusx.SystemBus, string(path), dBusNMObj, dbusNMActiveConnIntr+".Type")
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve connection type.")
	}

	ip4ConfigPath, err := dbusx.GetProp[dbus.ObjectPath](ctx, dbusx.SystemBus, string(path), dBusNMObj, dbusNMActiveConnIntr+".Ip4Config")
	if err != nil {
		log.Warn().Err(err).Str("connection", newConnection.name).Msg("Could not fetch IPv4 address.")
	}

	newConnection.attrs.Ipv4, newConnection.attrs.IPv4Mask = getAddr(ctx, 4, ip4ConfigPath)

	ip6ConfigPath, err := dbusx.GetProp[dbus.ObjectPath](ctx, dbusx.SystemBus, string(path), dBusNMObj, dbusNMActiveConnIntr+".Ip6Config")
	if err != nil {
		log.Warn().Err(err).Str("connection", newConnection.name).Msg("Could not fetch IPv4 address.")
	}

	newConnection.attrs.Ipv6, newConnection.attrs.IPv6Mask = getAddr(ctx, 6, ip6ConfigPath)

	return newConnection
}

//nolint:cyclop
//revive:disable:unnecessary-stmt
func monitorConnection(ctx context.Context, connPath dbus.ObjectPath) <-chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	updateCh := make(chan any)

	// create a new connection sensor
	conn := newConnection(ctx, connPath)

	// process updates and handle cancellation
	connCtx, connCancel := context.WithCancel(ctx)

	go func() {
		defer close(sensorCh)
		defer close(updateCh)

		for {
			select {
			case <-connCtx.Done():
				log.Debug().Str("connection", conn.Name()).Str("path", string(conn.path)).
					Msg("Connection deactivated.")

				return
			case <-ctx.Done():
				log.Debug().Str("connection", conn.Name()).Str("path", string(conn.path)).
					Msg("Stopped monitoring connection.")

				return
			case u := <-updateCh:
				conn.mu.Lock()
				switch update := u.(type) {
				case connState:
					conn.updateState(update)
				case address:
					conn.updateAddrs(update)
				}
				sensorCh <- conn
				conn.mu.Unlock()
			}
		}
	}()

	// send the initial connection state as a sensor
	go func() {
		sensorCh <- conn
	}()

	// monitor state changes
	go func() {
		defer connCancel()

		for state := range monitorConnectionState(connCtx, string(conn.path)) {
			updateCh <- state
		}
	}()

	// monitor address changes
	go func() {
		for addr := range monitorAddresses(connCtx, string(conn.path)) {
			updateCh <- addr
		}
	}()

	// monitor for additional states depending on the type of connection
	switch conn.attrs.ConnectionType {
	case "802-11-wireless":
		go func() {
			for s := range monitorWifi(connCtx, conn.path) {
				sensorCh <- s
			}
		}()
	}

	log.Debug().Str("connection", conn.Name()).Msg("Monitoring connection.")

	return sensorCh
}

//nolint:exhaustruct,mnd
func monitorConnectionState(ctx context.Context, path string) chan connState {
	stateCh := make(chan connState)

	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{"State"},
		Interface: dbusNMActiveConnIntr,
		Path:      path,
	})
	if err != nil {
		log.Debug().Err(err).Str("path", path).
			Msg("Failed to create connection state D-Bus watch.")
		close(stateCh)

		return stateCh
	}

	log.Debug().Str("path", path).Msg("Monitoring connection state.")

	go func() {
		defer close(stateCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Str("path", path).Msg("Unmonitoring connection state.")

				return
			case event := <-events:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					continue
				}

				stateProp, stateChanged := props.Changed["State"]
				if !stateChanged {
					continue
				}

				currentState := dbusx.VariantToValue[connState](stateProp)
				stateCh <- currentState

				if currentState == 4 {
					log.Debug().Str("path", path).Msg("Unmonitoring connection state.")

					return
				}
			}
		}
	}()

	return stateCh
}

//nolint:exhaustruct,mnd
func monitorAddresses(ctx context.Context, path string) chan address {
	sensorCh := make(chan address)

	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SystemBus,
		Names:     []string{ipv4ConfigProp, ipv6ConfigProp},
		Interface: dbusNMActiveConnIntr,
		Path:      path,
	})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create address changes D-Bus watch.")
		close(sensorCh)

		return sensorCh
	}

	log.Debug().Str("path", path).Msg("Monitoring address changes.")

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Str("path", path).Msg("Unmonitoring address changes.")

				return
			case event := <-events:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					continue
				}

				if ipv4Change, ipv4Changed := props.Changed[ipv4ConfigProp]; ipv4Changed {
					addr, mask := getAddr(ctx, 4, dbusx.VariantToValue[dbus.ObjectPath](ipv4Change))
					sensorCh <- address{address: addr, mask: mask, class: 4}
				}

				if ipv6Change, ipv6Changed := props.Changed[ipv6ConfigProp]; ipv6Changed {
					addr, mask := getAddr(ctx, 4, dbusx.VariantToValue[dbus.ObjectPath](ipv6Change))
					sensorCh <- address{address: addr, mask: mask, class: 6}
				}
			}
		}
	}()

	return sensorCh
}

//nolint:mnd
func getAddr(ctx context.Context, ver int, path dbus.ObjectPath) (addr string, mask int) {
	if path == "/" {
		return "", 0
	}

	var connProp string

	switch ver {
	case 4:
		connProp = dBusNMObj + ".IP4Config"
	case 6:
		connProp = dBusNMObj + ".IP6Config"
	}

	addrDetails, err := dbusx.GetProp[[]map[string]dbus.Variant](ctx, dbusx.SystemBus, string(path), dBusNMObj, connProp+".AddressData")
	if err != nil {
		return "", 0
	}

	var address string

	var prefix int

	if len(addrDetails) > 0 {
		address = dbusx.VariantToValue[string](addrDetails[0]["address"])
		prefix = dbusx.VariantToValue[int](addrDetails[0]["prefix"])
		log.Debug().Str("path", string(path)).Str("address", address).Int("prefix", prefix).
			Msg("Retrieved address.")
	}

	return address, prefix
}

func getActiveConnections(ctx context.Context) []dbus.ObjectPath {
	connectionPaths, err := dbusx.GetProp[[]dbus.ObjectPath](ctx, dbusx.SystemBus, dBusNMPath, dBusNMObj, dBusNMObj+".ActiveConnections")
	if err != nil {
		log.Debug().Err(err).
			Msg("Could not retrieve active connection list.")

		return nil
	}

	return connectionPaths
}

type address struct {
	address string
	class   int
	mask    int
}

type connectionsWorker struct {
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

//nolint:exhaustruct,mnd
func (w *connectionsWorker) Events(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	w.list = getActiveConnections(ctx)
	handleConn := func(path dbus.ObjectPath) {
		go func() {
			for s := range monitorConnection(ctx, path) {
				sensorCh <- s
			}

			w.untrack(path)
		}()
	}

	events, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:           dbusx.SystemBus,
		Names:         []string{"StateChanged"},
		PathNamespace: dbusNMActiveConnPath,
		// Path:      dBusNMPath,
		Interface: dbusNMActiveConnIntr,
	})
	if err != nil {
		log.Debug().Err(err).
			Msg("Failed to create network connections D-Bus watch.")
		close(sensorCh)

		return sensorCh
	}

	go func() {
		log.Debug().Msg("Monitoring for network connection changes.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped network connection monitoring.")

				return
			case event := <-events:
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

	return sensorCh
}

//nolint:exhaustruct
func NewConnectionWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Network Connection Sensors",
			WorkerDesc: "Sensors to track network connection states and other connection specific properties.",
			Value:      &connectionsWorker{},
		},
		nil
}
