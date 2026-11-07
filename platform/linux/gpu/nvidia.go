// Copyright 2026 opticon454
// SPDX-License-Identifier: MIT

package gpu

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	nvidiaUpdateInterval = 10 * time.Second
	nvidiaUpdateJitter   = 500 * time.Millisecond
	nvidiaPreferencesID  = prefPrefix + "nvidia"

	nvidiaSMICommand = "nvidia-smi"
	nvidiaSMITimeout = 10 * time.Second
	nvidiaDataSource = "nvidia-smi"
	nvidiaIcon       = "mdi:expansion-card"

	// The first three columns identify the GPU. The rest map, in order, to nvidiaMetrics. Newer drivers can report
	// the instantaneous board power; older ones only have power.draw, which is the same figure averaged over a second.
	nvidiaQueryFields       = "index,uuid,name,power.draw.instant,temperature.gpu,utilization.gpu,memory.used,memory.total,fan.speed,clocks.sm"
	nvidiaQueryFieldsLegacy = "index,uuid,name,power.draw,temperature.gpu,utilization.gpu,memory.used,memory.total,fan.speed,clocks.sm"
	nvidiaIdentityColumns   = 3
)

var (
	ErrInitNvidiaWorker = errors.New("could not init Nvidia GPU worker")
	ErrGetNvidiaStats   = errors.New("could not get Nvidia GPU stats")
)

var (
	_ quartz.Job                  = (*nvidiaWorker)(nil)
	_ workers.PollingEntityWorker = (*nvidiaWorker)(nil)
)

// nvidiaMetric describes one sensor created per GPU from one nvidia-smi column.
type nvidiaMetric struct {
	key         string
	label       string
	units       string
	deviceClass models.SensorDeviceClass
	hasClass    bool
}

// nvidiaMetrics are in the same order as the nvidia-smi columns that follow the identity columns.
var nvidiaMetrics = []nvidiaMetric{
	{"power", "Power", "W", models.SensorClassPower, true},
	{"temperature", "Temperature", "°C", models.SensorClassTemperature, true},
	{"utilization", "Utilization", "%", 0, false},
	{"memory_used", "Memory Used", "MiB", models.SensorClassDataSize, true},
	{"memory_total", "Memory Total", "MiB", models.SensorClassDataSize, true},
	{"fan_speed", "Fan Speed", "%", 0, false},
	{"clock", "Graphics Clock", "MHz", models.SensorClassFrequency, true},
}

// nvidiaGPU is one GPU as reported by nvidia-smi. A metric the card does not report (for example the fan on a
// passively cooled card shows [N/A]) is absent from stats.
type nvidiaGPU struct {
	index int
	uuid  string
	name  string
	stats map[string]float64
}

// parseNvidiaSMI parses the output of nvidia-smi --query-gpu=... --format=csv,noheader,nounits.
func parseNvidiaSMI(reader io.Reader) ([]nvidiaGPU, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetNvidiaStats, err)
	}

	gpus := make([]nvidiaGPU, 0, len(records))

	for _, record := range records {
		if len(record) != nvidiaIdentityColumns+len(nvidiaMetrics) {
			return nil, fmt.Errorf("%w: expected %d columns, got %d", ErrGetNvidiaStats,
				nvidiaIdentityColumns+len(nvidiaMetrics), len(record))
		}

		index, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			return nil, fmt.Errorf("%w: bad GPU index %q", ErrGetNvidiaStats, record[0])
		}

		gpu := nvidiaGPU{index: index, uuid: record[1], name: record[2], stats: make(map[string]float64)}

		for i, metric := range nvidiaMetrics {
			// Values such as [N/A] or [Not Supported] do not parse and are left out.
			if value, err := strconv.ParseFloat(strings.TrimSpace(record[nvidiaIdentityColumns+i]), 64); err == nil {
				gpu.stats[metric.key] = value
			}
		}

		gpus = append(gpus, gpu)
	}

	return gpus, nil
}

