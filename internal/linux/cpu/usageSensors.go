// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cpu

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	totalCPUString = "cpu"
)

var ErrParseCPUUsage = errors.New("could not parse CPU usage")

//nolint:lll
var times = [...]string{"user_time", "nice_time", "system_time", "idle_time", "iowait_time", "irq_time", "softirq_time", "steal_time", "guest_time", "guest_nice_time"}

func newRate(valueStr string) *linux.RateValue[uint64] {
	r := &linux.RateValue[uint64]{}
	valueInt, _ := strconv.ParseUint(valueStr, 10, 64) //nolint:errcheck // if we can't parse it, value will be 0.
	r.Calculate(valueInt, 0)

	return r
}

//revive:disable:argument-limit // Not very useful to reduce the number of arguments.
func newRateSensor(ctx context.Context, name, icon, units string, value uint64, total string) (models.Entity, error) {
	entity, err := sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithUnits(units),
		sensor.WithIcon(icon),
		sensor.WithState(value),
		sensor.WithAttribute("Total", total),
		sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
	)
	if err != nil {
		return entity, fmt.Errorf("could not generate %s sensor: %w", name, err)
	}

	return entity, nil
}

func newUsageSensor(ctx context.Context, clktck int64, details []string, category models.EntityCategory) (models.Entity, error) {
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

	value, attributes := generateUsage(clktck, details[1:])

	entity, err := sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.WithUnits("%"),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithState(value),
		sensor.WithAttributes(attributes),
		sensor.WithIcon("mdi:chip"),
		sensor.WithCategory(category),
	)
	if err != nil {
		return entity, fmt.Errorf("could not generate %s sensor: %w", name, err)
	}

	return entity, nil
}

func newCountSensor(ctx context.Context, name, icon, units, valueStr string) (models.Entity, error) {
	valueInt, _ := strconv.Atoi(valueStr) //nolint:errcheck // if we can't parse it, value will be 0.

	entity, err := sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithUnits(units),
		sensor.WithIcon(icon),
		sensor.WithState(valueInt),
		sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
	)
	if err != nil {
		return entity, fmt.Errorf("could not generate %s sensor: %w", name, err)
	}

	return entity, nil
}

func generateUsage(clktck int64, details []string) (float64, map[string]any) {
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
