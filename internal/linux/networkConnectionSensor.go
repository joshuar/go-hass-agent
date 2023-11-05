// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/device"
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

	dBusNMPath = "/org/freedesktop/NetworkManager"
	dBusNMObj  = "org.freedesktop.NetworkManager"
)

type connState uint32

type connections struct {
	list map[dbus.ObjectPath]*connection
	mu   sync.Mutex
}

type connection struct {
	name   string
	state  connState
	attrs  *connectionAttributes
	doneCh chan struct{}
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

func (c *connection) monitorConnectionState(ctx context.Context, updateCh chan interface{}, p dbus.ObjectPath) {
	log.Debug().Str("path", string(p)).Str("connection", c.name).
		Msg("Monitoring connection state.")
	err := NewBusRequest(ctx, SystemBus).
		Path(p).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(dBusNMPath + "/ActiveConnection"),
		}).
		Event(dBusNMObj + ".Connection.Active.StateChanged").
		Handler(func(s *dbus.Signal) {
			props, ok := s.Body[1].(map[string]dbus.Variant)
			if ok {
				state, ok := props["State"]
				if ok {
					c.state = variantToValue[connState](state)
					updateCh <- c
					if variantToValue[uint32](state) == 4 {
						close(c.doneCh)
					}
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create network connections D-Bus watch.")
	}
}

func (c *connection) monitorAddresses(ctx context.Context, updateCh chan interface{}, p dbus.ObjectPath) {
	r := NewBusRequest(ctx, SystemBus).
		Path(p).
		Destination(dBusNMObj)
	propBase := dBusNMObj + ".Connection.Active"
	v, _ := r.GetProp(propBase + ".Ip4Config")
	if !v.Signature().Empty() {
		c.attrs.Ipv4, c.attrs.IPv4Mask = getAddr(ctx, 4, variantToValue[dbus.ObjectPath](v))
	}
	v, _ = r.GetProp(propBase + ".Ip6Config")
	if !v.Signature().Empty() {
		c.attrs.Ipv6, c.attrs.IPv6Mask = getAddr(ctx, 6, variantToValue[dbus.ObjectPath](v))
	}
	err := NewBusRequest(ctx, SystemBus).
		Path(p).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(dBusNMPath + "/ActiveConnection"),
			dbus.WithMatchArg0Namespace("org.freedesktop.NetworkManager.Connection.Active"),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(func(s *dbus.Signal) {
			props, ok := s.Body[1].(map[string]dbus.Variant)
			if ok {
				p, ok := props["Ip4Config"]
				if ok {
					c.attrs.Ipv4, c.attrs.IPv4Mask = getAddr(ctx, 4, p.Value().(dbus.ObjectPath))
					updateCh <- c
				}
				p, ok = props["Ip6Config"]
				if ok {
					c.attrs.Ipv6, c.attrs.IPv6Mask = getAddr(ctx, 6, p.Value().(dbus.ObjectPath))
					updateCh <- c
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create network connections D-Bus watch.")
	}
}

func newConnection(ctx context.Context, updateCh chan interface{}, p dbus.ObjectPath) *connection {
	c := &connection{
		attrs: &connectionAttributes{
			DataSource: srcDbus,
		},
		doneCh: make(chan struct{}),
	}
	c.sensorType = connectionState
	c.diagnostic = true

	r := NewBusRequest(ctx, SystemBus).
		Path(p).
		Destination(dBusNMObj)
	propBase := dBusNMObj + ".Connection.Active"
	v, _ := r.GetProp(propBase + ".Id")
	if !v.Signature().Empty() {
		c.name = variantToValue[string](v)
	}
	v, _ = r.GetProp(propBase + ".State")
	if !v.Signature().Empty() {
		c.state = variantToValue[connState](v)
	}
	v, _ = r.GetProp(propBase + ".Type")
	if !v.Signature().Empty() {
		c.attrs.ConnectionType = variantToValue[string](v)
	}
	connCtx, cancelFunc := context.WithCancel(ctx)
	c.monitorConnectionState(connCtx, updateCh, p)
	c.monitorAddresses(connCtx, updateCh, p)
	switch c.attrs.ConnectionType {
	case "802-11-wireless":
		getWifiProperties(connCtx, updateCh, p)
	}
	go func() {
		<-c.doneCh
		cancelFunc()
	}()
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
	v, err := NewBusRequest(ctx, SystemBus).
		Path(path).
		Destination(dBusNMObj).
		GetProp(connProp + ".AddressData")
	if err != nil {
		return
	}
	a := variantToValue[[]map[string]dbus.Variant](v)
	return variantToValue[string](a[0]["address"]), variantToValue[int](a[0]["prefix"])
}

func getActiveConnections(ctx context.Context, updateCh chan interface{}) {
	v, err := NewBusRequest(ctx, SystemBus).
		Path(dBusNMPath).
		Destination(dBusNMObj).
		GetProp(dBusNMObj + ".ActiveConnections")
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve active connection list.")
		return
	}
	paths := variantToValue[[]dbus.ObjectPath](v)

	c := &connections{
		list: make(map[dbus.ObjectPath]*connection),
	}

	for _, p := range paths {
		c.list[p] = newConnection(ctx, updateCh, p)
		updateCh <- c.list[p]
	}
	monitorActiveConnections(ctx, updateCh, c)
}

func monitorActiveConnections(ctx context.Context, updateCh chan interface{}, conns *connections) {
	err := NewBusRequest(ctx, SystemBus).
		Path(dBusNMPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(dBusNMPath + "/ActiveConnection"),
			dbus.WithMatchArg(0, dBusNMObj+".Connection.Active"),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), dBusNMPath+"/ActiveConnection") {
				return
			}
			_, ok := conns.list[s.Path]
			if !ok {
				conns.list[s.Path] = newConnection(ctx, updateCh, s.Path)
				updateCh <- conns.list[s.Path]
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create connection state change D-Bus watch.")
	}
}

func NetworkConnectionsUpdater(ctx context.Context, tracker device.SensorTracker) {
	var wg sync.WaitGroup
	sensorCh := make(chan interface{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-sensorCh:
				if err := tracker.UpdateSensors(ctx, s); err != nil {
					log.Error().Err(err).
						Msg("Could not update property.")
				}
			}
		}
	}()
	getActiveConnections(ctx, sensorCh)
	wg.Wait()
}
