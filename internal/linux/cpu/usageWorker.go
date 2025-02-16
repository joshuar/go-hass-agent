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
	"strconv"
	"strings"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
)

const (
	cpuUsageUpdateInterval = 10 * time.Second
	cpuUsageUpdateJitter   = 500 * time.Millisecond

	cpuUsageWorkerID      = "cpu_usage_sensors"
	cpuUsagePreferencesID = prefPrefix + "usage"
)

var ErrInitUsageWorker = errors.New("could not init CPU usage worker")

type usageWorker struct {
	boottime    time.Time
	rateSensors map[string]*linux.RateValue[uint64]
	path        string
	linux.PollingSensorWorker
	clktck int64
	delta  time.Duration
}

func (w *usageWorker) UpdateDelta(delta time.Duration) {
	w.delta = delta
}

func (w *usageWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	return w.getUsageStats(ctx)
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
		return nil, errors.Join(ErrInitUsageWorker, fmt.Errorf("%w: no clktck value", linux.ErrInvalidCtx))
	}

	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return nil, errors.Join(ErrInitUsageWorker, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx))
	}

	cpuUsageWorker := &usageWorker{
		path:     filepath.Join(linux.ProcFSRoot, "stat"),
		boottime: boottime,
		clktck:   clktck,
		rateSensors: map[string]*linux.RateValue[uint64]{
			"ctxt":      newRate("0"),
			"processes": newRate("0"),
		},
	}

	prefs, err := preferences.LoadWorker(cpuUsageWorker)
	if err != nil {
		return nil, errors.Join(ErrInitUsageWorker, err)
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

func (w *usageWorker) getUsageStats(ctx context.Context) ([]models.Entity, error) {
	var sensors []models.Entity

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
			entity, err := newUsageSensor(ctx, w.clktck, cols, "")
			if err != nil {
				logging.FromContext(ctx).Warn("Could not generate CPU usage models.", slog.Any("error", err))
				continue
			}

			sensors = append(sensors, entity)
		case strings.Contains(cols[0], "cpu"):
			entity, err := newUsageSensor(ctx, w.clktck, cols, models.Diagnostic)
			if err != nil {
				logging.FromContext(ctx).Warn("Could not generate CPU usage models.", slog.Any("error", err))
				continue
			}

			sensors = append(sensors, entity)
		case cols[0] == "ctxt":
			var state uint64

			if _, found := w.rateSensors["ctxt"]; found {
				currValue, _ := strconv.ParseUint(cols[1], 10, 64) //nolint:errcheck // if we can't parse it, value will be 0.
				state = w.rateSensors["ctxt"].Calculate(currValue, w.delta)
			} else {
				w.rateSensors["ctxt"] = newRate(cols[1])
			}

			entity, err := newRateSensor(ctx, "CPU Context Switch Rate", "mdi:counter", "ctx/s", state, cols[1])
			if err != nil {
				logging.FromContext(ctx).Warn("Could not generate context switch rate sensor.", slog.Any("error", err))
				continue
			}

			sensors = append(sensors, entity)
		case cols[0] == "processes":
			var state uint64

			if _, found := w.rateSensors["processes"]; found {
				currValue, _ := strconv.ParseUint(cols[1], 10, 64) //nolint:errcheck // if we can't parse it, value will be 0.
				state = w.rateSensors["processes"].Calculate(currValue, w.delta)
			} else {
				w.rateSensors["processes"] = newRate(cols[1])
			}

			entity, err := newRateSensor(ctx, "Processes Creation Rate", "mdi:application-cog", "processes/s", state, cols[1])
			if err != nil {
				logging.FromContext(ctx).Warn("Could not generate context switch rate sensor.", slog.Any("error", err))
				continue
			}

			sensors = append(sensors, entity)
		case cols[0] == "procs_running":
			entity, err := newCountSensor(ctx, "Processes Running", "mdi:application-cog", "processes", cols[1])
			if err != nil {
				logging.FromContext(ctx).Warn("Could not generate CPU usage sensor.", slog.Any("error", err))
				continue
			}

			sensors = append(sensors, entity)
		case cols[0] == "procs_blocked":
			entity, err := newCountSensor(ctx, "Processes Blocked", "mdi:application-cog", "processes", cols[1])
			if err != nil {
				logging.FromContext(ctx).Warn("Could not generate CPU usage sensor.", slog.Any("error", err))
				continue
			}

			sensors = append(sensors, entity)
		}
	}

	return sensors, nil
}
