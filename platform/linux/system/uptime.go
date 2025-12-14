// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

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
	uptimePollInterval = 15 * time.Minute
	uptimePollJitter   = time.Minute
)

var (
	_ quartz.Job                  = (*uptimeWorker)(nil)
	_ workers.PollingEntityWorker = (*uptimeWorker)(nil)
)

type uptimeWorker struct {
	*workers.PollingEntityWorkerData
	*models.WorkerMetadata

	prefs *UptimePrefs
}

func NewUptimeTimeWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &uptimeWorker{
		WorkerMetadata:          models.SetWorkerMetadata("uptime", "System uptime"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &UptimePrefs{
		UpdateInterval: uptimePollInterval.String(),
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(infoWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = uptimePollInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, uptimePollJitter)

	return worker, nil
}

func (w *uptimeWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk IO worker: %w", err)
	}
	return w.OutCh, nil
}

func (w *uptimeWorker) Execute(ctx context.Context) error {
	w.OutCh <- sensor.NewSensor(ctx,
		sensor.WithName("Uptime"),
		sensor.WithID("uptime"),
		sensor.AsDiagnostic(),
		sensor.WithDeviceClass(class.SensorClassDuration),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits("h"),
		sensor.WithIcon("mdi:restart"),
		sensor.WithState(w.getUptime(ctx)/60/60),
		sensor.WithDataSourceAttribute(linux.ProcFSRoot),
		sensor.WithAttribute("native_unit_of_measurement", "h"),
	)
	return nil
}

func (w *uptimeWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

// getUptime retrieve the uptime of the device running Go Hass Agent, in
// seconds. If the uptime cannot be retrieved, it will return 0.
func (w *uptimeWorker) getUptime(ctx context.Context) float64 {
	data, err := os.Open(linux.UptimeFile)
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Unable to retrieve uptime.", slog.Any("error", err))

		return 0
	}

	defer data.Close()

	line := bufio.NewScanner(data)
	line.Split(bufio.ScanWords)

	if !line.Scan() {
		slogctx.FromCtx(ctx).Debug("Could not parse uptime.")

		return 0
	}

	uptimeValue, err := strconv.ParseFloat(line.Text(), 64)
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Could not parse uptime.")

		return 0
	}

	return uptimeValue
}
