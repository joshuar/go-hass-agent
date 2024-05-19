// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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
	ConnectionType string `json:"Connection Type,omitempty"`
	Ipv4           string `json:"IPv4 Address,omitempty"`
	Ipv6           string `json:"IPv6 Address,omitempty"`
	DataSource     string `json:"Data Source"`
	IPv4Mask       int    `json:"IPv4 Mask,omitempty"`
	IPv6Mask       int    `json:"IPv6 Mask,omitempty"`
}

func (c *connection) Name() string {
	return c.name + " Connection State"
}

func (c *connection) ID() string {
	return strcase.ToSnake(c.name) + "_connection_state"
}

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

func newConnection(ctx context.Context, path dbus.ObjectPath) *connection {
	c := &connection{
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
	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(path).
		Destination(dBusNMObj)
	var err error
	c.name, err = dbusx.GetProp[string](req, dbusNMActiveConnIntr+".Id")
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve connection ID.")
	}
	c.state, err = dbusx.GetProp[connState](req, dbusNMActiveConnIntr+".State")
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve connection state.")
	}
	c.attrs.ConnectionType, err = dbusx.GetProp[string](req, dbusNMActiveConnIntr+".Type")
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve connection type.")
	}
	ip4ConfigPath, err := dbusx.GetProp[dbus.ObjectPath](req, dbusNMActiveConnIntr+".Ip4Config")
	if err != nil {
		log.Warn().Err(err).Str("connection", c.name).Msg("Could not fetch IPv4 address.")
	}
	c.attrs.Ipv4, c.attrs.IPv4Mask = getAddr(ctx, 4, ip4ConfigPath)
	ip6ConfigPath, err := dbusx.GetProp[dbus.ObjectPath](req, dbusNMActiveConnIntr+".Ip6Config")
	if err != nil {
		log.Warn().Err(err).Str("connection", c.name).Msg("Could not fetch IPv4 address.")
	}
	c.attrs.Ipv6, c.attrs.IPv6Mask = getAddr(ctx, 6, ip6ConfigPath)
	return c
}

func monitorConnection(ctx context.Context, p dbus.ObjectPath) <-chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	updateCh := make(chan any)

	// create a new connection sensor
	c := newConnection(ctx, p)

	// process updates and handle cancellation
	connCtx, connCancel := context.WithCancel(ctx)
	go func() {
		defer close(sensorCh)
		defer close(updateCh)
		for {
			select {
			case <-connCtx.Done():
				log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
					Msg("Connection deactivated.")
				return
			case <-ctx.Done():
				log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
					Msg("Stopped monitoring connection.")
				return
			case u := <-updateCh:
				c.mu.Lock()
				switch update := u.(type) {
				case connState:
					if c.state != update {
						c.state = update
					}
				case address:
					if update.class == 4 {
						if c.attrs.Ipv4 != update.address {
							c.attrs.Ipv4 = update.address
							c.attrs.IPv4Mask = update.mask
						}
					}
					if update.class == 6 {
						if c.attrs.Ipv6 != update.address {
							c.attrs.Ipv6 = update.address
							c.attrs.IPv6Mask = update.mask
						}
					}
				}
				sensorCh <- c
				c.mu.Unlock()
			}
		}
	}()

	// send the initial connection state as a sensor
	go func() {
		sensorCh <- c
	}()

	// monitor state changes
	go func() {
		defer connCancel()
		for state := range monitorConnectionState(connCtx, string(c.path)) {
			updateCh <- state
		}
	}()

	// monitor address changes
	go func() {
		for addr := range monitorAddresses(connCtx, string(c.path)) {
			updateCh <- addr
		}
	}()

	// monitor for additional states depending on the type of connection
	switch c.attrs.ConnectionType {
	case "802-11-wireless":
		go func() {
			for s := range monitorWifi(connCtx, c.path) {
				sensorCh <- s
			}
		}()
	}

	log.Debug().Str("connection", c.Name()).Msg("Monitoring connection.")
	return sensorCh
}

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
	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(path).
		Destination(dBusNMObj)
	addrDetails, err := dbusx.GetProp[[]map[string]dbus.Variant](req, connProp+".AddressData")
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
	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(dBusNMPath).
		Destination(dBusNMObj)
	v, err := dbusx.GetProp[[]dbus.ObjectPath](req, dBusNMObj+".ActiveConnections")
	if err != nil {
		log.Debug().Err(err).
			Msg("Could not retrieve active connection list.")
		return nil
	}
	return v
}

type address struct {
	address string
	class   int
	mask    int
}

type connectionTracker struct {
	list []dbus.ObjectPath
	mu   sync.Mutex
}

func (t *connectionTracker) Track(path dbus.ObjectPath) {
	t.mu.Lock()
	t.list = append(t.list, path)
	t.mu.Unlock()
}

func (t *connectionTracker) Untrack(path dbus.ObjectPath) {
	t.mu.Lock()
	t.list = slices.DeleteFunc(t.list, func(p dbus.ObjectPath) bool {
		return path == p
	})
	t.mu.Unlock()
}

func (t *connectionTracker) Tracked(path dbus.ObjectPath) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return slices.Contains(t.list, path)
}

func ConnectionsUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	tracker := &connectionTracker{
		list: getActiveConnections(ctx),
	}

	handleConn := func(path dbus.ObjectPath) {
		go func() {
			for s := range monitorConnection(ctx, path) {
				sensorCh <- s
			}
			tracker.Untrack(path)
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
				if !tracker.Tracked(connectionPath) {
					tracker.Track(connectionPath)
					handleConn(connectionPath)
				}
			}
		}
	}()

	// monitor all current active connections
	for _, path := range tracker.list {
		handleConn(path)
	}

	return sensorCh
}
