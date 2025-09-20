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

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
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

var ErrInitUsageWorker = errors.New("could not init CPU usage worker")

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
	path        string
	clktck      int64
}

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
		rateSensors: map[string]*linux.RateValue[uint64]{
			"ctxt":      newRate("0"),
			"processes": newRate("0"),
		},
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
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

// calculateRate takes a sensor name and value string and calculates the uint64 rate
// value for the sensor.
func (w *usageWorker) calculateRate(name, value string) uint64 {
	var state uint64

	if _, found := w.rateSensors[name]; found {
		currValue, _ := strconv.ParseUint(value, 10, 64)
		state = w.rateSensors[name].Calculate(currValue, w.GetDelta())
	} else {
		w.rateSensors[name] = newRate(value)
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
		return nil, fmt.Errorf("fetch cpu usage: %w", err)
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
			return nil, ErrParseCPUUsage
		}
		// Create a sensor depending on the line.
		switch {
		case cols[0] == totalCPUString:
			sensors = append(sensors, newUsageSensor(ctx, w.clktck, cols, ""))
		case strings.Contains(cols[0], "cpu"):
			sensors = append(sensors, newUsageSensor(ctx, w.clktck, cols, models.EntityCategoryDiagnostic))
		case cols[0] == "ctxt":
			rate := w.calculateRate("ctxt", cols[1])
			sensors = append(sensors, newRateSensor(ctx, "CPU Context Switch Rate", "mdi:counter", "ctx/s", rate, cols[1]))
		case cols[0] == "processes":
			rate := w.calculateRate("processes", cols[1])
			sensors = append(sensors, newRateSensor(ctx, "Processes Creation Rate", "mdi:application-cog", "processes/s", rate, cols[1]))
		case cols[0] == "procs_running":
			sensors = append(sensors, newCountSensor(ctx, "Processes Running", "mdi:application-cog", "processes", cols[1]))
		case cols[0] == "procs_blocked":
			sensors = append(sensors, newCountSensor(ctx, "Processes Blocked", "mdi:application-cog", "processes", cols[1]))
		}
	}

	return sensors, warnings
}
