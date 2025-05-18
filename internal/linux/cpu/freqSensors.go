// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	freqFile     = "cpufreq/scaling_cur_freq"
	governorFile = "cpufreq/scaling_governor"
	driverFile   = "cpufreq/scaling_driver"

	cpuFreqIcon  = "mdi:chip"
	cpuFreqUnits = "kHz"
)

var ErrNewCPUFreqSensor = errors.New("error creating new CPU Freq sensor")

var totalCPUs = runtime.NumCPU()

type cpuFreq struct {
	cpu      string
	governor string
	driver   string
	freq     string
}

func newCPUFreqSensor(ctx context.Context, id string) models.Entity {
	info := getCPUFreqs(id)
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
	prop, err := os.ReadFile(path) // #nosec:G304
	if err != nil {
		slog.Debug("Could not read CPU freq property.",
			slog.String("cpu", id),
			slog.String("property", file),
			slog.Any("error", err))

		return "unknown"
	}

	return string(bytes.TrimSpace(prop))
}
