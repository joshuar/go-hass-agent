// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Parts of the code for collecting stats was adapted from Prometheus:
// https://github.com/prometheus/node_exporter//collector/netdev_linux.go

//go:generate go tool golang.org/x/tools/cmd/stringer -type=netStatsType -output networkRates_generated.go -linecomment
//revive:disable:unused-receiver
package net

import (
	"context"
	"errors"
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

var (
	ErrNewStatsSensor = errors.New("could not create stats sensor")
	ErrNewRatesSensor = errors.New("could not create rates sensor")
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

// newStatsTotalEntity creates an entity for tracking total stats for a network
// device.
//
//revive:disable:argument-limit
func newStatsTotalEntity(ctx context.Context, name string, entityType netStatsType, value uint64, attributes models.Attributes) (*models.Entity, error) {
	statsSensor, err := sensor.NewSensor(ctx,
		sensor.WithName(name+" "+entityType.String()),
		sensor.WithID(strings.ToLower(name)+"_"+strcase.ToSnake(entityType.String())),
		sensor.WithDeviceClass(class.SensorClassDataSize),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits(countUnit),
		sensor.WithIcon(entityType.Icon()),
		sensor.WithState(value),
		sensor.WithAttributes(attributes),
		sensor.WithDataSourceAttribute(linux.DataSrcNetlink),
		sensor.AsDiagnostic(),
	)
	if err != nil {
		return nil, errors.Join(ErrNewStatsSensor, err)
	}

	return &statsSensor, nil
}

// newStatsTotalEntity creates an entity for tracking rate stats for a network device.
func newStatsRateEntity(ctx context.Context, name string, entityType netStatsType, value uint64) (*models.Entity, error) {
	ratesSensor, err := sensor.NewSensor(ctx,
		sensor.WithName(name+" "+entityType.String()),
		sensor.WithID(strings.ToLower(name)+"_"+strcase.ToSnake(entityType.String())),
		sensor.WithDeviceClass(class.SensorClassDataRate),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits(rateUnit),
		sensor.WithIcon(entityType.Icon()),
		sensor.WithState(value),
		sensor.WithDataSourceAttribute(linux.DataSrcNetlink),
		sensor.AsDiagnostic(),
	)
	if err != nil {
		return nil, errors.Join(ErrNewRatesSensor, err)
	}

	return &ratesSensor, nil
}

type netRate struct {
	linux.RateValue[uint64]
	rateType netStatsType
}

func newStatsRates() map[netStatsType]*netRate {
	return map[netStatsType]*netRate{
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
