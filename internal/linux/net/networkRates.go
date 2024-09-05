// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate stringer -type=rateSensor -output networkRates_generated.go -linecomment
//revive:disable:unused-receiver
package net

import (
	"context"
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/shirou/gopsutil/v3/net"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	countUnit = "B"
	rateUnit  = "B/s"

	rateInterval = 5 * time.Second
	rateJitter   = time.Second

	bytesSent     rateSensor = iota // Bytes Sent
	bytesRecv                       // Bytes Received
	bytesSentRate                   // Bytes Sent Throughput
	bytesRecvRate                   // Bytes Received Throughput

	netRatesWorkerID = "network_rates_sensors"
)

type rateSensor int

type netIOSensorAttributes struct {
	Packets    uint64 `json:"packets"`     // number of packets
	Errors     uint64 `json:"errors"`      // total number of errors
	Drops      uint64 `json:"drops"`       // total number of packets which were dropped
	FifoErrors uint64 `json:"fifo_errors"` // total number of FIFO buffers errors
}

type netIOSensor struct {
	linux.Sensor
	sensorType rateSensor
	netIOSensorAttributes
}

func (s *netIOSensor) Attributes() map[string]any {
	attributes := s.Sensor.Attributes()
	attributes["native_unit_of_measurement"] = s.UnitsString
	attributes["stats"] = s.netIOSensorAttributes

	return attributes
}

//nolint:exhaustive
func (s *netIOSensor) Icon() string {
	switch s.sensorType {
	case bytesRecv:
		return "mdi:download-network"
	case bytesSent:
		return "mdi:upload-network"
	}

	return "mdi:help-network"
}

//nolint:exhaustive
func (s *netIOSensor) update(counters *net.IOCountersStat) {
	switch s.sensorType {
	case bytesRecv:
		s.Value = counters.BytesRecv
		s.Packets = counters.PacketsRecv
		s.Errors = counters.Errin
		s.Drops = counters.Dropin
		s.FifoErrors = counters.Fifoin
	case bytesSent:
		s.Value = counters.BytesSent
		s.Packets = counters.PacketsSent
		s.Errors = counters.Errout
		s.Drops = counters.Dropout
		s.FifoErrors = counters.Fifoout
	}
}

func newNetIOSensor(t rateSensor) *netIOSensor {
	return &netIOSensor{
		sensorType: t,
		Sensor: linux.Sensor{
			DisplayName:      t.String(),
			UniqueID:         strcase.ToSnake(t.String()),
			UnitsString:      countUnit,
			DeviceClassValue: types.DeviceClassDataSize,
			StateClassValue:  types.StateClassMeasurement,
			DataSource:       linux.DataSrcProcfs,
		},
	}
}

type netIORateSensor struct {
	linux.Sensor
	sensorType rateSensor
	lastValue  uint64
}

//nolint:exhaustive
func (s *netIORateSensor) Icon() string {
	switch s.sensorType {
	case bytesRecvRate:
		return "mdi:transfer-down"
	case bytesSentRate:
		return "mdi:transfer-up"
	}

	return "mdi:help-network"
}

func (s *netIORateSensor) update(d time.Duration, currentValue uint64) {
	if uint64(d.Seconds()) > 0 && s.lastValue != 0 {
		s.Value = (currentValue - s.lastValue) / uint64(d.Seconds())
	} else {
		s.Value = 0
	}

	s.lastValue = currentValue
}

func newNetIORateSensor(t rateSensor) *netIORateSensor {
	return &netIORateSensor{
		sensorType: t,
		Sensor: linux.Sensor{
			DisplayName:      t.String(),
			UniqueID:         strcase.ToSnake(t.String()),
			UnitsString:      rateUnit,
			DeviceClassValue: types.DeviceClassDataRate,
			StateClassValue:  types.StateClassMeasurement,
			DataSource:       linux.DataSrcProcfs,
		},
	}
}

type ratesWorker struct {
	bytesRx, bytesTx         *netIOSensor
	bytesRxRate, bytesTxRate *netIORateSensor
}

func (w *ratesWorker) Interval() time.Duration { return rateInterval }

func (w *ratesWorker) Jitter() time.Duration { return rateJitter }

func (w *ratesWorker) Sensors(ctx context.Context, duration time.Duration) ([]sensor.Details, error) {
	// Retrieve new stats.
	netIO, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("problem fetching network stats: %w", err)
	}
	// Update all sensors.
	w.bytesRx.update(&netIO[0])
	w.bytesTx.update(&netIO[0])
	w.bytesRxRate.update(duration, netIO[0].BytesRecv)
	w.bytesTxRate.update(duration, netIO[0].BytesSent)
	// Return sensors with new values.
	return []sensor.Details{w.bytesRx, w.bytesTx, w.bytesRxRate, w.bytesTxRate}, nil
}

func NewRatesWorker(_ context.Context) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &ratesWorker{
				bytesRx:     newNetIOSensor(bytesRecv),
				bytesTx:     newNetIOSensor(bytesSent),
				bytesRxRate: newNetIORateSensor(bytesRecvRate),
				bytesTxRate: newNetIORateSensor(bytesSentRate),
			},
			WorkerID: netRatesWorkerID,
		},
		nil
}