// newNvidiaSensors creates the sensors for one GPU.
func newNvidiaSensors(ctx context.Context, gpu nvidiaGPU) []models.Entity {
	sensors := make([]models.Entity, 0, len(nvidiaMetrics))

	for _, metric := range nvidiaMetrics {
		value, found := gpu.stats[metric.key]
		if !found {
			continue
		}

		options := []models.SensorOption{
			models.WithName(fmt.Sprintf("Nvidia GPU %d %s", gpu.index, metric.label)),
			models.WithID(fmt.Sprintf("nvidia_gpu_%d_%s", gpu.index, metric.key)),
			models.WithUnits(metric.units),
			models.WithStateClass(models.StateMeasurement),
			models.WithIcon(nvidiaIcon),
			models.WithState(value),
			models.WithDataSourceAttribute(nvidiaDataSource),
			models.WithAttribute("native_unit_of_measurement", metric.units),
			models.WithAttribute("gpu_name", gpu.name),
			models.WithAttribute("gpu_uuid", gpu.uuid),
		}
		if metric.hasClass {
			options = append(options, models.WithDeviceClass(metric.deviceClass))
		}

		sensors = append(sensors, models.NewSensor(ctx, options...))
	}

	return sensors
}

type nvidiaWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *NvidiaPrefs
	path  string
	// query runs nvidia-smi for the given query fields. It is a field so that tests can supply fixed output.
	query func(ctx context.Context, fields string) ([]byte, error)
}

// NewNvidiaWorker creates a worker to generate sensors for Nvidia GPUs (power, temperature, utilization, memory, fan
// speed and clock) using nvidia-smi. The worker is disabled when nvidia-smi is not installed.
func NewNvidiaWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &nvidiaWorker{
		WorkerMetadata:          models.SetWorkerMetadata("nvidia_gpu", "Nvidia GPU metrics"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}
	worker.query = worker.runNvidiaSMI

	// Not finding nvidia-smi is normal on machines without an Nvidia GPU. It leaves the path empty, which disables the worker.
	worker.path, _ = exec.LookPath(nvidiaSMICommand) //nolint:errcheck

	defaultPrefs := &NvidiaPrefs{
		UpdateInterval: nvidiaUpdateInterval.String(),
	}

	var err error

	worker.prefs, err = workers.LoadWorkerPreferences(nvidiaPreferencesID, defaultPrefs)
	if err != nil {
		return worker, errors.Join(ErrInitNvidiaWorker, err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = nvidiaUpdateInterval
	}

	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, nvidiaUpdateJitter)

	return worker, nil
}

func (w *nvidiaWorker) runNvidiaSMI(ctx context.Context, fields string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, nvidiaSMITimeout)
	defer cancel()

	//nolint:gosec // the command path comes from exec.LookPath and the arguments are constants.
	output, err := exec.CommandContext(ctx, w.path, "--query-gpu="+fields, "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetNvidiaStats, err)
	}

	return output, nil
}

// Execute queries all the GPUs and sends their sensors.
func (w *nvidiaWorker) Execute(ctx context.Context) error {
	output, err := w.query(ctx, nvidiaQueryFields)
	if err != nil {
		// Older drivers do not know power.draw.instant.
		slogctx.FromCtx(ctx).Debug("Query with instantaneous power failed, trying the legacy query.", slog.Any("error", err))

		output, err = w.query(ctx, nvidiaQueryFieldsLegacy)
		if err != nil {
			return fmt.Errorf("could not query nvidia-smi: %w", err)
		}
	}

	gpus, err := parseNvidiaSMI(strings.NewReader(string(output)))
	if err != nil {
		return err
	}

	for gpu := range slices.Values(gpus) {
		for sensor := range slices.Values(newNvidiaSensors(ctx, gpu)) {
			w.OutCh <- sensor
		}
	}

	return nil
}

// IsDisabled is true when disabled in the preferences or when nvidia-smi is not installed.
func (w *nvidiaWorker) IsDisabled() bool {
	return w.prefs.Disabled || w.path == ""
}

// Start schedules the worker.
func (w *nvidiaWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start Nvidia GPU worker: %w", err)
	}

	return w.OutCh, nil
}
