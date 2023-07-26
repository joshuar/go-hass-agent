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
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/net"
)

type networkStatsAttributes struct {
	Packets    uint64 `json:"Packets"`     // number of packets
	Errors     uint64 `json:"Errors"`      // total number of errors
	Drops      uint64 `json:"Drops"`       // total number of packets which were dropped
	FifoErrors uint64 `json:"Fifo Errors"` // total number of FIFO buffers errors
}

type networkStatsSensor struct {
	linuxSensor
	networkStatsAttributes
}

func (s *networkStatsSensor) Attributes() interface{} {
	return struct {
		NativeUnit string `json:"native_unit_of_measurement"`
		DataSource string `json:"Data Source"`
		networkStatsAttributes
	}{
		NativeUnit:             s.units,
		DataSource:             "procfs",
		networkStatsAttributes: s.networkStatsAttributes,
	}
}

func NetworkStatsUpdater(ctx context.Context, status chan interface{}) {

	sendNetStats := func() {
		statTypes := []sensorType{bytesRecv, bytesSent}
		var allInterfaces []net.IOCountersStat
		var err error
		if allInterfaces, err = net.IOCountersWithContext(ctx, false); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching network stats.")
			return
		}
		for _, interfaceStats := range allInterfaces {
			for _, stat := range statTypes {
				s := &networkStatsSensor{}
				s.sensorType = stat
				s.units = "B"
				s.deviceClass = sensor.Data_size
				s.stateClass = sensor.StateTotal
				switch stat {
				case bytesRecv:
					s.value = interfaceStats.BytesRecv
					s.icon = "mdi:download-network"
					s.Packets = interfaceStats.PacketsRecv
					s.Errors = interfaceStats.Errin
					s.Drops = interfaceStats.Dropin
					s.FifoErrors = interfaceStats.Fifoin
				case bytesSent:
					s.value = interfaceStats.BytesSent
					s.icon = "mdi:upload-network"
					s.Packets = interfaceStats.PacketsSent
					s.Errors = interfaceStats.Errout
					s.Drops = interfaceStats.Dropout
					s.FifoErrors = interfaceStats.Fifoout
				}
				status <- s
			}
		}
	}

	helpers.PollSensors(ctx, sendNetStats, time.Minute, time.Second*5)
}
