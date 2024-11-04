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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	usageUpdateInterval = 10 * time.Second
	usageUpdateJitter   = 500 * time.Millisecond

	usageWorkerID = "cpu_usage"

	totalCPUString = "cpu"
)

var ErrParseCPUUsage = errors.New("could not parse CPU usage")

type usageWorker struct {
	boottime    time.Time
	path        string
	rateSensors map[string]*rateSensor
	linux.PollingSensorWorker
	clktck int64
	delta  time.Duration
	prefs  WorkerPrefs
}

func (w *usageWorker) UpdateDelta(delta time.Duration) {
	w.delta = delta
}

func (w *usageWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	return w.getUsageStats()
}

func (w *usageWorker) PreferencesID() string {
	return preferencesID
}

func (w *usageWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{}
}

func NewUsageWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	var err error

	worker := linux.NewPollingSensorWorker(usageWorkerID, usageUpdateInterval, usageUpdateJitter)

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

	cpuUsageWorker.prefs, err = preferences.LoadWorkerPreferences(ctx, cpuUsageWorker)
	if err != nil {
		return worker, fmt.Errorf("could not load preferences: %w", err)
	}

	// If disabled, don't use the addressWorker.
	if cpuUsageWorker.prefs.Disabled {
		return worker, nil
	}

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
			if !w.prefs.DisableCPUFreq {
				sensors = append(sensors, newCPUFreqSensor(cols[0]))
			}
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
