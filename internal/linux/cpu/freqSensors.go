// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cpu

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	freqFile     = "cpufreq/scaling_cur_freq"
	governorFile = "cpufreq/scaling_governor"
	driverFile   = "cpufreq/scaling_driver"

	cpuFreqIcon  = "mdi:chip"
	cpuFreqUnits = "kHz"
)

var totalCPUs = runtime.NumCPU()

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
