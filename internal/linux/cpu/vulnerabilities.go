// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cpu

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	cpuVulnWorkerID = "cpu_vulnerabilities"
	preferencesID   = cpuVulnWorkerID
	cpuVulnPath     = "devices/system/cpu/vulnerabilities"
)

type cpuVulnWorker struct {
	path  string
	prefs *preferences.CommonWorkerPrefs
}

func (w *cpuVulnWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	var cpuVulnerabilitiesFound bool

	vulnerabilities, err := filepath.Glob(w.path + "/*")
	if err != nil {
		return nil, fmt.Errorf("could not fetch vulnerabilities from SysFS: %w", err)
	}

	attrs := make(map[string]any)

	for _, vulnerability := range vulnerabilities {
		detailsRaw, err := os.ReadFile(vulnerability)
		if err != nil {
			continue
		}

		name := filepath.Base(vulnerability)
		details := strings.TrimSpace(string(detailsRaw))

		if strings.Contains(details, "Vulnerable") {
			cpuVulnerabilitiesFound = true
		}

		attrs[name] = details
	}

	cpuVulnSensor := sensor.NewSensor(
		sensor.WithName("CPU Vulnerabilities"),
		sensor.WithID("cpu_vulnerabilities"),
		sensor.AsTypeBinarySensor(),
		sensor.WithDeviceClass(types.BinarySensorDeviceClassProblem),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon("mdi:security"),
			sensor.WithValue(cpuVulnerabilitiesFound),
			sensor.WithAttributes(attrs),
		),
	)

	return []sensor.Entity{cpuVulnSensor}, nil
}

func (w *cpuVulnWorker) PreferencesID() string {
	return preferencesID
}

func (w *cpuVulnWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewCPUVulnerabilityWorker(ctx context.Context) (*linux.OneShotSensorWorker, error) {
	var err error

	worker := linux.NewOneShotSensorWorker(cpuVulnWorkerID)

	cpuVulnWorker := &cpuVulnWorker{
		path: filepath.Join(linux.SysFSRoot, cpuVulnPath),
	}

	cpuVulnWorker.prefs, err = preferences.LoadWorker(ctx, cpuVulnWorker)
	if err != nil {
		return worker, fmt.Errorf("could not load preferences: %w", err)
	}

	// If disabled, don't use the addressWorker.
	if cpuVulnWorker.prefs.Disabled {
		return worker, nil
	}

	worker.OneShotSensorType = cpuVulnWorker

	return worker, nil
}
