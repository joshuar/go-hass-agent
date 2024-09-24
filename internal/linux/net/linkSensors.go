// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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

type linkSensor struct {
	attributes map[string]any
	name       string
	linux.Sensor
	state rtnetlink.OperationalState
}

func (s *linkSensor) Name() string {
	return s.name + " Link State"
}

func (s *linkSensor) ID() string {
	return strings.ToLower(s.name + "_link_state")
}

func (s *linkSensor) Icon() string {
	switch s.state {
	case rtnetlink.OperStateUp:
		return "mdi:network"
	case rtnetlink.OperStateNotPresent:
		return "mdi:close-network"
	default:
		return "mdi:network-off"
	}
}

func (s *linkSensor) State() any {
	switch s.state {
	case rtnetlink.OperStateUp:
		return "up"
	case rtnetlink.OperStateNotPresent:
		return "invalid"
	default:
		return "down"
	}
}

func (s *linkSensor) Attributes() map[string]any {
	return s.attributes
}

func newLinkSensor(msg rtnetlink.LinkMessage) *linkSensor {
	link := &linkSensor{
		name:       msg.Attributes.Name,
		state:      msg.Attributes.OperationalState,
		attributes: make(map[string]any),
	}

	link.attributes["data_source"] = linux.DataSrcNetlink

	if msg.Attributes.Info != nil {
		link.attributes["link_type"] = msg.Attributes.Info.Kind
	}

	link.IsDiagnostic = true

	return link
}

type addressSensor struct {
	attributes map[string]any
	name       string
	address    net.IP
	linux.Sensor
	family ipFamily
}

func (s *addressSensor) Name() string {
	return s.name + " " + s.family.String() + " Address"
}

func (s *addressSensor) ID() string {
	return strings.ToLower(s.name + "_" + s.family.String() + "_address")
}

func (s *addressSensor) Icon() string {
	return s.family.Icon()
}

func (s *addressSensor) State() any {
	return s.address.String()
}

func (s *addressSensor) Attributes() map[string]any {
	return s.attributes
}

func newAddressSensor(link rtnetlink.LinkMessage, msg rtnetlink.AddressMessage) *addressSensor {
	addr := &addressSensor{
		name:       link.Attributes.Name,
		family:     ipFamily(msg.Family),
		address:    msg.Attributes.Address,
		attributes: make(map[string]any),
	}

	if link.Attributes.OperationalState != rtnetlink.OperStateUp {
		if addr.family == unix.AF_INET {
			addr.address = net.IPv4zero
		}

		addr.address = net.IPv4zero
	}

	addr.attributes["data_source"] = linux.DataSrcNetlink

	if msg.Attributes.Broadcast != nil {
		addr.attributes["broadcast"] = msg.Attributes.Broadcast.String()
	}

	if msg.Attributes.Local != nil {
		addr.attributes["local"] = msg.Attributes.Local.String()
	}

	if msg.Attributes.Multicast != nil {
		addr.attributes["multicast"] = msg.Attributes.Multicast.String()
	}

	if msg.Attributes.Anycast != nil {
		addr.attributes["anycast"] = msg.Attributes.Anycast.String()
	}

	// Address sensors are diagnostic category.
	addr.IsDiagnostic = true

	return addr
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
	linux.EventSensorWorker
}

func (w *AddressWorker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	addresses := w.getAddresses(ctx)
	links := w.getLinks(ctx)

	sensors := make([]sensor.Details, len(addresses)+len(links))

	for _, addressSensor := range addresses {
		sensors = append(sensors, addressSensor)
	}

	for _, linkSensor := range links {
		sensors = append(sensors, linkSensor)
	}

	return sensors, nil
}

// TODO: reduce complexity?
//
//nolint:gocognit
func (w *AddressWorker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

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

	// Listen for address changes and generate new sensors.
	go func() {
		defer close(sensorCh)

		done := false

		for !done {
			select {
			case <-ctx.Done():
				done = true
			default:
				nlmsgs, _, err := w.nlconn.Receive()
				if err != nil {
					slog.Error("received error", slog.Any("error", err))
					break
				}

				for _, msg := range nlmsgs {
					switch value := any(msg).(type) {
					case *rtnetlink.LinkMessage:
						link := filterLink(*value)
						if link != nil {
							sensorCh <- link
						}
					case *rtnetlink.AddressMessage:
						addr := w.filterAddress(*value)
						if addr != nil {
							sensorCh <- addr
						}
					}
				}
			}
		}
	}()

	return sensorCh, nil
}

func NewAddressWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventWorker(addressWorkerID)

	conn, err := rtnetlink.Dial(nlConfig)
	if err != nil {
		return worker, fmt.Errorf("could not connect to netlink: %w", err)
	}

	go func() {
		<-ctx.Done()

		if err := conn.Close(); err != nil {
			logging.FromContext(ctx).Debug("Could not close netlink connection.",
				slog.Any("error", err))
		}
	}()

	addressWorker := &AddressWorker{nlconn: conn}

	worker.EventType = addressWorker

	return worker, nil
}

func (w *AddressWorker) getAddresses(ctx context.Context) []*addressSensor {
	var addrs []*addressSensor

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

func (w *AddressWorker) filterAddress(msg rtnetlink.AddressMessage) *addressSensor {
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
	// Ignore addresses from unwanted links.
	if slices.Contains(ifaceFilters, link.Attributes.Name) {
		return nil
	}

	return newAddressSensor(link, msg)
}

func (w *AddressWorker) getLinks(ctx context.Context) []*linkSensor {
	var links []*linkSensor

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

func filterLink(msg rtnetlink.LinkMessage) *linkSensor {
	if slices.Contains(ifaceFilters, msg.Attributes.Name) {
		return nil
	}

	return newLinkSensor(msg)
}
