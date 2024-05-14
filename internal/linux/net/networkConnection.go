// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"context"
	"slices"
	"strings"
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
	i := c.state + 5
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

func (c *connection) monitorConnectionState(ctx context.Context) chan sensor.Details {
	log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
		Msg("Monitoring connection state.")
	sensorCh := make(chan sensor.Details)
	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(dbusNMActiveConnPath),
			dbus.WithMatchInterface(dbusNMActiveConnIntr),
			dbus.WithMatchMember("StateChanged"),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Path != c.path || len(s.Body) <= 1 {
				log.Trace().Str("runner", "net").Msg("Not my signal or empty signal body.")
				return
			}
			var props map[string]dbus.Variant
			var ok bool
			if props, ok = s.Body[1].(map[string]dbus.Variant); !ok {
				log.Trace().Str("runner", "net").Msgf("Could not cast signal body, got %T, want %T", s.Body, props)
				return
			}
			if state, ok := props["State"]; ok {
				currentState := dbusx.VariantToValue[connState](state)
				if currentState != c.state {
					c.mu.Lock()
					c.state = currentState
					c.mu.Unlock()
					sensorCh <- c
				}
				if currentState == 4 {
					log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
						Msg("Unmonitoring connection state.")
					close(sensorCh)
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create network connections D-Bus watch.")
		close(sensorCh)
	}
	return sensorCh
}

func (c *connection) monitorAddresses(ctx context.Context) chan sensor.Details {
	log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
		Msg("Monitoring address changes.")
	sensorCh := make(chan sensor.Details)
	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(dbusNMActiveConnPath),
			dbus.WithMatchInterface(dbusNMActiveConnIntr),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Path != c.path || len(s.Body) <= 1 {
				log.Trace().Str("runner", "net").Msg("Not my signal or empty signal body.")
				return
			}
			var props map[string]dbus.Variant
			var ok bool
			if props, ok = s.Body[1].(map[string]dbus.Variant); !ok {
				log.Trace().Str("runner", "net").Msgf("Could not cast signal body, got %T, want %T", s.Body, props)
				return
			}
			go func() {
				for k, v := range props {
					switch k {
					case "Ip4Config":
						addr, mask := getAddr(ctx, 4, dbusx.VariantToValue[dbus.ObjectPath](v))
						if addr != c.attrs.Ipv4 {
							c.mu.Lock()
							c.attrs.Ipv4 = addr
							c.attrs.IPv4Mask = mask
							c.mu.Unlock()
							sensorCh <- c
						}
					case "Ip6Config":
						addr, mask := getAddr(ctx, 6, dbusx.VariantToValue[dbus.ObjectPath](v))
						if addr != c.attrs.Ipv6 {
							c.mu.Lock()
							c.attrs.Ipv6 = addr
							c.attrs.IPv6Mask = mask
							c.mu.Unlock()
							sensorCh <- c
						}
					}
				}
			}()
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create network connections D-Bus watch.")
		close(sensorCh)
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
			Msg("Unmonitoring address changes.")
	}()
	return sensorCh
}

func newConnection(ctx context.Context, p dbus.ObjectPath) <-chan sensor.Details {
	connCh := make(chan sensor.Details)

	c := &connection{
		path: p,
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorConnectionState,
			IsDiagnostic:    true,
		},
		attrs: &connectionAttributes{
			DataSource: linux.DataSrcDbus,
		},
	}
	connCtx, connCancel := context.WithCancel(ctx)

	// fetch properties for the connection
	req := dbusx.NewBusRequest(connCtx, dbusx.SystemBus).
		Path(p).
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
	c.attrs.Ipv4, c.attrs.IPv4Mask = getAddr(connCtx, 4, ip4ConfigPath)
	ip6ConfigPath, err := dbusx.GetProp[dbus.ObjectPath](req, dbusNMActiveConnIntr+".Ip6Config")
	if err != nil {
		log.Warn().Err(err).Str("connection", c.name).Msg("Could not fetch IPv4 address.")
	}
	c.attrs.Ipv6, c.attrs.IPv6Mask = getAddr(connCtx, 6, ip6ConfigPath)

	// send the initial connection state as a sensor
	go func() {
		connCh <- c
	}()

	// monitor connection state changes
	go func() {
		for s := range c.monitorConnectionState(connCtx) {
			connCh <- s
		}
		connCancel()
	}()

	// monitor address changes
	go func() {
		for s := range c.monitorAddresses(connCtx) {
			connCh <- s
		}
	}()

	// monitor for additional states depending on the type of connection
	switch c.attrs.ConnectionType {
	case "802-11-wireless":
		go func() {
			for s := range monitorWifi(connCtx, c.path) {
				connCh <- s
			}
		}()
	}

	go func() {
		defer close(connCh)
		<-connCtx.Done()
		log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
			Msg("Stopped monitoring connection.")
	}()

	return connCh
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
		conn := newConnection(ctx, path)
		go func() {
			for s := range conn {
				sensorCh <- s
			}
			tracker.Untrack(path)
		}()
	}

	// watch for any connection activations/deactivations
	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(dbusNMActiveConnPath),
			dbus.WithMatchArg(0, dbusNMActiveConnIntr),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), dbusNMActiveConnPath) {
				return
			}
			if !tracker.Tracked(s.Path) {
				tracker.Track(s.Path)
				handleConn(s.Path)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to monitor for network connection state changes.")
		close(sensorCh)
	}

	// monitor all current active connections
	for _, path := range tracker.list {
		handleConn(path)
	}

	go func() {
		<-ctx.Done()
		log.Debug().Msg("Stopped network connection monitoring.")
	}()
	return sensorCh
}
