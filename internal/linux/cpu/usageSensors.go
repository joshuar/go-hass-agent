// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cpu

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	totalCPUString = "cpu"
)

var ErrParseCPUUsage = errors.New("could not parse CPU usage")

//nolint:lll
var times = [...]string{"user_time", "nice_time", "system_time", "idle_time", "iowait_time", "irq_time", "softirq_time", "steal_time", "guest_time", "guest_nice_time"}

type rateSensor struct {
	*sensor.Entity
	prevState uint64
}

func (s *rateSensor) update(delta time.Duration, valueStr string) {
	valueInt, _ := strconv.ParseUint(valueStr, 10, 64) //nolint:errcheck // if we can't parse it, value will be 0.

	if uint64(delta.Seconds()) > 0 {
		s.UpdateValue((valueInt - s.prevState) / uint64(delta.Seconds()) / 2)
	} else {
		s.UpdateValue(0)
	}

	s.UpdateAttribute("Total", valueInt)

	s.prevState = valueInt
}

func newRateSensor(name, icon, units string) *rateSensor {
	sensorDetails := sensor.NewSensor(
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithStateClass(types.StateClassMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithUnits(units),
		sensor.WithState(
			sensor.WithIcon(icon),
			sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
		),
	)

	return &rateSensor{
		Entity: &sensorDetails,
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

	usageSensor := sensor.NewSensor(
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.WithUnits("%"),
		sensor.WithStateClass(types.StateClassMeasurement),
		sensor.WithState(
			sensor.WithValue(value),
			sensor.WithAttributes(attributes),
			sensor.WithIcon("mdi:chip"),
		),
	)

	if category == types.CategoryDiagnostic {
		usageSensor = sensor.AsDiagnostic()(usageSensor)
	}

	return usageSensor
}

func newCountSensor(name, icon, valueStr string) sensor.Entity {
	valueInt, _ := strconv.Atoi(valueStr) //nolint:errcheck // if we can't parse it, value will be 0.

	return sensor.NewSensor(
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithStateClass(types.StateClassMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon(icon),
			sensor.WithValue(valueInt),
			sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
		),
	)
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

	//nolint:forcetypeassert,mnd,errcheck // we already parsed the value as a float
	value := attrs["user_time"].(float64) / totalTime * 100

	return value, attrs
}
