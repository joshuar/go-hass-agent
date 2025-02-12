// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package net

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jsimonetti/rtnetlink"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	statsWorkerPrefID = prefPrefix + "usage"
)

var ErrInitStatsWorker = errors.New("could not init network stats worker")

type StatsWorkerPrefs struct {
	WorkerPrefs
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of network stats sensors."`
}

// netStatsWorker is the object used for tracking network stats sensors. It
// holds a netlink connection and a map of links with their stats sensors.
type netStatsWorker struct {
	statsSensors map[string]map[netStatsType]*netStatsSensor
	nlconn       *rtnetlink.Conn
	prefs        *StatsWorkerPrefs
	delta        time.Duration
	mu           sync.Mutex
}

// updateTotals takes the total Rx/Tx bytes and updates the total sensors.
func (w *netStatsWorker) updateTotals(totalBytesRx, totalBytesTx uint64) {
	stats := &rtnetlink.LinkStats64{
		RXBytes: totalBytesRx,
		TXBytes: totalBytesTx,
	}
	for _, sensorType := range sensorList {
		w.statsSensors[totalsName][sensorType].update(totalsName, sensorType, stats, w.delta)
	}
}

func (w *netStatsWorker) UpdateDelta(delta time.Duration) {
	w.delta = delta
}

func (w *netStatsWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	// Get all links.
	links, err := w.getLinks()
	if err != nil {
		return nil, fmt.Errorf("error accessing netlink: %w", err)
	}

	// Get link stats, filtering "uninteresting" links.
	stats := w.getLinkStats(links)
	sensors := make([]sensor.Entity, 0, len(stats)*4+4)

	w.mu.Lock()

	// Counters for totals.
	var totalBytesRx, totalBytesTx uint64

	// For each link, update the link sensors with the new stats.
	for _, link := range stats {
		name := link.name
		stats := link.stats
		totalBytesRx += stats.RXBytes
		totalBytesTx += stats.TXBytes

		// Skip ignored devices.
		if slices.ContainsFunc(w.prefs.IgnoredDevices, func(e string) bool {
			return strings.HasPrefix(name, e)
		}) {
			continue
		}

		if _, ok := w.statsSensors[name]; ok { // Existing link/sensors, update.
			for sensorType := range w.statsSensors[name] {
				w.statsSensors[name][sensorType].update(name, sensorType, stats, w.delta)
			}
		} else { // New link, add to tracking map.
			w.statsSensors[name] = generateSensors(name, stats)
		}
		// Create a list of sensors.
		for _, s := range w.statsSensors[name] {
			sensors = append(sensors, s.Entity)
		}
	}

	if len(stats) > 0 {
		// Update the totals sensors based on the counters.
		w.updateTotals(totalBytesRx, totalBytesTx)
		// Append the total sensors to the list of sensors.
		for _, s := range w.statsSensors[totalsName] {
			sensors = append(sensors, s.Entity)
		}
	}

	w.mu.Unlock()

	return sensors, nil
}

func (w *netStatsWorker) PreferencesID() string {
	return statsWorkerPrefID
}

func (w *netStatsWorker) DefaultPreferences() StatsWorkerPrefs {
	prefs := StatsWorkerPrefs{
		UpdateInterval: rateInterval.String(),
	}

	prefs.IgnoredDevices = defaultIgnoredDevices

	return prefs
}

// getLinks returns all available links on this device. If a problem occurred, a
// non-nil error is returned.
func (w *netStatsWorker) getLinks() ([]rtnetlink.LinkMessage, error) {
	links, err := w.nlconn.Link.List()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of devices from netlink: %w", err)
	}

	return links, nil
}

// getLinkStats collates the network stats for all links. It filters
// links to those with stats, not in an exclusion list and currently active.
func (w *netStatsWorker) getLinkStats(links []rtnetlink.LinkMessage) []linkStats {
	allLinkStats := make([]linkStats, 0, len(links))

	for _, msg := range links {
		if msg.Attributes == nil {
			continue
		}

		// Ignore loopback.
		if msg.Attributes.Name == loopbackDeviceName {
			continue
		}

		// Ignore devices that are not currently active.
		if *msg.Attributes.Carrier == 0 {
			continue
		}

		name := msg.Attributes.Name
		stats := msg.Attributes.Stats64

		if stats32 := msg.Attributes.Stats; stats == nil && stats32 != nil {
			stats = &rtnetlink.LinkStats64{
				RXPackets:          uint64(stats32.RXPackets),
				TXPackets:          uint64(stats32.TXPackets),
				RXBytes:            uint64(stats32.RXBytes),
				TXBytes:            uint64(stats32.TXBytes),
				RXErrors:           uint64(stats32.RXErrors),
				TXErrors:           uint64(stats32.TXErrors),
				RXDropped:          uint64(stats32.RXDropped),
				TXDropped:          uint64(stats32.TXDropped),
				Multicast:          uint64(stats32.Multicast),
				Collisions:         uint64(stats32.Collisions),
				RXLengthErrors:     uint64(stats32.RXLengthErrors),
				RXOverErrors:       uint64(stats32.RXOverErrors),
				RXCRCErrors:        uint64(stats32.RXCRCErrors),
				RXFrameErrors:      uint64(stats32.RXFrameErrors),
				RXFIFOErrors:       uint64(stats32.RXFIFOErrors),
				RXMissedErrors:     uint64(stats32.RXMissedErrors),
				TXAbortedErrors:    uint64(stats32.TXAbortedErrors),
				TXCarrierErrors:    uint64(stats32.TXCarrierErrors),
				TXFIFOErrors:       uint64(stats32.TXFIFOErrors),
				TXHeartbeatErrors:  uint64(stats32.TXHeartbeatErrors),
				TXWindowErrors:     uint64(stats32.TXWindowErrors),
				RXCompressed:       uint64(stats32.RXCompressed),
				TXCompressed:       uint64(stats32.TXCompressed),
				RXNoHandler:        uint64(stats32.RXNoHandler),
				RXOtherhostDropped: 0,
			}
		}

		if stats != nil {
			allLinkStats = append(allLinkStats, linkStats{
				name:  name,
				stats: stats,
			})
		}
	}

	return allLinkStats
}

// NewNetStatsWorker sets up a sensor worker that tracks network stats.
func NewNetStatsWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		return nil, errors.Join(ErrInitStatsWorker,
			fmt.Errorf("could not connect to netlink: %w", err))
	}

	go func() {
		<-ctx.Done()

		if err = conn.Close(); err != nil {
			logging.FromContext(ctx).Debug("Could not close netlink connection.",
				slog.String("worker", netRatesWorkerID),
				slog.Any("error", err))
		}
	}()

	ratesWorker := &netStatsWorker{
		statsSensors: make(map[string]map[netStatsType]*netStatsSensor),
		nlconn:       conn,
	}
	ratesWorker.statsSensors[totalsName] = generateSensors(totalsName, nil)

	ratesWorker.prefs, err = preferences.LoadWorker(ratesWorker)
	if err != nil {
		return nil, errors.Join(ErrInitStatsWorker, err)
	}

	//nolint:nilnil
	if ratesWorker.prefs.IsDisabled() {
		return nil, nil
	}

	pollInterval, err := time.ParseDuration(ratesWorker.prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", netRatesWorkerID),
			slog.String("given_interval", ratesWorker.prefs.UpdateInterval),
			slog.String("default_interval", rateInterval.String()))

		pollInterval = rateInterval
	}

	worker := linux.NewPollingSensorWorker(netRatesWorkerID, pollInterval, rateJitter)
	worker.PollingSensorType = ratesWorker

	return worker, nil
}
