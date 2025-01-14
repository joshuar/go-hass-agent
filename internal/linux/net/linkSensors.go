// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package net

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"

	"github.com/jsimonetti/rtnetlink"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	addressWorkerID = "network_addresses"
)

var (
	ifaceFilters = []string{"lo"}
	addrFamilies = []uint8{uint8(unix.AF_INET), uint8(unix.AF_INET6)}
	nlConfig     = &netlink.Config{
		Groups: unix.RTNLGRP_LINK | unix.RTNLGRP_IPV4_NETCONF | unix.RTNLGRP_IPV6_NETCONF,
	}
)

func newLinkSensor(msg rtnetlink.LinkMessage) sensor.Entity {
	var (
		value any
		icon  string
	)

	name := msg.Attributes.Name
	attributes := map[string]any{
		"data_source": linux.DataSrcNetlink,
	}

	switch msg.Attributes.OperationalState {
	case rtnetlink.OperStateUp:
		value = "up"
		icon = "mdi:network"
	case rtnetlink.OperStateNotPresent:
		value = "invalid"
		icon = "mdi:close-network"
	case rtnetlink.OperStateDown:
		value = "down"
		icon = "mdi:network-off"
	}

	if msg.Attributes.Info != nil {
		attributes["link_type"] = msg.Attributes.Info.Kind
	}

	return sensor.NewSensor(
		sensor.WithName(name+" Link State"),
		sensor.WithID(strings.ToLower(name+"_link_state")),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon(icon),
			sensor.WithValue(value),
			sensor.WithAttributes(attributes),
		),
	)
}

func newAddressSensor(link rtnetlink.LinkMessage, msg rtnetlink.AddressMessage) sensor.Entity {
	name := link.Attributes.Name
	value := msg.Attributes.Address.String()
	attributes := map[string]any{
		"data_source": linux.DataSrcNetlink,
	}

	if link.Attributes.OperationalState != rtnetlink.OperStateUp {
		if ipFamily(msg.Family) == unix.AF_INET {
			value = net.IPv4zero.String()
		} else {
			value = net.IPv6zero.String()
		}
	}

	if msg.Attributes.Broadcast != nil {
		attributes["broadcast"] = msg.Attributes.Broadcast.String()
	}

	if msg.Attributes.Local != nil {
		attributes["local"] = msg.Attributes.Local.String()
	}

	if msg.Attributes.Multicast != nil {
		attributes["multicast"] = msg.Attributes.Multicast.String()
	}

	if msg.Attributes.Anycast != nil {
		attributes["anycast"] = msg.Attributes.Anycast.String()
	}

	return sensor.NewSensor(
		sensor.WithName(name+" "+ipFamily(msg.Family).String()+" Address"),
		sensor.WithID(strings.ToLower(name+"_"+ipFamily(msg.Family).String()+"_address")),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon(ipFamily(msg.Family).Icon()),
			sensor.WithValue(value),
			sensor.WithAttributes(attributes),
		),
	)
}

type ipFamily int

func (f ipFamily) String() string {
	switch {
	case f == unix.AF_INET:
		return "IPv4"
	case f == unix.AF_INET6:
		return "IPv6"
	default:
		return "unknown"
	}
}

func (f ipFamily) Icon() string {
	switch {
	case f == unix.AF_INET:
		return "mdi:numeric-4-box-outline"
	case f == unix.AF_INET6:
		return "mdi:numeric-6-box-outline"
	default:
		return "mdi:help"
	}
}

type AddressWorker struct {
	nlconn *rtnetlink.Conn
	donech chan struct{}
	prefs  *WorkerPrefs
	linux.EventSensorWorker
}

func (w *AddressWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	addresses := w.getAddresses(ctx)
	links := w.getLinks(ctx)

	var sensors []sensor.Entity //nolint:prealloc

	for _, addressSensor := range addresses {
		sensors = append(sensors, *addressSensor)
	}

	for _, linkSensor := range links {
		sensors = append(sensors, *linkSensor)
	}

	return sensors, nil
}

