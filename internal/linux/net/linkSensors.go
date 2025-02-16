// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package net

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"

	"github.com/jsimonetti/rtnetlink"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	addressWorkerID     = "network_addresses"
	addressWorkerPrefID = prefPrefix + "links"
)

var (
	ifaceFilters = []string{"lo"}
	addrFamilies = []uint8{uint8(unix.AF_INET), uint8(unix.AF_INET6)}
	nlConfig     = &netlink.Config{
		Groups: unix.RTNLGRP_LINK | unix.RTNLGRP_IPV4_NETCONF | unix.RTNLGRP_IPV6_NETCONF,
	}
)

var ErrInitLinkWorker = errors.New("could not init network link state worker")

func newLinkSensor(ctx context.Context, msg rtnetlink.LinkMessage) (models.Entity, error) {
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

	return sensor.NewSensor(ctx,
		sensor.WithName(name+" Link State"),
		sensor.WithID(strings.ToLower(name+"_link_state")),
		sensor.AsDiagnostic(),
		sensor.WithIcon(icon),
		sensor.WithState(value),
		sensor.WithAttributes(attributes),
	)
}

func newAddressSensor(ctx context.Context, link rtnetlink.LinkMessage, msg rtnetlink.AddressMessage) (models.Entity, error) {
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

	return sensor.NewSensor(ctx,
		sensor.WithName(name+" "+ipFamily(msg.Family).String()+" Address"),
		sensor.WithID(strings.ToLower(name+"_"+ipFamily(msg.Family).String()+"_address")),
		sensor.AsDiagnostic(),
		sensor.WithIcon(ipFamily(msg.Family).Icon()),
		sensor.WithState(value),
		sensor.WithAttributes(attributes),
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

func (w *AddressWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	var (
		sensors  []models.Entity
		warnings error
	)

	addresses, err := w.getAddresses(ctx)
	if err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("problem fetching address: %w", err))
	}

	sensors = append(sensors, addresses...)

	links, err := w.getLinks(ctx)
	if err != nil {
		warnings = errors.Join(warnings, fmt.Errorf("problem fetching links: %w", err))
	}

	sensors = append(sensors, links...)

	return sensors, warnings
}

// TODO: reduce complexity?
//
//nolint:gocognit
func (w *AddressWorker) Events(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)

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
						link, err := filterLink(ctx, *value)
						if err != nil {
							logging.FromContext(ctx).Warn("Could not generate link sensor.", slog.Any("error", err))
						} else {
							sensorCh <- link
						}
					case *rtnetlink.AddressMessage:
						addr, err := w.filterAddress(ctx, *value)
						if err != nil {
							logging.FromContext(ctx).Warn("Could not generate address sensor.", slog.Any("error", err))
						} else {
							sensorCh <- addr
						}
					}
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *AddressWorker) PreferencesID() string {
	return addressWorkerPrefID
}

func (w *AddressWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{
		IgnoredDevices: defaultIgnoredDevices,
	}
}

func (w *AddressWorker) getAddresses(ctx context.Context) ([]models.Entity, error) {
	var (
		addrs    []models.Entity
		warnings error
	)

	// Request a list of addresses
	msgs, err := w.nlconn.Address.List()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve link addresses from netlink: %w", err)
	}

	// Filter for valid addresses.
	for _, msg := range msgs {
		entity, err := w.filterAddress(ctx, msg)
		if err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not generate address sensor: %w", err))
		} else if entity.Valid() {
			addrs = append(addrs, entity)
		}
	}

	return addrs, warnings
}

func (w *AddressWorker) filterAddress(ctx context.Context, msg rtnetlink.AddressMessage) (models.Entity, error) {
	// Only include Ipv4/6 addresses.
	if !slices.Contains(addrFamilies, msg.Family) {
		return models.Entity{}, nil
	}
	// Only include global addresses.
	if msg.Scope != unix.RT_SCOPE_UNIVERSE {
		return models.Entity{}, nil
	}

	link, err := w.nlconn.Link.Get(msg.Index)
	if err != nil {
		return models.Entity{}, nil
	}

	if link.Attributes.Name == loopbackDeviceName {
		return models.Entity{}, nil
	}

	// Skip ignored devices.
	if slices.ContainsFunc(w.prefs.IgnoredDevices, func(e string) bool {
		return strings.HasPrefix(link.Attributes.Name, e)
	}) {
		return models.Entity{}, nil
	}

	return newAddressSensor(ctx, link, msg)
}

func (w *AddressWorker) getLinks(ctx context.Context) ([]models.Entity, error) {
	var (
		links    []models.Entity
		warnings error
	)

	// Request a list of addresses
	msgs, err := w.nlconn.Link.List()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve links from netlink: %w", err)
	}

	// Filter for valid links.
	for _, msg := range msgs {
		entity, err := filterLink(ctx, msg)
		if err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not generate link sensor: %w", err))
		} else if entity.Valid() {
			links = append(links, entity)
		}
	}

	return links, warnings
}

func filterLink(ctx context.Context, msg rtnetlink.LinkMessage) (models.Entity, error) {
	if slices.Contains(ifaceFilters, msg.Attributes.Name) {
		return models.Entity{}, nil
	}

	return newLinkSensor(ctx, msg)
}

func NewAddressWorker(_ context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventSensorWorker(addressWorkerID)

	conn, err := rtnetlink.Dial(nlConfig)
	if err != nil {
		return worker, errors.Join(ErrInitLinkWorker,
			fmt.Errorf("could not connect to netlink: %w", err))
	}

	addressWorker := &AddressWorker{
		nlconn: conn,
		donech: make(chan struct{}),
	}

	addressWorker.prefs, err = preferences.LoadWorker(addressWorker)
	if err != nil {
		return worker, errors.Join(ErrInitLinkWorker, err)
	}

	// If disabled, don't use the addressWorker.
	if addressWorker.prefs.Disabled {
		return worker, nil
	}

	worker.EventSensorType = addressWorker

	return worker, nil
}
