// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Parts of the code for collecting stats was adapted from Prometheus:
// https://github.com/prometheus/node_exporter//collector/netdev_linux.go

//go:generate go run golang.org/x/tools/cmd/stringer -type=netStatsType -output networkRates_generated.go -linecomment
//revive:disable:unused-receiver
package net

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/jsimonetti/rtnetlink"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	countUnit = "B"
	rateUnit  = "B/s"

	rateInterval = 5 * time.Second
	rateJitter   = time.Second

	bytesSent     netStatsType = iota // Bytes Sent
	bytesRecv                         // Bytes Received
	bytesSentRate                     // Bytes Sent Throughput
	bytesRecvRate                     // Bytes Received Throughput

	netRatesWorkerID = "network_stats_worker"

	totalsName = "Total"
)

var sensorList = []netStatsType{bytesRecv, bytesSent, bytesRecvRate, bytesSentRate}

// linkStats represents a link and its stats.
type linkStats struct {
	stats *rtnetlink.LinkStats64
	name  string
}

type netStatsType int

type netStatsSensor struct {
	sensor.Entity
	previousValue uint64
}

// newNetStatsSensor creates a new network stats sensor.
func newNetStatsSensor(name string, sensorType netStatsType, stats *rtnetlink.LinkStats64) *netStatsSensor {
	var (
		icon, units string
		deviceClass types.DeviceClass
		stateClass  types.StateClass
	)

	netSensor := &netStatsSensor{}

	// Set type-specific values.
	switch sensorType {
	case bytesRecv:
		icon = "mdi:download-network"
		units = countUnit
		deviceClass = types.SensorDeviceClassDataSize
		stateClass = types.StateClassMeasurement
	case bytesSent:
		icon = "mdi:upload-network"
		units = countUnit
		deviceClass = types.SensorDeviceClassDataSize
		stateClass = types.StateClassMeasurement
	case bytesRecvRate:
		icon = "mdi:transfer-down"
		units = rateUnit
		deviceClass = types.SensorDeviceClassDataRate
		stateClass = types.StateClassMeasurement
	case bytesSentRate:
		icon = "mdi:transfer-up"
		units = rateUnit
		deviceClass = types.SensorDeviceClassDataRate
		stateClass = types.StateClassMeasurement
	}

	netSensor.Entity = sensor.NewSensor(
		sensor.WithName(name+" "+sensorType.String()),
		sensor.WithID(strings.ToLower(name)+"_"+strcase.ToSnake(sensorType.String())),
		sensor.WithDeviceClass(deviceClass),
		sensor.WithStateClass(stateClass),
		sensor.WithUnits(units),
		sensor.WithState(
			sensor.WithIcon(icon),
			sensor.WithDataSourceAttribute(linux.DataSrcNetlink),
		),
	)

	// Set device sensors to category diagnostic.
	if name != totalsName {
		netSensor.Entity = sensor.AsDiagnostic()(netSensor.Entity)
	}

	// Set current value.
	netSensor.update(name, sensorType, stats, 0)

	return netSensor
}

// update will update the value for a sensor. For count sensors, the value is
// updated directly based on the new stats. For rates sensors, the new rate is
// calculated and the previous value saved.
func (s *netStatsSensor) update(link string, sensorType netStatsType, stats *rtnetlink.LinkStats64, delta time.Duration) {
	if stats == nil {
		return
	}

	switch sensorType {
	case bytesRecv:
		s.UpdateValue(stats.RXBytes)

		if link != totalsName {
			maps.Copy(s.Attributes, getRXAttributes(stats))
		}
	case bytesSent:
		s.UpdateValue(stats.TXBytes)

		if link != totalsName {
			maps.Copy(s.Attributes, getTXAttributes(stats))
		}
	case bytesRecvRate:
		rate := calculateRate(stats.RXBytes, s.previousValue, delta)
		s.UpdateValue(rate)
		s.previousValue = stats.RXBytes
	case bytesSentRate:
		rate := calculateRate(stats.TXBytes, s.previousValue, delta)
		s.UpdateValue(rate)
		s.previousValue = stats.TXBytes
	}
}

// calculate rate calculates a sensor value as a rate based on the
// current/previous values and a delta (time since last measurement).
func calculateRate(currentValue, previousValue uint64, delta time.Duration) uint64 {
	if uint64(delta.Seconds()) > 0 && previousValue != 0 {
		return (currentValue - previousValue) / uint64(delta.Seconds())
	}

	return 0
}

// getRXAttributes returns all sundry receive stats which can be added to a
// sensor as extra attributes.
func getRXAttributes(stats *rtnetlink.LinkStats64) map[string]any {
	return map[string]any{
		"receive_packets":       stats.RXPackets,
		"receive_errors":        stats.RXErrors,
		"receive_dropped":       stats.RXDropped,
		"multicast":             stats.Multicast,
		"collisions":            stats.Collisions,
		"receive_length_errors": stats.RXLengthErrors,
		"receive_over_errors":   stats.RXOverErrors,
		"receive_crc_errors":    stats.RXCRCErrors,
		"receive_frame_errors":  stats.RXFrameErrors,
		"receive_fifo_errors":   stats.RXFIFOErrors,
		"receive_missed_errors": stats.RXMissedErrors,
		"receive_compressed":    stats.RXCompressed,
		"transmit_compressed":   stats.TXCompressed,
		"receive_nohandler":     stats.RXNoHandler,
	}
}

// getTXAttributes returns all sundry transmit stats which can be added to a
// sensor as extra attributes.
func getTXAttributes(stats *rtnetlink.LinkStats64) map[string]any {
	return map[string]any{
		"transmit_packets":          stats.TXPackets,
		"transmit_errors":           stats.TXErrors,
		"transmit_dropped":          stats.TXDropped,
		"multicast":                 stats.Multicast,
		"collisions":                stats.Collisions,
		"transmit_aborted_errors":   stats.TXAbortedErrors,
		"transmit_carrier_errors":   stats.TXCarrierErrors,
		"transmit_fifo_errors":      stats.TXFIFOErrors,
		"transmit_heartbeat_errors": stats.TXHeartbeatErrors,
		"transmit_window_errors":    stats.TXWindowErrors,
		"transmit_compressed":       stats.TXCompressed,
	}
}

// netStatsWorker is the object used for tracking network stats sensors. It
// holds a netlink connection and a map of links with their stats sensors.
type netStatsWorker struct {
	statsSensors map[string]map[netStatsType]*netStatsSensor
	nlconn       *rtnetlink.Conn
	prefs        *WorkerPrefs
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
	return preferencesID
}

func (w *netStatsWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{
		IgnoredDevices: defaultIgnoredDevices,
	}
}

// NewNetStatsWorker sets up a sensor worker that tracks network stats.
func NewNetStatsWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingSensorWorker(netRatesWorkerID, rateInterval, rateJitter)

	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		return worker, fmt.Errorf("could not connect to netlink: %w", err)
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

	ratesWorker.prefs, err = preferences.LoadWorker(ctx, ratesWorker)
	if err != nil {
		return worker, fmt.Errorf("could not load preferences: %w", err)
	}

	if ratesWorker.prefs.Disabled {
		return worker, nil
	}

	worker.PollingSensorType = ratesWorker

	return worker, nil
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

// generateSensors creates a map of sensors for the given link.
func generateSensors(name string, stats *rtnetlink.LinkStats64) map[netStatsType]*netStatsSensor {
	sensors := make(map[netStatsType]*netStatsSensor, 4)
	for _, sensorType := range sensorList {
		sensors[sensorType] = newNetStatsSensor(name, sensorType, stats)
	}

	return sensors
}
