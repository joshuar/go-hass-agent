// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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
	freqFile     = "cpufreq/scaling_cur_freq"
	governorFile = "cpufreq/scaling_governor"
	driverFile   = "cpufreq/scaling_driver"

	cpuFreqIcon  = "mdi:chip"
	cpuFreqUnits = "kHz"

	cpuFreqUpdateInterval = 30 * time.Second
	cpuFreqUpdateJitter   = time.Second

	cpuFreqPreferencesID = prefPrefix + "frequencies"
)

var (
	_ quartz.Job                  = (*freqWorker)(nil)
	_ workers.PollingEntityWorker = (*freqWorker)(nil)
)

type freqWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *FreqPrefs
}

// NewFreqWorker creates a worker that will monitor and report CPU frequencies.
func NewFreqWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &freqWorker{
		WorkerMetadata:          models.SetWorkerMetadata("cpu_frequency", "CPU Frequency metrics"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &FreqPrefs{
		UpdateInterval: cpuFreqUpdateInterval.String(),
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(cpuFreqPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("unable to load CPU frequency preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = cpuFreqUpdateInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, cpuFreqUpdateJitter)

	return worker, nil
}

func (w *freqWorker) Execute(ctx context.Context) error {
	var warnings error
	for idx := range runtime.NumCPU() {
		w.OutCh <- newCPUFreqSensor(ctx, "cpu"+strconv.Itoa(idx))
	}
	return warnings
}

func (w *freqWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *freqWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	err := workers.SchedulePollingWorker(ctx, w, w.OutCh)
	if err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

type cpuFreq struct {
	cpu      string
	governor string
	driver   string
	freq     string
}

func newCPUFreqSensor(ctx context.Context, id string) models.Entity {
	info := getCPUFreqs(ctx, id)
	num := strings.TrimPrefix(info.cpu, "cpu")

	return sensor.NewSensor(ctx,
		sensor.WithName("Core "+num+" Frequency"),
		sensor.WithID("cpufreq_core"+num+"_frequency"),
		sensor.AsTypeSensor(),
		sensor.WithUnits(cpuFreqUnits),
		sensor.WithDeviceClass(class.SensorClassFrequency),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithIcon(cpuFreqIcon),
		sensor.WithState(info.freq),
		sensor.WithAttributes(models.Attributes{
			"governor":                   info.governor,
			"driver":                     info.driver,
			"data_source":                linux.DataSrcSysFS,
			"native_unit_of_measurement": cpuFreqUnits,
		}),
	)
}

func getCPUFreqs(ctx context.Context, id string) *cpuFreq {
	return &cpuFreq{
		cpu:      id,
		freq:     readCPUFreqProp(ctx, id, freqFile),
		governor: readCPUFreqProp(ctx, id, governorFile),
		driver:   readCPUFreqProp(ctx, id, driverFile),
	}
}

// readCPUFreqProp retrieves the current cpu freq governor for this cpu. If
// it cannot be found, it returns "unknown".
func readCPUFreqProp(ctx context.Context, id, file string) string {
	path := filepath.Join(linux.SysFSRoot, "devices", "system", "cpu", id, file)

	// Read the current scaling driver.
	prop, err := os.ReadFile(path) // #nosec:G304
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Could not read CPU freq property.",
			slog.String("cpu", id),
			slog.String("property", file),
			slog.Any("error", err))

		return "unknown"
	}

	return string(bytes.TrimSpace(prop))
}
