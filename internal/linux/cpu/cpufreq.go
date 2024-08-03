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
	"strings"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	sysFSPath    = "/sys/devices/system/cpu/"
	freqFile     = "cpufreq/scaling_cur_freq"
	governorFile = "cpufreq/scaling_governor"
	driverFile   = "cpufreq/scaling_driver"

	cpuFreqWorkerID       = "cpu_freq_sensors"
	cpuFreqUpdateInterval = 5 * time.Second
	cpuFreqUpdateJitter   = 10 * time.Millisecond

	cpuFreqIcon  = "mdi:chip"
	cpuFreqUnits = "Hz"
)

type cpuFreq struct {
	cpu      string
	governor string
	driver   string
	freq     string
}

type cpuFreqSensor struct {
	*cpuFreq
	linux.Sensor
}

type cpuFreqWorker struct {
	path string
}

func (w *cpuFreqWorker) Interval() time.Duration { return cpuFreqUpdateInterval }

func (w *cpuFreqWorker) Jitter() time.Duration { return cpuFreqUpdateJitter }

func (w *cpuFreqWorker) Sensors(_ context.Context, _ time.Duration) ([]sensor.Details, error) {
	freqs, err := getCPUFreqs(w.path)
	if err != nil {
		return nil, fmt.Errorf("could not fetch cpu frequencies: %w", err)
	}

	sensors := make([]sensor.Details, 0, len(freqs))

	for _, freq := range freqs {
		sensors = append(sensors, newCPUFreqSensor(freq))
	}

	return sensors, nil
}

func (s *cpuFreqSensor) Name() string {
	num := strings.TrimPrefix(s.cpu, "cpu")

	return "Core " + num + " Frequency"
}

func (s *cpuFreqSensor) ID() string {
	num := strings.TrimPrefix(s.cpu, "cpu")

	return "cpufreq_core" + num + "_frequency"
}

func (s *cpuFreqSensor) Attributes() map[string]any {
	return map[string]any{
		"governor":                   s.governor,
		"driver":                     s.driver,
		"data_source":                s.SensorSrc,
		"native_unit_of_measurement": s.UnitsString,
	}
}

func newCPUFreqSensor(info cpuFreq) *cpuFreqSensor {
	return &cpuFreqSensor{
		cpuFreq: &info,
		Sensor: linux.Sensor{
			UnitsString:      cpuFreqUnits,
			IconString:       cpuFreqIcon,
			SensorSrc:        linux.DataSrcSysfs,
			DeviceClassValue: types.DeviceClassFrequency,
			StateClassValue:  types.StateClassMeasurement,
			IsDiagnostic:     true,
			SensorTypeValue:  linux.SensorCPUFreq,
			Value:            info.freq,
		},
	}
}

func getCPUFreqs(path string) ([]cpuFreq, error) {
	matches, err := filepath.Glob(path)
	if err != nil {
		return nil, fmt.Errorf("could not read frequency files: %w", err)
	}

	freqDetails := make([]cpuFreq, 0, len(matches))

	for _, file := range matches {
		// Extract an id for this cpu.
		id, _ := strings.CutPrefix(file, sysFSPath)
		id, _ = strings.CutSuffix(id, "/"+freqFile)

		// Read the frequency value.
		freq, err := os.ReadFile(file)
		if err != nil {
			slog.Warn("Could not read frequency for cpu.", slog.String("cpu", id), slog.Any("error", err))
		}

		freq = bytes.TrimSpace(freq)

		// Read the current scaling governor.
		governor, err := os.ReadFile(filepath.Join(sysFSPath, id, governorFile))
		if err != nil {
			slog.Warn("Could not read scaling governor for cpu.", slog.String("cpu", id), slog.Any("error", err))
		}

		governor = bytes.TrimSpace(governor)

		// Read the current scaling driver.
		driver, err := os.ReadFile(filepath.Join(sysFSPath, id, driverFile))
		if err != nil {
			slog.Warn("Could not read scaling driver for cpu.", slog.String("cpu", id), slog.Any("error", err))
		}

		driver = bytes.TrimSpace(driver)

		freqDetails = append(freqDetails, cpuFreq{cpu: id, freq: string(freq), governor: string(governor), driver: string(driver)})
	}

	return freqDetails, nil
}

func NewCPUFreqWorker(_ context.Context, _ *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &cpuFreqWorker{
				path: filepath.Join(sysFSPath, "cpu*", freqFile),
			},
			WorkerID: cpuFreqWorkerID,
		},
		nil
}
