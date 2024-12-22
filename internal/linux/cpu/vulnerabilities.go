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

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	cpuVulnWorkerID = "cpu_vulnerabilities"
	cpuVulnPath     = "devices/system/cpu/vulnerabilities"
)

type cpuVulnWorker struct {
	path string
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

func NewCPUVulnerabilityWorker(_ context.Context) (*linux.OneShotSensorWorker, error) {
	worker := linux.NewOneShotSensorWorker(cpuVulnWorkerID)

	worker.OneShotSensorType = &cpuVulnWorker{
		path: filepath.Join(linux.SysFSRoot, cpuVulnPath),
	}

	return worker, nil
}
