// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"time"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/net"
)

type netIOSensorAttributes struct {
	Packets    uint64 `json:"Packets"`     // number of packets
	Errors     uint64 `json:"Errors"`      // total number of errors
	Drops      uint64 `json:"Drops"`       // total number of packets which were dropped
	FifoErrors uint64 `json:"Fifo Errors"` // total number of FIFO buffers errors
}

type netIOSensor struct {
	linuxSensor
	netIOSensorAttributes
}

func (s *netIOSensor) Attributes() any {
	return struct {
		NativeUnit string `json:"native_unit_of_measurement"`
		DataSource string `json:"Data Source"`
		netIOSensorAttributes
	}{
		NativeUnit:            s.units,
		DataSource:            srcProcfs,
		netIOSensorAttributes: s.netIOSensorAttributes,
	}
}

func (s *netIOSensor) Icon() string {
	switch s.sensorType {
	case bytesRecv:
		return "mdi:download-network"
	case bytesSent:
		return "mdi:upload-network"
	}
	return "mdi:help-network"
}

func (s *netIOSensor) update(c *net.IOCountersStat) {
	switch s.sensorType {
	case bytesRecv:
		s.value = c.BytesRecv
		s.Packets = c.PacketsRecv
		s.Errors = c.Errin
		s.Drops = c.Dropin
		s.FifoErrors = c.Fifoin
	case bytesSent:
		s.value = c.BytesSent
		s.Packets = c.PacketsSent
		s.Errors = c.Errout
		s.Drops = c.Dropout
		s.FifoErrors = c.Fifoout
	}
}

func newNetIOSensor(t sensorType) *netIOSensor {
	return &netIOSensor{
		linuxSensor: linuxSensor{
			units:       "B",
			sensorType:  t,
			deviceClass: sensor.Data_size,
			stateClass:  sensor.StateMeasurement,
		},
	}
}

type netIORateSensor struct {
	linuxSensor
	lastValue uint64
}

func (s *netIORateSensor) Icon() string {
	switch s.sensorType {
	case bytesRecvRate:
		return "mdi:transfer-down"
	case bytesSentRate:
		return "mdi:transfer-up"
	}
	return "mdi:help-network"
}

func (s *netIORateSensor) update(d time.Duration, b uint64) {
	if uint64(d.Seconds()) > 0 && s.lastValue != 0 {
		s.value = (b - s.lastValue) / uint64(d.Seconds())
	}
	s.lastValue = b
}

func newNetIORateSensor(t sensorType) *netIORateSensor {
	return &netIORateSensor{
		linuxSensor: linuxSensor{
			units:       "B/s",
			sensorType:  t,
			deviceClass: sensor.Data_rate,
			stateClass:  sensor.StateMeasurement,
			source:      srcProcfs,
		},
	}
}

func NetworkStatsUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 2)
	bytesRx := newNetIOSensor(bytesRecv)
	bytesTx := newNetIOSensor(bytesSent)
	bytesRxRate := newNetIORateSensor(bytesRecvRate)
	bytesTxRate := newNetIORateSensor(bytesSentRate)

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
