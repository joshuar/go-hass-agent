// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/net"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	countUnit = "B"
	rateUnit  = "B/s"
)

type netIOSensorAttributes struct {
	Packets    uint64 `json:"Packets"`     // number of packets
	Errors     uint64 `json:"Errors"`      // total number of errors
	Drops      uint64 `json:"Drops"`       // total number of packets which were dropped
	FifoErrors uint64 `json:"Fifo Errors"` // total number of FIFO buffers errors
}

type netIOSensor struct {
	linux.Sensor
	netIOSensorAttributes
}

func (s *netIOSensor) Attributes() any {
	return struct {
		NativeUnit string `json:"native_unit_of_measurement"`
		DataSource string `json:"Data Source"`
		netIOSensorAttributes
	}{
		NativeUnit:            s.UnitsString,
		DataSource:            linux.DataSrcProcfs,
		netIOSensorAttributes: s.netIOSensorAttributes,
	}
}

func (s *netIOSensor) Icon() string {
	switch s.SensorTypeValue {
	case linux.SensorBytesRecv:
		return "mdi:download-network"
	case linux.SensorBytesSent:
		return "mdi:upload-network"
	}
	return "mdi:help-network"
}

func (s *netIOSensor) update(c *net.IOCountersStat) {
	switch s.SensorTypeValue {
	case linux.SensorBytesRecv:
		s.Value = c.BytesRecv
		s.Packets = c.PacketsRecv
		s.Errors = c.Errin
		s.Drops = c.Dropin
		s.FifoErrors = c.Fifoin
	case linux.SensorBytesSent:
		s.Value = c.BytesSent
		s.Packets = c.PacketsSent
		s.Errors = c.Errout
		s.Drops = c.Dropout
		s.FifoErrors = c.Fifoout
	}
}

func newNetIOSensor(t linux.SensorTypeValue) *netIOSensor {
	return &netIOSensor{
		Sensor: linux.Sensor{
			UnitsString:      countUnit,
			SensorTypeValue:  t,
			DeviceClassValue: types.DeviceClassDataSize,
			StateClassValue:  types.StateClassMeasurement,
		},
	}
}

type netIORateSensor struct {
	linux.Sensor
	lastValue uint64
}

func (s *netIORateSensor) Icon() string {
	switch s.SensorTypeValue {
	case linux.SensorBytesRecvRate:
		return "mdi:transfer-down"
	case linux.SensorBytesSentRate:
		return "mdi:transfer-up"
	}
	return "mdi:help-network"
}

func (s *netIORateSensor) update(d time.Duration, b uint64) {
	if uint64(d.Seconds()) > 0 && s.lastValue != 0 {
		s.Value = (b - s.lastValue) / uint64(d.Seconds())
	}
	s.lastValue = b
}

func newNetIORateSensor(t linux.SensorTypeValue) *netIORateSensor {
	return &netIORateSensor{
		Sensor: linux.Sensor{
			UnitsString:      rateUnit,
			SensorTypeValue:  t,
			DeviceClassValue: types.DeviceClassDataRate,
			StateClassValue:  types.StateClassMeasurement,
			SensorSrc:        linux.DataSrcProcfs,
		},
	}
}

type ratesWorker struct {
	bytesRx, bytesTx         *netIOSensor
	bytesRxRate, bytesTxRate *netIORateSensor
}

func (w *ratesWorker) Interval() time.Duration { return 5 * time.Second }

func (w *ratesWorker) Jitter() time.Duration { return time.Second }

func (w *ratesWorker) Sensors(ctx context.Context, d time.Duration) ([]sensor.Details, error) {
	// Retrieve new stats.
	netIO, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("problem fetching network stats: %w", err)
	}
	// Update all sensors.
	w.bytesRx.update(&netIO[0])
	w.bytesTx.update(&netIO[0])
	w.bytesRxRate.update(d, netIO[0].BytesRecv)
	w.bytesTxRate.update(d, netIO[0].BytesSent)
	// Return sensors with new values.
	return []sensor.Details{w.bytesRx, w.bytesTx, w.bytesRxRate, w.bytesTxRate}, nil
}

func NewRatesWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "Network Rates Sensors",
			WorkerDesc: "Network transfer amount and speed sensors.",
			Value: &ratesWorker{
				bytesRx:     newNetIOSensor(linux.SensorBytesRecv),
				bytesTx:     newNetIOSensor(linux.SensorBytesSent),
				bytesRxRate: newNetIORateSensor(linux.SensorBytesRecvRate),
				bytesTxRate: newNetIORateSensor(linux.SensorBytesSentRate),
			},
		},
		nil
}
