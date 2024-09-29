// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cpu

import (
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

//nolint:lll
var times = [...]string{"user_time", "nice_time", "system_time", "idle_time", "iowait_time", "irq_time", "softirq_time", "steal_time", "guest_time", "guest_nice_time"}

type rateSensor struct {
	*sensor.Entity
	prevState uint64
}

func (s *rateSensor) update(delta time.Duration, valueStr string) {
	valueInt, _ := strconv.ParseUint(valueStr, 10, 64) //nolint:errcheck // if we can't parse it, value will be 0.

	if uint64(delta.Seconds()) > 0 {
		s.Value = (valueInt - s.prevState) / uint64(delta.Seconds()) / 2
	} else {
		s.Value = 0
	}

	s.Attributes["Total"] = valueInt

	s.prevState = valueInt
}

func newRateSensor(name, icon, units string) *rateSensor {
	return &rateSensor{
		Entity: &sensor.Entity{
			Name:       name,
			StateClass: types.StateClassMeasurement,
			Category:   types.CategoryDiagnostic,
			Units:      units,
			State: &sensor.State{
				ID:   strcase.ToSnake(name),
				Icon: icon,
				Attributes: map[string]any{
					"data_source": linux.DataSrcProcfs,
				},
			},
		},
	}
}

func newUsageSensor(clktck int64, details []string, category types.Category) sensor.Entity {
	var name, id string

	switch {
	case details[0] == totalCPUString:
		name = "Total CPU Usage"
		id = "total_cpu_usage"
	default:
		num := strings.TrimPrefix(details[0], "cpu")
		name = "Core " + num + " CPU Usage"
		id = "core_" + num + "_cpu_usage"
	}

	value, attributes := generateUsageValues(clktck, details[1:])

	return sensor.Entity{
		Name:       name,
		Units:      "%",
		StateClass: types.StateClassMeasurement,
		Category:   category,
		State: &sensor.State{
			ID:         id,
			Value:      value,
			Attributes: attributes,
			Icon:       "mdi:chip",
		},
	}
}

func generateUsageValues(clktck int64, details []string) (float64, map[string]any) {
	var totalTime float64

	attrs := make(map[string]any, len(times))
	attrs["data_source"] = linux.DataSrcProcfs

	for idx, name := range times {
		value, err := strconv.ParseFloat(details[idx], 64)
		if err != nil {
			continue
		}

		cpuTime := value / float64(clktck)
		attrs[name] = cpuTime
		totalTime += cpuTime
	}

	attrs["total_time"] = totalTime

	//nolint:forcetypeassert,mnd // we already parsed the value as a float
	value := attrs["user_time"].(float64) / totalTime * 100

	return value, attrs
}

func newCountSensor(name, icon, valueStr string) sensor.Entity {
	valueInt, _ := strconv.Atoi(valueStr) //nolint:errcheck // if we can't parse it, value will be 0.

	return sensor.Entity{
		Name:       name,
		StateClass: types.StateClassTotalIncreasing,
		Category:   types.CategoryDiagnostic,
		State: &sensor.State{
			ID:    strcase.ToSnake(name),
			Icon:  icon,
			Value: valueInt,
			Attributes: map[string]any{
				"data_source": linux.DataSrcProcfs,
			},
		},
	}
}
