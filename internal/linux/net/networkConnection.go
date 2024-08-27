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
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
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

	netConnWorkerID = "network_connection_sensors"
)

type connState uint32

type connection struct {
	attrs  *connectionAttributes
	logger *slog.Logger
	bus    *dbusx.Bus
	name   string
	path   dbus.ObjectPath
	linux.Sensor
	mu    sync.Mutex
	state connState
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

func (c *connection) Attributes() map[string]any {
	attributes := make(map[string]any)

	attributes["connection_type"] = c.attrs.ConnectionType
	attributes["data_source"] = linux.DataSrcDbus

	if c.attrs.Ipv4 != "" {
		attributes["ipv4_address"] = c.attrs.Ipv4
		attributes["ipv4_mask"] = c.attrs.IPv4Mask
	}

	if c.attrs.Ipv6 != "" {
		attributes["ipv6_address"] = c.attrs.Ipv6
		attributes["ipv6_mask"] = c.attrs.IPv6Mask
	}

	return attributes
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

//nolint:mnd
func (c *connection) monitorState(ctx context.Context) chan connState {
	stateCh := make(chan connState)

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(string(c.path)),
		dbusx.MatchInterface(dbusNMActiveConnIntr),
		dbusx.MatchMembers("State"),
	).Start(ctx, c.bus)
	if err != nil {
		c.logger.Debug("Could not create D-Bus watch for connection state.", "error", err.Error())
		close(stateCh)

		return stateCh
	}

	go func() {
		c.logger.Debug("Monitoring connection state.")

		defer close(stateCh)

		for {
			select {
			case <-ctx.Done():
				c.logger.Debug("Unmonitoring connection state.")

				return
			case event := <-triggerCh:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					continue
				}

				stateProp, stateChanged := props.Changed["State"]
				if !stateChanged {
					continue
				}

				currentState, err := dbusx.VariantToValue[connState](stateProp)
				if err != nil {
					continue
				}
				stateCh <- currentState

				if currentState == 4 {
					c.logger.Debug("Unmonitoring connection state.")

					return
				}
			}
		}
	}()

	return stateCh
}

//nolint:mnd,nestif,gocognit,cyclop
func (c *connection) monitorAddresses(ctx context.Context) chan address {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(string(c.path)),
		dbusx.MatchInterface(dbusNMActiveConnIntr),
		dbusx.MatchMembers(ipv4ConfigProp, ipv6ConfigProp),
	).Start(ctx, c.bus)
	if err != nil {
		c.logger.Warn("Unable to set-up D-Bus watch for address changes.", slog.Any("error", err))

		return nil
	}

	sensorCh := make(chan address)

	// events, err := dbusx.NewWatch(
	// 	dbusx.MatchPath(string(c.path)),
	// 	dbusx.MatchInterface(dbusNMActiveConnIntr),
	// 	dbusx.MatchMember(ipv4ConfigProp, ipv6ConfigProp),
	// ).Start(ctx, c.bus)
	// if err != nil {
	// 	c.logger.Debug("Failed to watch D-Bus for connection address changes.", "error", err.Error())
	// 	close(sensorCh)

	// 	return sensorCh
	// }

	go func() {
		c.logger.Debug("Monitoring connection address changes.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				c.logger.Debug("Unmonitoring connection address changes.")

				return
			case event := <-triggerCh:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					continue
				}

				if ipv4Change, ipv4Changed := props.Changed[ipv4ConfigProp]; ipv4Changed {
					value, err := dbusx.VariantToValue[dbus.ObjectPath](ipv4Change)
					if err != nil {
						c.logger.Warn("Could not retrieve IPv4 address.", "error", err.Error())
					} else {
						addr, mask, err := c.getAddr(4, value)
						if err != nil {
							c.logger.Warn("Could not retrieve IPv4 address.", "error", err.Error())
						} else {
							sensorCh <- address{address: addr, mask: mask, class: 4}
						}
					}
				}

				if ipv6Change, ipv6Changed := props.Changed[ipv6ConfigProp]; ipv6Changed {
					value, err := dbusx.VariantToValue[dbus.ObjectPath](ipv6Change)
					if err != nil {
						c.logger.Warn("Could not retrieve IPv4 address.", "error", err.Error())
					} else {
						addr, mask, err := c.getAddr(6, value)
						if err != nil {
							c.logger.Warn("Could not retrieve IPv4 address.", "error", err.Error())
						} else {
							sensorCh <- address{address: addr, mask: mask, class: 6}
						}
					}
				}
			}
		}
	}()

	return sensorCh
}

//nolint:mnd
func (c *connection) getAddr(ver int, path dbus.ObjectPath) (addr string, mask int, err error) {
	if path == "/" {
		return "", 0, dbusx.ErrInvalidPath
	}

	var connProp string

	switch ver {
	case 4:
		connProp = dBusNMObj + ".IP4Config"
	case 6:
		connProp = dBusNMObj + ".IP6Config"
	}

	addrDetails, err := dbusx.NewProperty[[]map[string]dbus.Variant](c.bus, string(path), dBusNMObj, connProp+".AddressData").Get()
	if err != nil {
		return "", 0, fmt.Errorf("could not retrieve address data from D-Bus: %w", err)
	}

	var (
		address string
		prefix  int
	)

	if len(addrDetails) > 0 {
		address, err = dbusx.VariantToValue[string](addrDetails[0]["address"])
		if err != nil {
			return "", 0, fmt.Errorf("could not parse address: %w", err)
		}

		prefix, err = dbusx.VariantToValue[int](addrDetails[0]["prefix"])
		if err != nil {
			return "", 0, fmt.Errorf("could not parse prefix: %w", err)
		}
	}

	return address, prefix, nil
}

