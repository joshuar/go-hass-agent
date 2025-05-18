// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	cpuFreqUpdateInterval = 30 * time.Second
	cpuFreqUpdateJitter   = time.Second

	cpuFreqWorkerID      = "cpu_freq_sensors"
	cpuFreqWorkerDesc    = "CPU frequency stats"
	cpuFreqPreferencesID = prefPrefix + "frequencies"
)

var (
	_ quartz.Job                  = (*freqWorker)(nil)
	_ workers.PollingEntityWorker = (*freqWorker)(nil)
)

var ErrInitFreqWorker = errors.New("could not start CPU frequencies worker")

type freqWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData
	prefs *FreqWorkerPrefs
}

func NewFreqWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &freqWorker{
		WorkerMetadata:          models.SetWorkerMetadata(cpuFreqWorkerID, cpuFreqWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitFreqWorker, err)
	}
	worker.prefs = prefs

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", cpuFreqWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", cpuFreqUpdateInterval.String()))

		pollInterval = cpuFreqUpdateInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, cpuFreqUpdateJitter)

	return worker, nil
}

func (w *freqWorker) Execute(ctx context.Context) error {
	var warnings error
	for idx := range totalCPUs {
		w.OutCh <- newCPUFreqSensor(ctx, "cpu"+strconv.Itoa(idx))
	}
	return warnings
}

func (w *freqWorker) PreferencesID() string {
	return cpuFreqPreferencesID
}

func (w *freqWorker) DefaultPreferences() FreqWorkerPrefs {
	return FreqWorkerPrefs{
		UpdateInterval: cpuFreqUpdateInterval.String(),
	}
}

func (w *freqWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *freqWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}
