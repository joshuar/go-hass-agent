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
					updateCh <- c
					if dbushelpers.VariantToValue[uint32](state) == 4 {
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

func getActiveConnections(ctx context.Context, updateCh chan interface{}) {
	v, err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(dBusNMPath).
		Destination(dBusNMObj).
		GetProp(dBusNMObj + ".ActiveConnections")
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve active connection list.")
		return
	}
	paths := dbushelpers.VariantToValue[[]dbus.ObjectPath](v)

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
	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(dbusNMActiveConnPath),
			dbus.WithMatchArg(0, dbusNMActiveConnIntr),
		}).
		Handler(func(s *dbus.Signal) {
			if !strings.Contains(string(s.Path), dbusNMActiveConnPath) {
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
