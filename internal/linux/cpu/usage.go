// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package cpu

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	cpuUsageUpdateInterval = 10 * time.Second
	cpuUsageUpdateJitter   = 500 * time.Millisecond

	cpuUsageWorkerID = "cpu_usage_sensors"

	totalCPUString = "cpu"
)

var ErrParseCPUUsage = errors.New("could not parse CPU usage")

// UsagePrefs are the preferences for the CPU usage worker.
type UsagePrefs struct {
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of CPU usage sensors (default 10s)."`
	preferences.CommonWorkerPrefs
}

type usageWorker struct {
	boottime    time.Time
	rateSensors map[string]*rateSensor
	path        string
	linux.PollingSensorWorker
	clktck int64
	delta  time.Duration
}

func (w *usageWorker) UpdateDelta(delta time.Duration) {
	w.delta = delta
}

func (w *usageWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	return w.getUsageStats()
}

func (w *usageWorker) PreferencesID() string {
	return cpuUsageWorkerID
}

func (w *usageWorker) DefaultPreferences() UsagePrefs {
	return UsagePrefs{
		UpdateInterval: cpuUsageUpdateInterval.String(),
	}
}

func NewUsageWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	var err error

	worker := linux.NewPollingSensorWorker(cpuUsageWorkerID, cpuUsageUpdateInterval, cpuUsageUpdateJitter)

	clktck, found := linux.CtxGetClkTck(ctx)
	if !found {
		return worker, fmt.Errorf("%w: no clktck value", linux.ErrInvalidCtx)
	}

	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return worker, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx)
	}

	cpuUsageWorker := &usageWorker{
		path:     filepath.Join(linux.ProcFSRoot, "stat"),
		boottime: boottime,
		clktck:   clktck,
		rateSensors: map[string]*rateSensor{
			"ctxt":      newRateSensor("CPU Context Switch Rate", "mdi:counter", "ctx/s"),
			"processes": newRateSensor("Processes Creation Rate", "mdi:application-cog", "processes/s"),
		},
	}

	prefs, err := preferences.LoadWorker(ctx, cpuUsageWorker)
	if err != nil {
		return worker, fmt.Errorf("could not load preferences: %w", err)
	}

	// If disabled, don't use the addressWorker.
	if prefs.Disabled {
		return worker, nil
	}

	interval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not parse update interval, using default value.",
			slog.String("requested_value", prefs.UpdateInterval),
			slog.String("default_value", cpuUsageUpdateInterval.String()))
		// Save preferences with default interval value.
		prefs.UpdateInterval = cpuUsageUpdateInterval.String()
		if err := preferences.SaveWorker(ctx, cpuUsageWorker, *prefs); err != nil {
			logging.FromContext(ctx).Warn("Could not save preferences.", slog.Any("error", err))
		}

		interval = cpuUsageUpdateInterval
	}

	worker.PollInterval = interval
	worker.PollingSensorType = cpuUsageWorker

	return worker, nil
}

func (w *usageWorker) getUsageStats() ([]sensor.Entity, error) {
	var sensors []sensor.Entity

	statsFH, err := os.Open(w.path)
	if err != nil {
		return nil, fmt.Errorf("fetch cpu usage: %w", err)
	}

	defer statsFH.Close()

	statsFile := bufio.NewScanner(statsFH)
	for statsFile.Scan() {
		// Set up word scanner for line.
		line := bufio.NewScanner(bytes.NewReader(statsFile.Bytes()))
		line.Split(bufio.ScanWords)
		// Split line by words
		var cols []string
		for line.Scan() {
			cols = append(cols, line.Text())
		}

		if len(cols) == 0 {
			return sensors, ErrParseCPUUsage
		}
		// Create a sensor depending on the line.
		switch {
		case cols[0] == totalCPUString:
			sensors = append(sensors, newUsageSensor(w.clktck, cols, types.CategoryDefault))
		case strings.Contains(cols[0], "cpu"):
			sensors = append(sensors, newUsageSensor(w.clktck, cols, types.CategoryDiagnostic))
		case cols[0] == "ctxt":
			if _, found := w.rateSensors["ctxt"]; found {
				w.rateSensors["ctxt"].update(w.delta, cols[1])
			} else {
				w.rateSensors["ctxt"] = newRateSensor("CPU Context Switch Rate", "mdi:counter", "ctx/s")
			}

			sensors = append(sensors, *w.rateSensors["ctxt"].Entity)
		case cols[0] == "processes":
			if _, found := w.rateSensors["processes"]; found {
				w.rateSensors["processes"].update(w.delta, cols[1])
			} else {
				w.rateSensors["processes"] = newRateSensor("Processes Creation Rate", "mdi:application-cog", "processes/s")
			}

			sensors = append(sensors, *w.rateSensors["processes"].Entity)
		case cols[0] == "procs_running":
			sensors = append(sensors, newCountSensor("Processes Running", "mdi:application-cog", cols[1]))
		case cols[0] == "procs_blocked":
			sensors = append(sensors, newCountSensor("Processes Blocked", "mdi:application-cog", cols[1]))
		}
	}

	return sensors, nil
}

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
