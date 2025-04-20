// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	uptimePollInterval = 15 * time.Minute
	uptimePollJitter   = time.Minute

	uptimeWorkerID   = "uptime_sensor"
	uptimeWorkerDesc = "Uptime stats"
)

var (
	_ quartz.Job                  = (*uptimeWorker)(nil)
	_ workers.PollingEntityWorker = (*uptimeWorker)(nil)
)

var ErrInitUptimeWorker = errors.New("could not init uptime worker")

type uptimeWorker struct {
	prefs *UptimePrefs
	*workers.PollingEntityWorkerData
	*models.WorkerMetadata
}

func (w *uptimeWorker) Execute(ctx context.Context) error {
	entity, err := sensor.NewSensor(ctx,
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
	if err != nil {
		return fmt.Errorf("could not generate uptime sensor: %w", err)
	}
	w.OutCh <- entity
	return nil
}

func (w *uptimeWorker) PreferencesID() string {
	return infoWorkerPreferencesID
}

func (w *uptimeWorker) DefaultPreferences() UptimePrefs {
	return UptimePrefs{
		UpdateInterval: uptimePollInterval.String(),
	}
}

func (w *uptimeWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *uptimeWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk IO worker: %w", err)
	}
	return w.OutCh, nil
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

func NewUptimeTimeWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &uptimeWorker{
		WorkerMetadata:          models.SetWorkerMetadata(uptimeWorkerID, uptimeWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitUptimeWorker, err)
	}
	worker.prefs = prefs

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", uptimeWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", uptimePollInterval.String()))

		pollInterval = uptimePollInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, uptimePollJitter)

	return worker, nil
}
