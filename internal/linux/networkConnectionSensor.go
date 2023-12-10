// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"slices"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
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
	name  string
	state connState
	attrs *connectionAttributes
	path  dbus.ObjectPath
	linuxSensor
}

type connectionAttributes struct {
	ConnectionType string `json:"Connection Type,omitempty"`
	Ipv4           string `json:"IPv4 Address,omitempty"`
	IPv4Mask       int    `json:"IPv4 Mask,omitempty"`
	Ipv6           string `json:"IPv6 Address,omitempty"`
	IPv6Mask       int    `json:"IPv6 Mask,omitempty"`
	DataSource     string `json:"Data Source"`
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

func (c *connection) Attributes() interface{} {
	return c.attrs
}

func (c *connection) State() interface{} {
	return c.state.String()
}

func (c *connection) monitorConnectionState(ctx context.Context, sensorCh chan tracker.Sensor, p dbus.ObjectPath) {
	log.Debug().Str("path", string(p)).Str("connection", c.name).
		Msg("Monitoring connection state.")
	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(dbusNMActiveConnPath),
			dbus.WithMatchInterface(dbusNMActiveConnIntr),
			dbus.WithMatchMember("StateChanged"),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Path != p {
				return
			}
			if len(s.Body) <= 1 {
				log.Debug().Caller().Interface("body", s.Body).Msg("Unexpected body length.")
				return
			}
			props, ok := s.Body[1].(map[string]dbus.Variant)
			if ok {
				state, ok := props["State"]
				if ok {
					c.state = dbushelpers.VariantToValue[connState](state)
					sensorCh <- c
					// if dbushelpers.VariantToValue[uint32](state) == 4 {
					// 	close(c.doneCh)
					// }
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create network connections D-Bus watch.")
	}
}

func (c *connection) monitorAddresses(ctx context.Context, sensorCh chan tracker.Sensor, p dbus.ObjectPath) {
	r := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(p).
		Destination(dBusNMObj)
	v, _ := r.GetProp(dbusNMActiveConnIntr + ".Ip4Config")
	if !v.Signature().Empty() {
		c.attrs.Ipv4, c.attrs.IPv4Mask = getAddr(ctx, 4, dbushelpers.VariantToValue[dbus.ObjectPath](v))
	}
	v, _ = r.GetProp(dbusNMActiveConnIntr + ".Ip6Config")
	if !v.Signature().Empty() {
		c.attrs.Ipv6, c.attrs.IPv6Mask = getAddr(ctx, 6, dbushelpers.VariantToValue[dbus.ObjectPath](v))
	}
	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(dbusNMActiveConnPath),
			dbus.WithMatchInterface(dbusNMActiveConnIntr),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Path != p {
				return
			}
			if len(s.Body) <= 1 {
				log.Debug().Caller().Interface("body", s.Body).Msg("Unexpected body length.")
				return
			}
			props, ok := s.Body[1].(map[string]dbus.Variant)
			if ok {
				p, ok := props["Ip4Config"]
				if ok {
					c.attrs.Ipv4, c.attrs.IPv4Mask = getAddr(ctx, 4, p.Value().(dbus.ObjectPath))
					sensorCh <- c
				}
				p, ok = props["Ip6Config"]
				if ok {
					c.attrs.Ipv6, c.attrs.IPv6Mask = getAddr(ctx, 6, p.Value().(dbus.ObjectPath))
					sensorCh <- c
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create network connections D-Bus watch.")
	}
}

func (c *connection) monitor(ctx context.Context, sensorCh chan tracker.Sensor) {
	connCtx, _ := context.WithCancel(ctx)
	c.monitorConnectionState(connCtx, sensorCh, c.path)
	c.monitorAddresses(connCtx, sensorCh, c.path)
	switch c.attrs.ConnectionType {
	case "802-11-wireless":
		getWifiProperties(connCtx, sensorCh, c.path)
	}
	// go func() {
	// 	<-c.doneCh
	// 	cancelFunc()
	// }()
}

func newConnection(ctx context.Context, p dbus.ObjectPath) *connection {
	c := &connection{
		attrs: &connectionAttributes{
			DataSource: srcDbus,
		},
		// doneCh: make(chan struct{}),
		path: p,
	}
	c.sensorType = connectionState
	c.diagnostic = true

	r := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(p).
		Destination(dBusNMObj)
	v, _ := r.GetProp(dbusNMActiveConnIntr + ".Id")
	if !v.Signature().Empty() {
		c.name = dbushelpers.VariantToValue[string](v)
	}
	v, _ = r.GetProp(dbusNMActiveConnIntr + ".State")
	if !v.Signature().Empty() {
		c.state = dbushelpers.VariantToValue[connState](v)
	}
	v, _ = r.GetProp(dbusNMActiveConnIntr + ".Type")
	if !v.Signature().Empty() {
		c.attrs.ConnectionType = dbushelpers.VariantToValue[string](v)
	}
	return c
}

func getAddr(ctx context.Context, ver int, path dbus.ObjectPath) (addr string, mask int) {
	if path == "/" {
		return
	}
	var connProp string
	switch ver {
	case 4:
		connProp = dBusNMObj + ".IP4Config"
	case 6:
		connProp = dBusNMObj + ".IP6Config"
	}
	v, err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(path).
		Destination(dBusNMObj).
		GetProp(connProp + ".AddressData")
	if err != nil {
		return
	}
	addrDetails := dbushelpers.VariantToValue[[]map[string]dbus.Variant](v)
	var address string
	var prefix int
	if len(addrDetails) > 0 {
		address = dbushelpers.VariantToValue[string](addrDetails[0]["address"])
		prefix = dbushelpers.VariantToValue[int](addrDetails[0]["prefix"])
		log.Debug().Str("path", string(path)).Str("address", address).Int("prefix", prefix).
			Msg("Retrieved address.")
	}
	return address, prefix
}

func getActiveConnections(ctx context.Context) []dbus.ObjectPath {
	v, err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(dBusNMPath).
		Destination(dBusNMObj).
		GetProp(dBusNMObj + ".ActiveConnections")
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve active connection list.")
		return nil
	}
	return dbushelpers.VariantToValue[[]dbus.ObjectPath](v)
}

func monitorActiveConnections(ctx context.Context, sensorCh chan tracker.Sensor, conns []dbus.ObjectPath) {
	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(dbusNMActiveConnPath),
			dbus.WithMatchArg(0, dbusNMActiveConnIntr),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), dbusNMActiveConnPath) {
				return
			}
			if !slices.Contains(conns, s.Path) {
				conn := newConnection(ctx, s.Path)
				conn.monitor(ctx, sensorCh)
				conns = append(conns, s.Path)
				sensorCh <- conn
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create connection state change D-Bus watch.")
	}
}

func NetworkConnectionsUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor)

	conns := getActiveConnections(ctx)
	go func() {
		for _, p := range conns {
			conn := newConnection(ctx, p)
			conn.monitor(ctx, sensorCh)
		}
	}()
	go monitorActiveConnections(ctx, sensorCh, conns)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped network connection state sensors.")
	}()
	return sensorCh
}
