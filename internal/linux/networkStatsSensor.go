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

type netIOStats struct {
	Packets    uint64 `json:"Packets"`     // number of packets
	Errors     uint64 `json:"Errors"`      // total number of errors
	Drops      uint64 `json:"Drops"`       // total number of packets which were dropped
	FifoErrors uint64 `json:"Fifo Errors"` // total number of FIFO buffers errors
}

type netIOSensor struct {
	linuxSensor
	netIOStats
}

func (s *netIOSensor) Attributes() interface{} {
	return struct {
		NativeUnit string `json:"native_unit_of_measurement"`
		DataSource string `json:"Data Source"`
		netIOStats
	}{
		NativeUnit: s.units,
		DataSource: srcProcfs,
		netIOStats: s.netIOStats,
	}
}

type netIORateSensor struct {
	linuxSensor
	lastValue uint64
}

func NetworkStatsUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor)
	bytesRx := &netIOSensor{
		linuxSensor: linuxSensor{
			units:       "B",
			icon:        "mdi:download-network",
			sensorType:  bytesRecv,
			deviceClass: sensor.Data_size,
			stateClass:  sensor.StateMeasurement,
		},
	}
	bytesTx := &netIOSensor{
		linuxSensor: linuxSensor{
			units:       "B",
			icon:        "mdi:upload-network",
			sensorType:  bytesSent,
			deviceClass: sensor.Data_size,
			stateClass:  sensor.StateMeasurement,
		},
	}
	bytesRxRate := &netIORateSensor{
		linuxSensor: linuxSensor{
			units:       "B/s",
			icon:        "mdi:transfer-down",
			sensorType:  bytesRecvRate,
			deviceClass: sensor.Data_rate,
			stateClass:  sensor.StateMeasurement,
			source:      srcProcfs,
		},
	}
	bytesTxRate := &netIORateSensor{
		linuxSensor: linuxSensor{
			units:       "B/s",
			icon:        "mdi:transfer-up",
			sensorType:  bytesSentRate,
			deviceClass: sensor.Data_rate,
			stateClass:  sensor.StateMeasurement,
			source:      srcProcfs,
		},
	}

	sendNetStats := func(delta time.Duration) {
		netIO, err := net.IOCountersWithContext(ctx, false)
		if err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching network stats.")
			return
		}
		bytesRx.value = netIO[0].BytesRecv
		bytesRx.Packets = netIO[0].PacketsRecv
		bytesRx.Errors = netIO[0].Errin
		bytesRx.Drops = netIO[0].Dropin
		bytesRx.FifoErrors = netIO[0].Fifoin
		sensorCh <- bytesRx

		bytesTx.value = netIO[0].BytesSent
		bytesTx.Packets = netIO[0].PacketsSent
		bytesTx.Errors = netIO[0].Errout
		bytesTx.Drops = netIO[0].Dropout
		bytesTx.FifoErrors = netIO[0].Fifoout
		sensorCh <- bytesTx

		if uint64(delta.Seconds()) > 0 && bytesRxRate.lastValue != 0 {
			bytesRxRate.value = (netIO[0].BytesRecv - bytesRxRate.lastValue) / uint64(delta.Seconds())
			bytesTxRate.value = (netIO[0].BytesSent - bytesTxRate.lastValue) / uint64(delta.Seconds())
			sensorCh <- bytesRxRate
			sensorCh <- bytesTxRate
		}
		bytesRxRate.lastValue = netIO[0].BytesRecv
		bytesTxRate.lastValue = netIO[0].BytesSent
	}

	go helpers.PollSensors(ctx, sendNetStats, 5*time.Second, time.Second*1)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
	}()
	return sensorCh
}
