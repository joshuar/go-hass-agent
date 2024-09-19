// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// Parts of the code for collecting stats was adapted from Prometheus:
// https://github.com/prometheus/node_exporter//collector/netdev_linux.go

//go:generate stringer -type=rateSensor -output networkRates_generated.go -linecomment
//revive:disable:unused-receiver
package net

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/jsimonetti/rtnetlink"
	"golang.org/x/exp/maps"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
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

var (
	sensorList     = []netStatsType{bytesRecv, bytesSent, bytesRecvRate, bytesSentRate}
	ignoredDevices = []string{"lo"}
)

// linkStats represents a link and its stats.
type linkStats struct {
	stats *rtnetlink.LinkStats64
	name  string
}

type netStatsType int

type netStatsSensor struct {
	*linux.Sensor
	attributes    map[string]any
	link          string
	sensorType    netStatsType
	previousValue uint64
}

func (s *netStatsSensor) Attributes() map[string]any {
	if s.attributes != nil {
		maps.Copy(s.attributes, s.Sensor.Attributes())
		return s.attributes
	}

	return s.Sensor.Attributes()
}

// newNetStatsSensor creates a new network stats sensor.
func newNetStatsSensor(name string, sensorType netStatsType, stats *rtnetlink.LinkStats64) *netStatsSensor {
	netSensor := &netStatsSensor{
		sensorType: sensorType,
		link:       name,
		Sensor: &linux.Sensor{
			DisplayName: name + " " + sensorType.String(),
			UniqueID:    strings.ToLower(name) + "_" + strcase.ToSnake(sensorType.String()),
			DataSource:  linux.DataSrcProcfs,
		},
	}

	// Set device sensors to category diagnostic.
	if name != totalsName {
		netSensor.IsDiagnostic = true
	}

	// Set type-specific values.
	switch sensorType {
	case bytesRecv:
		netSensor.IconString = "mdi:download-network"
		netSensor.UnitsString = countUnit
		netSensor.DeviceClassValue = types.DeviceClassDataSize
		netSensor.StateClassValue = types.StateClassMeasurement
	case bytesSent:
		netSensor.IconString = "mdi:upload-network"
		netSensor.UnitsString = countUnit
		netSensor.DeviceClassValue = types.DeviceClassDataSize
		netSensor.StateClassValue = types.StateClassMeasurement
	case bytesRecvRate:
		netSensor.IconString = "mdi:transfer-down"
		netSensor.UnitsString = rateUnit
		netSensor.DeviceClassValue = types.DeviceClassDataRate
		netSensor.StateClassValue = types.StateClassMeasurement
	case bytesSentRate:
		netSensor.IconString = "mdi:transfer-up"
		netSensor.UnitsString = rateUnit
		netSensor.DeviceClassValue = types.DeviceClassDataRate
		netSensor.StateClassValue = types.StateClassMeasurement
	}

	// Set current value.
	netSensor.update(stats, 0)

	return netSensor
}

// update will update the value for a sensor. For count sensors, the value is
// updated directly based on the new stats. For rates sensors, the new rate is
// calculated and the previous value saved.
func (s *netStatsSensor) update(stats *rtnetlink.LinkStats64, delta time.Duration) {
	if stats == nil {
		return
	}

	switch s.sensorType {
	case bytesRecv:
		s.Value = stats.RXBytes
		if s.link != totalsName {
			s.attributes = getRXAttributes(stats)
		}
	case bytesSent:
		s.Value = stats.TXBytes
		if s.link != totalsName {
			s.attributes = getTXAttributes(stats)
		}
	case bytesRecvRate:
		rate := calculateRate(stats.RXBytes, s.previousValue, delta)
		s.Value = rate
		s.previousValue = stats.RXBytes
	case bytesSentRate:
		rate := calculateRate(stats.TXBytes, s.previousValue, delta)
		s.Value = rate
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
		w.statsSensors[totalsName][sensorType].update(stats, w.delta)
	}
}

func (w *netStatsWorker) UpdateDelta(delta time.Duration) {
	w.delta = delta
}

func (w *netStatsWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	// Get all links.
	links, err := w.getLinks()
	if err != nil {
		return nil, fmt.Errorf("error accessing netlink: %w", err)
	}

	// Get link stats, filtering "uninteresting" links.
	stats := w.getLinkStats(links)
	sensors := make([]sensor.Details, 0, len(stats)*4+4)

	w.mu.Lock()

	// Counters for totals.
	var totalBytesRx, totalBytesTx uint64

	// For each link, update the link sensors with the new stats.
	for _, link := range stats {
		name := link.name
		stats := link.stats
		totalBytesRx += stats.RXBytes
		totalBytesTx += stats.RXBytes

		if _, ok := w.statsSensors[name]; ok { // Existing link/sensors, update.
			for sensorType := range w.statsSensors[name] {
				w.statsSensors[name][sensorType].update(stats, w.delta)
			}
		} else { // New link, add to tracking map.
			w.statsSensors[name] = generateSensors(name, stats)
		}
		// Create a list of sensors.
		for _, s := range w.statsSensors[name] {
			sensors = append(sensors, s)
		}
	}

	// Update the totals sensors based on the counters.
	w.updateTotals(totalBytesRx, totalBytesTx)
	// Append the total sensors to the list of sensors.
	for _, s := range w.statsSensors[totalsName] {
		sensors = append(sensors, s)
	}

	w.mu.Unlock()

	return sensors, nil
}

// NewNetStatsWorker sets up a sensor worker that tracks network stats.
func NewNetStatsWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingWorker(netRatesWorkerID, rateInterval, rateJitter)

	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		return worker, fmt.Errorf("could not connect to netlink: %w", err)
	}

	go func() {
		<-ctx.Done()

		if err := conn.Close(); err != nil {
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

	worker.PollingType = ratesWorker

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

		// Skip ignored devices.
		if slices.Contains(ignoredDevices, msg.Attributes.Name) {
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
