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
	"strings"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	freqFile     = "cpufreq/scaling_cur_freq"
	governorFile = "cpufreq/scaling_governor"
	driverFile   = "cpufreq/scaling_driver"

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
		"data_source":                s.DataSource,
		"native_unit_of_measurement": s.UnitsString,
	}
}

func newCPUFreqSensor(id string) *cpuFreqSensor {
	info := getCPUFreqs(id)

	sensor := &cpuFreqSensor{
		cpuFreq: info,
		Sensor: linux.Sensor{
			UnitsString:      cpuFreqUnits,
			IconString:       cpuFreqIcon,
			DataSource:       linux.DataSrcSysfs,
			DeviceClassValue: types.DeviceClassFrequency,
			StateClassValue:  types.StateClassMeasurement,
			IsDiagnostic:     true,
			Value:            info.freq,
		},
	}

	return sensor
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
		slog.Debug("Could not read CPU freq property.", slog.String("cpu", id), slog.String("property", file), slog.Any("error", err))

		return "unknown"
	}

	return string(bytes.TrimSpace(prop))
}
