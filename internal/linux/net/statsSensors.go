// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Parts of the code for collecting stats was adapted from Prometheus:
// https://github.com/prometheus/node_exporter//collector/netdev_linux.go

//go:generate go run golang.org/x/tools/cmd/stringer -type=netStatsType -output networkRates_generated.go -linecomment
//revive:disable:unused-receiver
package net

import (
	"context"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/jsimonetti/rtnetlink"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
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

// linkStats represents a link and its stats.
type linkStats struct {
	stats *rtnetlink.LinkStats64
	name  string
}

// netStatsType is the type of stat being tracked by an entity.
type netStatsType int

// Icon returns an material design icon representation of the network stat.
func (t netStatsType) Icon() string {
	switch t {
	case bytesSent:
		return "mdi:upload-network"
	case bytesRecv:
		return "mdi:download-network"
	case bytesSentRate:
		return "mdi:transfer-up"
	case bytesRecvRate:
		return "mdi:transfer-down"
	}

	return ""
}

// newStatsTotalEntity creates an entity for tracking total stats for a network device.
func newStatsTotalEntity(ctx context.Context, name string, entityType netStatsType, category models.EntityCategory, value uint64, attributes models.Attributes) (models.Entity, error) {
	return sensor.NewSensor(ctx,
		sensor.WithName(name+" "+entityType.String()),
		sensor.WithID(strings.ToLower(name)+"_"+strcase.ToSnake(entityType.String())),
		sensor.WithDeviceClass(class.SensorClassDataSize),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits(countUnit),
		sensor.WithIcon(entityType.Icon()),
		sensor.WithState(value),
		sensor.WithAttributes(attributes),
		sensor.WithDataSourceAttribute(linux.DataSrcNetlink),
		sensor.WithCategory(category),
	)
}

// newStatsTotalEntity creates an entity for tracking rate stats for a network device.
func newStatsRateEntity(ctx context.Context, name string, entityType netStatsType, category models.EntityCategory, value uint64) (models.Entity, error) {
	return sensor.NewSensor(ctx,
		sensor.WithName(name+" "+entityType.String()),
		sensor.WithID(strings.ToLower(name)+"_"+strcase.ToSnake(entityType.String())),
		sensor.WithDeviceClass(class.SensorClassDataRate),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits(rateUnit),
		sensor.WithIcon(entityType.Icon()),
		sensor.WithState(value),
		sensor.WithDataSourceAttribute(linux.DataSrcNetlink),
		sensor.WithCategory(category),
	)
}

// statsRate hold data for tracking the rate of change of a stat.
type statsRate struct {
	rateType  netStatsType
	prevValue uint64
}

// update will update the value for a sensor. For count sensors, the value is
// updated directly based on the new stats. For rates sensors, the new rate is
// calculated and the previous value saved.
func (r *statsRate) calculateRate(stats *rtnetlink.LinkStats64, delta time.Duration) uint64 {
	var (
		curr uint64
		prev uint64
	)

	switch r.rateType {
	case bytesRecvRate:
		curr = stats.TXBytes
		prev = r.prevValue
		r.prevValue = stats.RXBytes
	case bytesSentRate:
		curr = stats.RXBytes
		prev = r.prevValue
		r.prevValue = stats.TXBytes
	}

	if uint64(delta.Seconds()) > 0 && prev != 0 {
		return (curr - prev) / uint64(delta.Seconds())
	}

	return 0
}

func newStatsRates() map[netStatsType]*statsRate {
	return map[netStatsType]*statsRate{
		bytesRecvRate: {rateType: bytesRecvRate},
		bytesSentRate: {rateType: bytesSentRate},
	}
}

// getRXAttributes returns all sundry receive stats which can be added to a
// sensor as extra attributes.
func getRXAttributes(stats *rtnetlink.LinkStats64) models.Attributes {
	return models.Attributes{
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
func getTXAttributes(stats *rtnetlink.LinkStats64) models.Attributes {
	return models.Attributes{
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
