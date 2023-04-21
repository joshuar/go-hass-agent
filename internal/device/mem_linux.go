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
	TotalMemory memoryStat = iota + 1
	AvailableMemory
	UsedMemory
	TotalSwapMemory
	UsedSwapMemory
	FreeSwapMemory
)

type memoryStat int

type memory struct {
	memStats *mem.VirtualMemoryStat
	name     memoryStat
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
	switch m.name {
	case TotalMemory:
		return m.memStats.Total
	case AvailableMemory:
		return m.memStats.Available
	case UsedMemory:
		return m.memStats.Used
	case TotalSwapMemory:
		return m.memStats.SwapTotal
	case FreeSwapMemory:
		return m.memStats.SwapFree
	// case UsedSwapMemory:
	// 	return m.memStats.SwapCached
	default:
		log.Debug().Caller().
			Msg("Unexpected memory state measurement requested.")
		return 0
	}
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
	latest := getStats()
	sendStats(latest, status)
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				latest := getStats()
				sendStats(latest, status)
			}
		}
	}()
}

func getStats() *memory {
	stats := &memory{}
	if m, err := mem.VirtualMemory(); err != nil {
		log.Debug().Err(err).Caller().
			Msg("Problem fetching memory stats.")
		stats.memStats = nil
	} else {
		stats.memStats = m
	}
	return stats
}

func sendStats(latest *memory, status chan interface{}) {
	stats := []memoryStat{TotalMemory, AvailableMemory, UsedMemory, TotalSwapMemory, FreeSwapMemory}
	if latest.memStats != nil {
		for _, stat := range stats {
			latest.name = stat
			status <- latest
		}
	}
}
