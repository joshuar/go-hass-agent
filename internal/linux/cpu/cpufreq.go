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
	freqWorker := &freqWorker{}

	prefs, err := preferences.LoadWorker(ctx, freqWorker)
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
			slog.String("worker", cpuFreqWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", cpuFreqUpdateInterval.String()))

		pollInterval = cpuFreqUpdateInterval
	}

	pollWorker := linux.NewPollingSensorWorker(cpuFreqWorkerID, pollInterval, cpuFreqUpdateJitter)
	pollWorker.PollingSensorType = freqWorker

	return pollWorker, nil
}