type address struct {
	address string
	class   int
	mask    int
}

type connectionsWorker struct {
	logger            *slog.Logger
	bus               *dbusx.Bus
	activeConnections *dbusx.Property[[]dbus.ObjectPath]
	list              []dbus.ObjectPath
	mu                sync.Mutex
}

//nolint:cyclop
//revive:disable:unnecessary-stmt
func (w *connectionsWorker) monitorConnection(ctx context.Context, connPath dbus.ObjectPath) <-chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	updateCh := make(chan any)

	// create a new connection sensor
	conn := w.newConnection(connPath)

	// process updates and handle cancellation
	connCtx, connCancel := context.WithCancel(ctx)

	go func() {
		defer close(sensorCh)
		defer close(updateCh)

		for {
			select {
			case <-connCtx.Done():
				return
			case <-ctx.Done():
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

		for state := range conn.monitorState(connCtx) {
			updateCh <- state
		}
	}()

	// monitor address changes
	go func() {
		for addr := range conn.monitorAddresses(connCtx) {
			updateCh <- addr
		}
	}()

	// monitor for additional states depending on the type of connection
	switch conn.attrs.ConnectionType {
	case "802-11-wireless":
		go func() {
			for s := range conn.monitorWifi(connCtx) {
				sensorCh <- s
			}
		}()
	}

	return sensorCh
}

//nolint:mnd
func (w *connectionsWorker) newConnection(path dbus.ObjectPath) *connection {
	newConnection := &connection{
		path: path,
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorConnectionState,
			IsDiagnostic:    true,
		},
		attrs: &connectionAttributes{
			DataSource: linux.DataSrcDbus,
		},
		bus: w.bus,
	}

	// fetch properties for the connection
	var err error

	newConnection.name, err = dbusx.NewProperty[string](w.bus, string(path), dBusNMObj, dbusNMActiveConnIntr+".Id").Get()
	if err != nil {
		w.logger.Warn("Could not retrieve connection name.", "error", err.Error())
	}

	newConnection.state, err = dbusx.NewProperty[connState](w.bus, string(path), dBusNMObj, dbusNMActiveConnIntr+".State").Get()
	if err != nil {
		w.logger.Warn("Could not retrieve connection state.", "error", err.Error())
	}

	newConnection.attrs.ConnectionType, err = dbusx.NewProperty[string](
		w.bus, string(path), dBusNMObj, dbusNMActiveConnIntr+".Type").Get()
	if err != nil {
		w.logger.Warn("Could not retrieve connection type.", "error", err.Error())
	}

	ip4ConfigPath, err := dbusx.NewProperty[dbus.ObjectPath](w.bus, string(path), dBusNMObj, dbusNMActiveConnIntr+".Ip4Config").Get()
	if err != nil {
		w.logger.Warn("Could not retrieve IPv4 address for connection.", "error", err.Error())
	}

	newConnection.attrs.Ipv4, newConnection.attrs.IPv4Mask, err = newConnection.getAddr(4, ip4ConfigPath)
	if err != nil {
		w.logger.Warn("Could not retrieve IPv4 address for connection.", "error", err.Error())
	}

	ip6ConfigPath, err := dbusx.NewProperty[dbus.ObjectPath](w.bus, string(path), dBusNMObj, dbusNMActiveConnIntr+".Ip6Config").Get()
	if err != nil {
		w.logger.Warn("Could not retrieve IPv6 address for connection.", "error", err.Error())
	}

	newConnection.attrs.Ipv6, newConnection.attrs.IPv6Mask, err = newConnection.getAddr(6, ip6ConfigPath)
	if err != nil {
		w.logger.Warn("Could not retrieve IPv6 address for connection.", "error", err.Error())
	}

	newConnection.logger = w.logger.
		With(slog.Group("connection_info"),
			slog.String("name", newConnection.name),
			slog.String("connection_type", newConnection.attrs.ConnectionType),
			slog.String("dbus_path", string(newConnection.path)),
		)

	return newConnection
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
func (w *connectionsWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	connectionlist, err := w.activeConnections.Get()
	if err != nil {
		w.logger.Warn("Failed to get any active connections", "error", err.Error())
	}

	w.list = connectionlist

	handleConn := func(path dbus.ObjectPath) {
		go func() {
			for s := range w.monitorConnection(ctx, path) {
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
		w.logger.Debug("Monitoring for network connection changes.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				w.logger.Debug("Stopped monitoring for network connection changes.")

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
				logger:            logging.FromContext(ctx).With(slog.String("worker", netConnWorkerID)),
				bus:               bus,
				activeConnections: dbusx.NewProperty[[]dbus.ObjectPath](bus, dBusNMPath, dBusNMObj, dBusNMObj+".ActiveConnections"),
			},
			WorkerID: netConnWorkerID,
		},
		nil
}
