// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/mem"
)

//go:generate stringer -type=memoryStat -output memorySensorProps.go -linecomment

const (
	memoryTotal     memoryStat = iota + 1 // Memory Total
	memoryAvailable                       // Memory Available
	memoryUsed                            // Memory Used
	swapMemoryTotal                       // Swap Memory Total
	swapMemoryUsed                        // Swap Memory Used
	swapMemoryFree                        // Swap Memory Free
)

type memoryStat int

type memory struct {
	value uint64
	name  memoryStat
}

// memory implements hass.SensorUpdate

func (m *memory) Name() string {
	return m.name.String()
}

func (m *memory) ID() string {
	return strings.ToLower(strcase.ToSnake(m.name.String()))
}

func (m *memory) Icon() string {
	return "mdi:memory"
}

func (m *memory) SensorType() sensor.SensorType {
	return sensor.TypeSensor
}

func (m *memory) DeviceClass() sensor.SensorDeviceClass {
	return sensor.Data_size
}

func (m *memory) StateClass() sensor.SensorStateClass {
	return sensor.StateTotal
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
	return struct {
		DataSource string `json:"Data Source"`
	}{
		DataSource: "procfs",
	}
}

func MemoryUpdater(ctx context.Context, status chan interface{}) {

	sendMemStats := func() {
		stats := []memoryStat{memoryTotal, memoryAvailable, memoryUsed, swapMemoryTotal, swapMemoryFree}
		var memDetails *mem.VirtualMemoryStat
		var err error
		if memDetails, err = mem.VirtualMemoryWithContext(ctx); err != nil {
			log.Debug().Err(err).Caller().
				Msg("Problem fetching memory stats.")
			return
		}
		for _, stat := range stats {
			var statValue uint64
			switch stat {
			case memoryTotal:
				statValue = memDetails.Total
			case memoryAvailable:
				statValue = memDetails.Available
			case memoryUsed:
				statValue = memDetails.Used
			case swapMemoryTotal:
				statValue = memDetails.SwapTotal
			case swapMemoryFree:
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

	helpers.PollSensors(ctx, sendMemStats, time.Minute, time.Second*5)
}