// TODO: reduce complexity?
//
//nolint:gocognit
func (w *AddressWorker) Events(ctx context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)

	// Get all current addresses and send as sensors.
	sensors, err := w.Sensors(ctx)
	if err != nil {
		logging.FromContext(ctx).Debug("Could not get address sensors.", slog.Any("error", err))
	} else {
		for _, addressSensor := range sensors {
			go func() {
				sensorCh <- addressSensor
			}()
		}
	}

	done := false

	go func() {
		defer close(w.donech)
		<-ctx.Done()

		if err := w.nlconn.Close(); err != nil {
			logging.FromContext(ctx).Debug("Could not close netlink connection.",
				slog.Any("error", err))
		}
	}()

	// Listen for address changes and generate new sensors.
	go func() {
		defer close(sensorCh)

		for !done {
			select {
			case <-w.donech:
				done = true
			default:
				nlmsgs, _, err := w.nlconn.Receive()
				if err != nil {
					logging.FromContext(ctx).Debug("Error closing netlink connection.", slog.Any("error", err))
					break
				}

				for _, msg := range nlmsgs {
					switch value := any(msg).(type) {
					case *rtnetlink.LinkMessage:
						link := filterLink(*value)
						if link != nil {
							sensorCh <- *link
						}
					case *rtnetlink.AddressMessage:
						addr := w.filterAddress(*value)
						if addr != nil {
							sensorCh <- *addr
						}
					}
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *AddressWorker) PreferencesID() string {
	return preferencesID
}

func (w *AddressWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{
		IgnoredDevices: defaultIgnoredDevices,
	}
}

func NewAddressWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventSensorWorker(addressWorkerID)

	conn, err := rtnetlink.Dial(nlConfig)
	if err != nil {
		return worker, fmt.Errorf("could not connect to netlink: %w", err)
	}

	addressWorker := &AddressWorker{
		nlconn: conn,
		donech: make(chan struct{}),
	}

	addressWorker.prefs, err = preferences.LoadWorker(ctx, addressWorker)
	if err != nil {
		return worker, fmt.Errorf("could not load preferences: %w", err)
	}

	// If disabled, don't use the addressWorker.
	if addressWorker.prefs.Disabled {
		return worker, nil
	}

	worker.EventSensorType = addressWorker

	return worker, nil
}

func (w *AddressWorker) getAddresses(ctx context.Context) []*sensor.Entity {
	var addrs []*sensor.Entity

	// Request a list of addresses
	msgs, err := w.nlconn.Address.List()
	if err != nil {
		logging.FromContext(ctx).Debug("Could not retrieve address list from netlink.",
			slog.Any("error", err))
	}

	// Filter for valid addresses.
	for _, msg := range msgs {
		if addr := w.filterAddress(msg); addr != nil {
			addrs = append(addrs, addr)
		}
	}

	return addrs
}

func (w *AddressWorker) filterAddress(msg rtnetlink.AddressMessage) *sensor.Entity {
	// Only include Ipv4/6 addresses.
	if !slices.Contains(addrFamilies, msg.Family) {
		return nil
	}
	// Only include global addresses.
	if msg.Scope != unix.RT_SCOPE_UNIVERSE {
		return nil
	}

	link, err := w.nlconn.Link.Get(msg.Index)
	if err != nil {
		return nil
	}

	if link.Attributes.Name == loopbackDeviceName {
		return nil
	}

	// Skip ignored devices.
	if slices.ContainsFunc(w.prefs.IgnoredDevices, func(e string) bool {
		return strings.HasPrefix(link.Attributes.Name, e)
	}) {
		return nil
	}

	s := newAddressSensor(link, msg)

	return &s
}

func (w *AddressWorker) getLinks(ctx context.Context) []*sensor.Entity {
	var links []*sensor.Entity

	// Request a list of addresses
	msgs, err := w.nlconn.Link.List()
	if err != nil {
		logging.FromContext(ctx).Debug("Could not retrieve link list from netlink.",
			slog.Any("error", err))
	}

	// Filter for valid links.
	for _, msg := range msgs {
		if link := filterLink(msg); link != nil {
			links = append(links, link)
		}
	}

	return links
}

func filterLink(msg rtnetlink.LinkMessage) *sensor.Entity {
	if slices.Contains(ifaceFilters, msg.Attributes.Name) {
		return nil
	}

	s := newLinkSensor(msg)

	return &s
}
