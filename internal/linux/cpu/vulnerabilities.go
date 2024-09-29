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

	cpuVulnSensor := sensor.Entity{
		Name:        "CPU Vulnerabilities",
		DeviceClass: types.BinarySensorDeviceClassProblem,
		Category:    types.CategoryDiagnostic,
		State: &sensor.State{
			ID:         "cpu_vulnerabilities",
			Value:      cpuVulnerabilitiesFound,
			Icon:       "mdi:security",
			Attributes: attrs,
			EntityType: types.BinarySensor,
		},
	}

	return []sensor.Entity{cpuVulnSensor}, nil
}

func NewCPUVulnerabilityWorker(_ context.Context) (*linux.OneShotSensorWorker, error) {
	worker := linux.NewOneShotWorker(cpuVulnWorkerID)

	worker.OneShotType = &cpuVulnWorker{
		path: filepath.Join(linux.SysFSRoot, cpuVulnPath),
	}

	return worker, nil
}
