// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
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
	slogctx "github.com/veqryn/slog-context"
	"golang.org/x/sys/unix"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	addressWorkerID     = "network_addresses"
	addressWorkerDesc   = "Network interface sensors"
	addressWorkerPrefID = prefPrefix + "links"
)

var _ workers.EntityWorker = (*NetlinkWorker)(nil)

var (
	addrFamilies = []uint8{uint8(unix.AF_INET), uint8(unix.AF_INET6)}
	nlConfig     = &netlink.Config{
		Groups: unix.RTNLGRP_LINK | unix.RTNLGRP_IPV4_NETCONF | unix.RTNLGRP_IPV6_NETCONF,
	}
)

type ipFamily int

func (f ipFamily) String() string {
	switch f {
	case unix.AF_INET:
		return "IPv4"
	case unix.AF_INET6:
		return "IPv6"
	default:
		return "unknown"
	}
}

func (f ipFamily) Icon() string {
	switch f {
	case unix.AF_INET:
		return "mdi:numeric-4-box-outline"
	case unix.AF_INET6:
		return "mdi:numeric-6-box-outline"
	default:
		return "mdi:help"
	}
}

// NetlinkWorker handles generating sensors from netlink.
type NetlinkWorker struct {
	*models.WorkerMetadata

	nlconn *rtnetlink.Conn
	donech chan struct{}
	prefs  *Preferences
}

// NewNetlinkWorker creates a new netlink worker. Once started, this worker will generate entities for network link
// states and addresses.
func NewNetlinkWorker(_ context.Context) (workers.EntityWorker, error) {
	conn, err := rtnetlink.Dial(nlConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to start netlink worker: %w", err)
	}

	worker := &NetlinkWorker{
		WorkerMetadata: models.SetWorkerMetadata(addressWorkerID, addressWorkerDesc),
		nlconn:         conn,
		donech:         make(chan struct{}),
	}

	defaultPrefs := &Preferences{
		IgnoredDevices: defaultIgnoredDevices,
	}
	worker.prefs, err = workers.LoadWorkerPreferences(addressWorkerPrefID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("unable to start netlink worker: %w", err)
	}

	return worker, nil
}

// Start will start the netlink worker. This will generate initial sensors and send updates for address/link state changes.
//
//nolint:gocognit,funlen
func (w *NetlinkWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)

	// Get all current addresses and send as sensors.
	sensors, err := w.generateSensors(ctx)
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Could not get address sensors.", slog.Any("error", err))
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
			slogctx.FromCtx(ctx).Debug("Could not close netlink connection.",
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
					slogctx.FromCtx(ctx).Debug("Error closing netlink connection.", slog.Any("error", err))
					break
				}

				for _, msg := range nlmsgs {
					switch value := any(msg).(type) {
					case *rtnetlink.LinkMessage:
						if slices.ContainsFunc(w.prefs.IgnoredDevices, func(filter string) bool {
							return strings.HasPrefix(value.Attributes.Name, filter)
						}) {
							// slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "Filtering device.",
							// 	slog.String("name", value.Attributes.Name))
							slogctx.FromCtx(ctx).Debug("Filtering device.",
								slog.String("name", value.Attributes.Name))
							continue
						}
						link := newLinkSensor(ctx, *value)
						if link.Valid() {
							sensorCh <- link
						}
					case *rtnetlink.AddressMessage:
						if !w.usableAddress(*value) {
							continue
						}

						addr, err := newAddressSensor(ctx, w.nlconn.Link, *value)
						if err != nil {
							slogctx.FromCtx(ctx).Warn("Could not generate address sensor.", slog.Any("error", err))
							continue
						}

						sensorCh <- *addr
					}
				}
			}
		}
	}()

	return sensorCh, nil
}

// IsDisabled will return a boolean indicating whether the netlink worker has been explicitly disabled in the
// preferences.
func (w *NetlinkWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *NetlinkWorker) generateSensors(ctx context.Context) ([]models.Entity, error) {
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

func (w *NetlinkWorker) getAddresses(ctx context.Context) ([]models.Entity, error) {
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
		if !w.usableAddress(msg) {
			continue
		}

		entity, err := newAddressSensor(ctx, w.nlconn.Link, msg)
		if err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not generate address sensor: %w", err))
		} else if entity.Valid() {
			addrs = append(addrs, *entity)
		}
	}

	return addrs, warnings
}

func (w *NetlinkWorker) usableAddress(msg rtnetlink.AddressMessage) bool {
	// Only include Ipv4/6 addresses.
	if !slices.Contains(addrFamilies, msg.Family) {
		return false
	}
	// Only include global addresses.
	if msg.Scope != unix.RT_SCOPE_UNIVERSE {
		return false
	}

	link, err := w.nlconn.Link.Get(msg.Index)
	if err != nil {
		return false
	}

	if link.Attributes.Name == loopbackDeviceName {
		return false
	}

	// Skip ignored devices.
	if slices.ContainsFunc(w.prefs.IgnoredDevices, func(e string) bool {
		return strings.HasPrefix(link.Attributes.Name, e)
	}) {
		return false
	}

	return true
}

func (w *NetlinkWorker) getLinks(ctx context.Context) ([]models.Entity, error) {
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
	for msg := range slices.Values(msgs) {
		if slices.ContainsFunc(w.prefs.IgnoredDevices, func(filter string) bool {
			return strings.HasPrefix(msg.Attributes.Name, filter)
		}) {
			continue
		}
		entity := newLinkSensor(ctx, msg)
		if entity.Valid() {
			links = append(links, entity)
		}
	}

	return links, warnings
}

func newLinkSensor(ctx context.Context, msg rtnetlink.LinkMessage) models.Entity {
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

func newAddressSensor(ctx context.Context, link *rtnetlink.LinkService, msg rtnetlink.AddressMessage) (*models.Entity, error) {
	linkMsg, err := link.Get(msg.Index)
	if err != nil {
		return nil, fmt.Errorf("unable to generate address sensor: %w", err)
	}

	name := linkMsg.Attributes.Name
	value := msg.Attributes.Address.String()
	attributes := map[string]any{
		"data_source": linux.DataSrcNetlink,
	}

	if linkMsg.Attributes.OperationalState != rtnetlink.OperStateUp {
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

	addrSensor := sensor.NewSensor(ctx,
		sensor.WithName(name+" "+ipFamily(msg.Family).String()+" Address"),
		sensor.WithID(strings.ToLower(name+"_"+ipFamily(msg.Family).String()+"_address")),
		sensor.AsDiagnostic(),
		sensor.WithIcon(ipFamily(msg.Family).Icon()),
		sensor.WithState(value),
		sensor.WithAttributes(attributes),
	)

	return &addrSensor, nil
}
