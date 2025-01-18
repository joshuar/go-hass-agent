// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Parts of the code for collecting stats was adapted from Prometheus:
// https://github.com/prometheus/node_exporter//collector/netdev_linux.go

//go:generate go run golang.org/x/tools/cmd/stringer -type=netStatsType -output networkRates_generated.go -linecomment
//revive:disable:unused-receiver
package net

import (
	"maps"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/jsimonetti/rtnetlink"

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

// generateSensors creates a map of sensors for the given link.
func generateSensors(name string, stats *rtnetlink.LinkStats64) map[netStatsType]*netStatsSensor {
	sensors := make(map[netStatsType]*netStatsSensor, 4)
	for _, sensorType := range sensorList {
		sensors[sensorType] = newNetStatsSensor(name, sensorType, stats)
	}

	return sensors
}
