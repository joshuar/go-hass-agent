// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	cpuUsageUpdateInterval = 10 * time.Second
	cpuUsageUpdateJitter   = 500 * time.Millisecond

	cpuUsageWorkerID      = "cpu_usage_sensors"
	cpuUsagePreferencesID = prefPrefix + "usage"
)

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
	return cpuUsagePreferencesID
}

func (w *usageWorker) DefaultPreferences() UsagePrefs {
	return UsagePrefs{
		UpdateInterval: cpuUsageUpdateInterval.String(),
	}
}

func NewUsageWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	clktck, found := linux.CtxGetClkTck(ctx)
	if !found {
		return nil, fmt.Errorf("%w: no clktck value", linux.ErrInvalidCtx)
	}

	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return nil, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx)
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

	prefs, err := preferences.LoadWorker(cpuUsageWorker)
	if err != nil {
		return nil, fmt.Errorf("could not load preferences: %w", err)
	}

	//nolint:nilnil
	if prefs.Disabled {
		return nil, nil
	}

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", cpuUsageWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", cpuUsageUpdateInterval.String()))

		pollInterval = cpuUsageUpdateInterval
	}

	worker := linux.NewPollingSensorWorker(cpuUsageWorkerID, pollInterval, cpuUsageUpdateJitter)
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
			sensors = append(sensors, newCountSensor("Processes Running", "mdi:application-cog", "processes", cols[1]))
		case cols[0] == "procs_blocked":
			sensors = append(sensors, newCountSensor("Processes Blocked", "mdi:application-cog", "processes", cols[1]))
		}
	}

	return sensors, nil
}
