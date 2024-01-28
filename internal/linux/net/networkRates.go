// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/net"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
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
			UnitsString:      "B",
			SensorTypeValue:  t,
			DeviceClassValue: sensor.Data_size,
			StateClassValue:  sensor.StateMeasurement,
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
			UnitsString:      "B/s",
			SensorTypeValue:  t,
			DeviceClassValue: sensor.Data_rate,
			StateClassValue:  sensor.StateMeasurement,
			SensorSrc:        linux.DataSrcProcfs,
		},
	}
}

func RatesUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 2)
	bytesRx := newNetIOSensor(linux.SensorBytesRecv)
	bytesTx := newNetIOSensor(linux.SensorBytesSent)
	bytesRxRate := newNetIORateSensor(linux.SensorBytesRecvRate)
	bytesTxRate := newNetIORateSensor(linux.SensorBytesSentRate)

	sendNetStats := func(delta time.Duration) {
		netIO, err := net.IOCountersWithContext(ctx, false)
		if err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching network stats.")
			return
		}

		bytesRx.update(&netIO[0])
		sensorCh <- bytesRx
		bytesTx.update(&netIO[0])
		sensorCh <- bytesTx

		bytesRxRate.update(delta, netIO[0].BytesRecv)
		sensorCh <- bytesRxRate
		bytesTxRate.update(delta, netIO[0].BytesSent)
		sensorCh <- bytesTxRate
	}

	go helpers.PollSensors(ctx, sendNetStats, 5*time.Second, time.Second*1)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped network stats sensors.")
	}()
	return sensorCh
}
