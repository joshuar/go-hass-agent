// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"context"
	"slices"
	"strings"

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
	return c.state.String()
}

func (c *connection) monitorConnectionState(ctx context.Context) chan sensor.Details {
	log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
		Msg("Monitoring connection state.")
	sensorCh := make(chan sensor.Details, 1)
	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(dbusNMActiveConnPath),
			dbus.WithMatchInterface(dbusNMActiveConnIntr),
			dbus.WithMatchMember("StateChanged"),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Path != c.path || len(s.Body) <= 1 {
				log.Trace().Caller().Msg("Not my signal or empty signal body.")
				return
			}
			var props map[string]dbus.Variant
			var ok bool
			if props, ok = s.Body[1].(map[string]dbus.Variant); !ok {
				log.Trace().Caller().
					Msgf("Could not cast signal body, got %T, want %T", s.Body, props)
				return
			}
			if state, ok := props["State"]; ok {
				currentState := dbusx.VariantToValue[connState](state)
				switch {
				case currentState == 4:
					log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
						Msg("Unmonitoring connection state.")
					close(sensorCh)
				case currentState != c.state:
					c.state = currentState
					sensorCh <- c
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
	sensorCh := make(chan sensor.Details, 1)
	go func() {
		req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
			Path(c.path).
			Destination(dBusNMObj)
		v, err := dbusx.GetProp[dbus.ObjectPath](req, dbusNMActiveConnIntr+".Ip4Config")
		if err != nil {
			log.Warn().Err(err).Msg("Could not retrieve IPv4 address.")
		} else {
			c.attrs.Ipv4, c.attrs.IPv4Mask = getAddr(ctx, 4, v)
		}
		v, err = dbusx.GetProp[dbus.ObjectPath](req, dbusNMActiveConnIntr+".Ip6Config")
		if err != nil {
			log.Warn().Err(err).Msg("Could not retrieve IPv6 address.")
		} else {
			c.attrs.Ipv6, c.attrs.IPv6Mask = getAddr(ctx, 6, v)
		}
		sensorCh <- c
	}()
	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(dbusNMActiveConnPath),
			dbus.WithMatchInterface(dbusNMActiveConnIntr),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Path != c.path || len(s.Body) <= 1 {
				log.Trace().Caller().Msg("Not my signal or empty signal body.")
				return
			}
			var props map[string]dbus.Variant
			var ok bool
			if props, ok = s.Body[1].(map[string]dbus.Variant); !ok {
				log.Trace().Caller().
					Msgf("Could not cast signal body, got %T, want %T", s.Body, props)
				return
			}
			go func() {
				for k, v := range props {
					switch k {
					case "Ip4Config":
						addr, mask := getAddr(ctx, 4, dbusx.VariantToValue[dbus.ObjectPath](v))
						if addr != c.attrs.Ipv4 {
							c.attrs.Ipv4 = addr
							c.attrs.IPv4Mask = mask
							sensorCh <- c
						}
					case "Ip6Config":
						addr, mask := getAddr(ctx, 6, dbusx.VariantToValue[dbus.ObjectPath](v))
						if addr != c.attrs.Ipv6 {
							c.attrs.Ipv6 = addr
							c.attrs.IPv6Mask = mask
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

func (c *connection) monitor(ctx context.Context) <-chan sensor.Details {
	log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
		Msg("Monitoring connection.")
	var outCh []<-chan sensor.Details
	connCtx, cancelFunc := context.WithCancel(ctx)
	go func() {
		sensorCh := make(chan sensor.Details, 1)
		defer close(sensorCh)
		outCh = append(outCh, sensorCh)
		for s := range c.monitorConnectionState(connCtx) {
			sensorCh <- s
		}
		log.Debug().Str("connection", c.Name()).Str("path", string(c.path)).
			Msg("Unmonitoring connection.")
		cancelFunc()
	}()
	outCh = append(outCh, c.monitorAddresses(connCtx))
	switch c.attrs.ConnectionType {
	case "802-11-wireless":
		outCh = append(outCh, getWifiProperties(connCtx, c.path))
	}
	return sensor.MergeSensorCh(ctx, outCh...)
}

func newConnection(ctx context.Context, p dbus.ObjectPath) *connection {
	c := &connection{
		attrs: &connectionAttributes{
			DataSource: linux.DataSrcDbus,
		},
		// doneCh: make(chan struct{}),
		path: p,
	}
	c.SensorTypeValue = linux.SensorConnectionState
	c.IsDiagnostic = true

	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
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
	return c
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

func monitorActiveConnections(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details, 1)
	conns := getActiveConnections(ctx)

	handleConn := func(path dbus.ObjectPath) {
		conn := newConnection(ctx, path)
		sensorCh <- conn
		for c := range conn.monitor(ctx) {
			sensorCh <- c
		}
		if i := slices.Index(conns, path); i > 0 {
			slices.Delete(conns, i, i)
		}
	}

	for _, p := range conns {
		go handleConn(p)
	}

	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(dbusNMActiveConnPath),
			dbus.WithMatchArg(0, dbusNMActiveConnIntr),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), dbusNMActiveConnPath) {
				return
			}
			if !slices.Contains(conns, s.Path) {
				conns = append(conns, s.Path)
				go handleConn(s.Path)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create connection state change D-Bus watch.")
		close(sensorCh)
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
	}()
	return sensorCh
}

func ConnectionsUpdater(ctx context.Context) chan sensor.Details {
	return monitorActiveConnections(ctx)
}
