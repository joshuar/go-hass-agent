// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	cpuUsageUpdateInterval = 10 * time.Second
	cpuUsageUpdateJitter   = 500 * time.Millisecond

	cpuUsageWorkerID      = "cpu_usage_sensors"
	cpuUsageWorkerDesc    = "CPU usage stats"
	cpuUsagePreferencesID = prefPrefix + "usage"
)

var (
	ErrInitUsageWorker = errors.New("could not init CPU usage worker")
	ErrGetCPUUsage     = errors.New("could not get CPU usage stats")
)

var (
	allTimes     = []string{"user_time", "nice_time", "system_time", "idle_time", "iowait_time", "irq_time", "softirq_time", "steal_time", "guest_time", "guest_nice_time"}
	idleTimes    = []string{"idle_time", "iowait_time"}
	nonIdleTimes = []string{"user_time", "nice_time", "system_time", "irq_time", "softirq_time", "steal_time"}
	guestTimes   = []string{"guest_time", "guest_nice_time"}
)

var (
	_ quartz.Job                  = (*usageWorker)(nil)
	_ workers.PollingEntityWorker = (*usageWorker)(nil)
)

type usageWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs       *UsagePrefs
	boottime    time.Time
	rateSensors map[string]*linux.RateValue[uint64]
	cpuSensors  map[string]float64
	path        string
	clktck      int64
}

// NewUsageWorker creates a new worker to generate entities for CPU usage and related stats.
func NewUsageWorker(ctx context.Context) (workers.EntityWorker, error) {
	clktck, found := linux.CtxGetClkTck(ctx)
	if !found {
		return nil, errors.Join(ErrInitUsageWorker, fmt.Errorf("%w: no clktck value", linux.ErrInvalidCtx))
	}

	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return nil, errors.Join(ErrInitUsageWorker, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx))
	}

	worker := &usageWorker{
		WorkerMetadata:          models.SetWorkerMetadata(cpuUsageWorkerID, cpuUsageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		path:                    filepath.Join(linux.ProcFSRoot, "stat"),
		boottime:                boottime,
		clktck:                  clktck,
		rateSensors:             make(map[string]*linux.RateValue[uint64]),
		cpuSensors:              make(map[string]float64),
	}

	defaultPrefs := &UsagePrefs{
		UpdateInterval: cpuUsageUpdateInterval.String(),
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(cpuUsagePreferencesID, defaultPrefs)
	if err != nil {
		return nil, errors.Join(ErrInitUsageWorker, err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", cpuUsageWorkerID),
			slog.String("given_interval", worker.prefs.UpdateInterval),
			slog.String("default_interval", cpuUsageUpdateInterval.String()))

		pollInterval = cpuUsageUpdateInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, cpuUsageUpdateJitter)

	return worker, nil
}

func (w *usageWorker) Execute(ctx context.Context) error {
	usageSensors, err := w.getUsageStats(ctx)
	if err != nil {
		return fmt.Errorf("could not get usage stats: %w", err)
	}
	for s := range slices.Values(usageSensors) {
		w.OutCh <- s
	}
	return nil
}

func (w *usageWorker) IsDisabled() bool {
	return w.prefs.Disabled
}

func (w *usageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	err := workers.SchedulePollingWorker(ctx, w, w.OutCh)
	if err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

// calculateRate takes a sensor name and value string and calculates the uint64 rate
// value for the sensor.
func (w *usageWorker) calculateRate(ctx context.Context, name, value string) uint64 {
	var state uint64

	if _, found := w.rateSensors[name]; found {
		currValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			slogctx.FromCtx(ctx).Debug("Could not convert current value of rate.",
				slog.Any("error", err))
		}
		state = w.rateSensors[name].Calculate(currValue, w.GetDelta())
	} else {
		r := &linux.RateValue[uint64]{}
		valueInt, _ := strconv.ParseUint(value, 10, 64)
		r.Calculate(valueInt, 0)
		w.rateSensors[name] = r
	}

	return state
}

func (w *usageWorker) getUsageStats(ctx context.Context) ([]models.Entity, error) {
	var (
		sensors  []models.Entity
		warnings error
	)

	statsFH, err := os.Open(w.path)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to open %s", ErrGetCPUUsage, w.path)
	}

	defer statsFH.Close() //nolint:errcheck

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
			return nil, fmt.Errorf("%w: unable to parse %s", ErrGetCPUUsage, w.path)
		}
		// Create a sensor depending on the line.
		switch {
		case cols[0] == "cpu":
			sensors = append(sensors, newUsageSensor(ctx, w, cols, ""))
		case strings.HasPrefix(cols[0], "cpu"):
			sensors = append(sensors, newUsageSensor(ctx, w, cols, models.EntityCategoryDiagnostic))
		case cols[0] == "ctxt":
			rate := w.calculateRate(ctx, "ctxt", cols[1])
			sensors = append(sensors, newRateSensor(ctx, "CPU Context Switch Rate", "mdi:counter", "ctx/s", rate, cols[1]))
		case cols[0] == "processes":
			rate := w.calculateRate(ctx, "processes", cols[1])
			sensors = append(sensors, newRateSensor(ctx, "Processes Creation Rate", "mdi:application-cog", "processes/s", rate, cols[1]))
		case cols[0] == "procs_running":
			sensors = append(sensors, newCountSensor(ctx, "Processes Running", "mdi:application-cog", "processes", cols[1]))
		case cols[0] == "procs_blocked":
			sensors = append(sensors, newCountSensor(ctx, "Processes Blocked", "mdi:application-cog", "processes", cols[1]))
		}
	}

	return sensors, warnings
}

