// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
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

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	cpuFreqUpdateInterval = 30 * time.Second
	cpuFreqUpdateJitter   = time.Second

	cpuFreqWorkerID = "cpu_freq_sensors"

	freqFile     = "cpufreq/scaling_cur_freq"
	governorFile = "cpufreq/scaling_governor"
	driverFile   = "cpufreq/scaling_driver"

	cpuFreqIcon  = "mdi:chip"
	cpuFreqUnits = "kHz"
)

var totalCPUs = runtime.NumCPU()

// FreqWorkerPrefs are the preferences for the CPU frequency worker.
type FreqWorkerPrefs struct {
	UpdateInterval string `toml:"update_interval" comment:"Time between updates of CPU frequency sensors (default 30s)."`
	preferences.CommonWorkerPrefs
}

type freqWorker struct{}

func (w *freqWorker) UpdateDelta(_ time.Duration) {}

func (w *freqWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	sensors := make([]sensor.Entity, totalCPUs)

	for i := range totalCPUs {
		sensors[i] = newCPUFreqSensor("cpu" + strconv.Itoa(i))
	}

	return sensors, nil
}

func (w *freqWorker) PreferencesID() string {
	return cpuFreqWorkerID
}

func (w *freqWorker) DefaultPreferences() FreqWorkerPrefs {
	return FreqWorkerPrefs{
		UpdateInterval: cpuFreqUpdateInterval.String(),
	}
}

type cpuFreq struct {
	cpu      string
	governor string
	driver   string
	freq     string
}

func newCPUFreqSensor(id string) sensor.Entity {
	info := getCPUFreqs(id)
	num := strings.TrimPrefix(info.cpu, "cpu")

	return sensor.NewSensor(
		sensor.WithName("Core "+num+" Frequency"),
		sensor.WithID("cpufreq_core"+num+"_frequency"),
		sensor.AsTypeSensor(),
		sensor.WithUnits(cpuFreqUnits),
		sensor.WithDeviceClass(types.SensorDeviceClassFrequency),
		sensor.WithStateClass(types.StateClassMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon(cpuFreqIcon),
			sensor.WithValue(info.freq),
			sensor.WithAttributes(map[string]any{
				"governor":                   info.governor,
				"driver":                     info.driver,
				"data_source":                linux.DataSrcSysfs,
				"native_unit_of_measurement": cpuFreqUnits,
			}),
		),
	)
}

func getCPUFreqs(id string) *cpuFreq {
	return &cpuFreq{
		cpu:      id,
		freq:     readCPUFreqProp(id, freqFile),
		governor: readCPUFreqProp(id, governorFile),
		driver:   readCPUFreqProp(id, driverFile),
	}
}

// readCPUFreqProp retrieves the current cpu freq governor for this cpu. If
// it cannot be found, it returns "unknown".
func readCPUFreqProp(id, file string) string {
	path := filepath.Join(linux.SysFSRoot, "devices", "system", "cpu", id, file)

	// Read the current scaling driver.
	prop, err := os.ReadFile(path)
	if err != nil {
		slog.Debug("Could not read CPU freq property.",
			slog.String("cpu", id),
			slog.String("property", file),
			slog.Any("error", err))

		return "unknown"
	}

	return string(bytes.TrimSpace(prop))
}

func NewFreqWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	var err error

	pollWorker := linux.NewPollingSensorWorker(cpuFreqWorkerID, cpuFreqUpdateInterval, cpuFreqUpdateJitter)

	worker := &freqWorker{}

	prefs, err := preferences.LoadWorker(ctx, worker)
	if err != nil {
		return pollWorker, fmt.Errorf("could not load preferences: %w", err)
	}

	// If disabled, don't use the addressWorker.
	if prefs.Disabled {
		return pollWorker, nil
	}

	interval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not parse update interval, using default value.",
			slog.String("requested_value", prefs.UpdateInterval),
			slog.String("default_value", cpuFreqUpdateInterval.String()))
		// Save preferences with default interval value.
		prefs.UpdateInterval = cpuFreqUpdateInterval.String()
		if err := preferences.SaveWorker(ctx, worker, *prefs); err != nil {
			logging.FromContext(ctx).Warn("Could not save preferences.", slog.Any("error", err))
		}

		interval = cpuUsageUpdateInterval
	}

	pollWorker.PollInterval = interval
	pollWorker.PollingSensorType = worker

	return pollWorker, nil
}
