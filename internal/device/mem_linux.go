// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/mem"
)

//go:generate stringer -type=memoryStat -output mem_stats_linux.go

const (
	MemoryTotal memoryStat = iota + 1
	MemoryAvailable
	MemoryUsed
	SwapMemoryTotal
	SwapMemoryUsed
	SwapMemoryFree
)

type memoryStat int

type memory struct {
	value uint64
	name  memoryStat
}

func (m *memory) Name() string {
	return strcase.ToDelimited(m.name.String(), ' ')
}

func (m *memory) ID() string {
	return strings.ToLower(strcase.ToSnake(m.name.String()))
}

func (m *memory) Icon() string {
	return "mdi:memory"
}

func (m *memory) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (m *memory) DeviceClass() hass.SensorDeviceClass {
	return hass.Data_size
}

func (m *memory) StateClass() hass.SensorStateClass {
	return hass.StateTotal
}

func (m *memory) State() interface{} {
	return m.value
}

func (m *memory) Units() string {
	return "B"
}

func (m *memory) Category() string {
	return ""
}

func (m *memory) Attributes() interface{} {
	return nil
}

func MemoryUpdater(ctx context.Context, status chan interface{}) {
	sendMemStats(status)
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sendMemStats(status)
			}
		}
	}()
}

func sendMemStats(status chan interface{}) {
	stats := []memoryStat{MemoryTotal, MemoryAvailable, MemoryUsed, SwapMemoryTotal, SwapMemoryFree}
	var memDetails *mem.VirtualMemoryStat
	var err error
	if memDetails, err = mem.VirtualMemory(); err != nil {
		log.Debug().Err(err).Caller().
			Msg("Problem fetching memory stats.")
		return
	}
	for _, stat := range stats {
		var statValue uint64
		switch stat {
		case MemoryTotal:
			statValue = memDetails.Total
		case MemoryAvailable:
			statValue = memDetails.Available
		case MemoryUsed:
			statValue = memDetails.Used
		case SwapMemoryTotal:
			statValue = memDetails.SwapTotal
		case SwapMemoryFree:
			statValue = memDetails.SwapFree
			// case UsedSwapMemory:
			// 	return m.memStats.SwapCached
		}
		state := &memory{
			value: statValue,
			name:  stat,
		}
		status <- state
	}
}