//revive:disable:argument-limit // Not very useful to reduce the number of arguments.
func newRateSensor(ctx context.Context, name, icon, units string, value uint64, total string) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithUnits(units),
		sensor.WithIcon(icon),
		sensor.WithState(value),
		sensor.WithAttribute("Total", total),
		sensor.WithDataSourceAttribute(linux.DataSrcProcFS),
	)
}

func newUsageSensor(ctx context.Context, worker *usageWorker, details []string, category models.EntityCategory) models.Entity {
	var name, id string

	switch details[0] {
	case "cpu":
		name = "Total CPU Usage"
		id = "total_cpu_usage"
	default:
		num := strings.TrimPrefix(details[0], "cpu")
		name = "Core " + num + " CPU Usage"
		id = "core_" + num + "_cpu_usage"
	}

	var totalTime float64
	var nonIdleTime float64
	var idleTime float64

	attrs := make(map[string]any)
	attrs["data_source"] = linux.DataSrcProcFS

	for idx, name := range allTimes {
		value, err := strconv.ParseFloat(details[idx+1], 64)
		if err != nil {
			continue
		}

		// Divide by clock tick to get a time.
		cpuTime := value / float64(worker.clktck)
		attrs[name] = cpuTime
		// Don't include guest counters in total times.
		if !slices.Contains(guestTimes, name) {
			totalTime += cpuTime
		}
		// Add up non-idle counters.
		if slices.Contains(nonIdleTimes, name) {
			nonIdleTime += cpuTime
		}
		// Add up idle counters.
		if slices.Contains(idleTimes, name) {
			idleTime += cpuTime
		}
	}

	// Add our total times as attributes.
	attrs["total_time"] = totalTime
	attrs["non_idle_total"] = nonIdleTime
	attrs["idle_total"] = idleTime

	// Calculate % for state value.
	var currValue, currTotal float64
	if _, found := worker.cpuSensors[details[0]]; found {
		currValue = nonIdleTime - worker.cpuSensors[details[0]]
	} else {
		currValue = nonIdleTime
	}
	worker.cpuSensors[details[0]] = currValue

	if _, found := worker.cpuSensors[details[0]+"_total"]; found {
		currTotal = totalTime - worker.cpuSensors[details[0]+"_total"]
	} else {
		currTotal = totalTime
	}
	worker.cpuSensors[details[0]+"_total"] = currTotal
	state := currValue / currTotal * 100

	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.WithUnits("%"),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithState(state),
		sensor.WithAttributes(attrs),
		sensor.WithIcon("mdi:chip"),
		sensor.WithCategory(category),
	)
}

func newCountSensor(ctx context.Context, name, icon, units, valueStr string) models.Entity {
	valueInt, _ := strconv.Atoi(valueStr)
	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithUnits(units),
		sensor.WithIcon(icon),
		sensor.WithState(valueInt),
		sensor.WithDataSourceAttribute(linux.DataSrcProcFS),
	)
}
