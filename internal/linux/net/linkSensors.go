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
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
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

func newLinkSensor(msg rtnetlink.LinkMessage) *sensor.Entity {
	name := msg.Attributes.Name

	link := &sensor.Entity{
		Name:     name + " Link State",
		Category: types.CategoryDiagnostic,
		State: &sensor.State{
			ID: strings.ToLower(name + "_link_state"),
			Attributes: map[string]any{
				"data_source": linux.DataSrcNetlink,
			},
		},
	}

	switch msg.Attributes.OperationalState {
	case rtnetlink.OperStateUp:
		link.Value = "up"
		link.Icon = "mdi:network"
	case rtnetlink.OperStateNotPresent:
		link.Value = "invalid"
		link.Icon = "mdi:close-network"
	case rtnetlink.OperStateUnknown:
		link.Value = "unknown"
		link.Icon = "mdi:help-network"
	default:
		link.Value = "down"
		link.Icon = "mdi:network-off"
	}

	if msg.Attributes.Info != nil {
		link.Attributes["link_type"] = msg.Attributes.Info.Kind
	}

	return link
}

func newAddressSensor(link rtnetlink.LinkMessage, msg rtnetlink.AddressMessage) *sensor.Entity {
	name := link.Attributes.Name

	addr := &sensor.Entity{
		Name:     name + " " + ipFamily(msg.Family).String() + " Address",
		Category: types.CategoryDiagnostic,
		State: &sensor.State{
			ID:    strings.ToLower(name + "_" + ipFamily(msg.Family).String() + "_address"),
			Icon:  ipFamily(msg.Family).Icon(),
			Value: msg.Attributes.Address.String(),
			Attributes: map[string]any{
				"data_source": linux.DataSrcNetlink,
			},
		},
	}

	if link.Attributes.OperationalState != rtnetlink.OperStateUp {
		if ipFamily(msg.Family) == unix.AF_INET {
			addr.Value = net.IPv4zero.String()
		} else {
			addr.Value = net.IPv6zero.String()
		}
	}

	if msg.Attributes.Broadcast != nil {
		addr.Attributes["broadcast"] = msg.Attributes.Broadcast.String()
	}

	if msg.Attributes.Local != nil {
		addr.Attributes["local"] = msg.Attributes.Local.String()
	}

	if msg.Attributes.Multicast != nil {
		addr.Attributes["multicast"] = msg.Attributes.Multicast.String()
	}

	if msg.Attributes.Anycast != nil {
		addr.Attributes["anycast"] = msg.Attributes.Anycast.String()
	}

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
	// Ignore addresses from unwanted links.
	if slices.Contains(ifaceFilters, link.Attributes.Name) {
		return nil
	}

	return newAddressSensor(link, msg)
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

	return newLinkSensor(msg)
}
