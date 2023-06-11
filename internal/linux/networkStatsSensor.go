// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/net"
)

//go:generate stringer -type=networkStat -output networkStatsSensorProps.go

const (
	bytesSent networkStat = iota + 1
	bytesRecv
)

type networkStat int

type statAttributes struct {
	Packets    uint64 `json:"Packets"`     // number of packets
	Errors     uint64 `json:"Errors"`      // total number of errors
	Drops      uint64 `json:"Drops"`       // total number of packets which were dropped
	FifoErrors uint64 `json:"Fifo Errors"` // total number of FIFO buffers errors
}

type networkStatsDetails struct {
	statType  networkStat
	statValue uint64
	statAttributes
}

func (i *networkStatsDetails) Name() string {
	return strcase.ToDelimited(i.statType.String(), ' ')
}

func (i *networkStatsDetails) ID() string {
	return strcase.ToSnake(i.statType.String())
}

func (i *networkStatsDetails) Icon() string {
	switch i.statType {
	case bytesRecv:
		return "mdi:download-network"
	case bytesSent:
		return "mdi:upload-network"
	default:
		return "mdi:help-network"
	}
}

func (i *networkStatsDetails) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (i *networkStatsDetails) DeviceClass() hass.SensorDeviceClass {
	return hass.Data_size
}

func (i *networkStatsDetails) StateClass() hass.SensorStateClass {
	return hass.StateTotal
}

func (i *networkStatsDetails) State() interface{} {
	return i.statValue
}

func (i *networkStatsDetails) Units() string {
	return "B"
}

func (i *networkStatsDetails) Category() string {
	return ""
}

func (i *networkStatsDetails) Attributes() interface{} {
	return i.statAttributes
}

func NetworkStatsUpdater(ctx context.Context, status chan interface{}) {

	sendNetStats := func() {
		statTypes := []networkStat{bytesRecv, bytesSent}
		var allInterfaces []net.IOCountersStat
		var err error
		if allInterfaces, err = net.IOCountersWithContext(ctx, false); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching network stats.")
			return
		}
		for _, interfaceStats := range allInterfaces {
			for _, stat := range statTypes {
				details := &networkStatsDetails{}
				details.statType = stat
				switch stat {
				case bytesRecv:
					details.statValue = interfaceStats.BytesRecv
					details.statAttributes.Packets = interfaceStats.PacketsRecv
					details.statAttributes.Errors = interfaceStats.Errin
					details.statAttributes.Drops = interfaceStats.Dropin
					details.statAttributes.FifoErrors = interfaceStats.Fifoin
				case bytesSent:
					details.statValue = interfaceStats.BytesSent
					details.statAttributes.Packets = interfaceStats.PacketsSent
					details.statAttributes.Errors = interfaceStats.Errout
					details.statAttributes.Drops = interfaceStats.Dropout
					details.statAttributes.FifoErrors = interfaceStats.Fifoout
				}
				status <- details
			}
		}
	}

	pollSensors(ctx, sendNetStats, time.Minute, time.Second*5)
}
