// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
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
	cpuVulnWorkerID      = "cpu_vulnerabilities"
	cpuVulnPreferencesID = cpuVulnWorkerID
	cpuVulnPath          = "devices/system/cpu/vulnerabilities"
)

var ErrInitVulnerabilitiesWorker = errors.New("could not init vulnerabilities worker")

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

func (w *cpuVulnWorker) PreferencesID() string {
	return infoWorkerPreferencesID
}

func (w *cpuVulnWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func NewCPUVulnerabilityWorker(_ context.Context) (*linux.OneShotSensorWorker, error) {
	cpuVulnWorker := &cpuVulnWorker{
		path: filepath.Join(linux.SysFSRoot, cpuVulnPath),
	}

	prefs, err := preferences.LoadWorker(cpuVulnWorker)
	if err != nil {
		return nil, errors.Join(ErrInitVulnerabilitiesWorker, err)
	}

	//nolint:nilnil
	if prefs.Disabled {
		return nil, nil
	}

	worker := linux.NewOneShotSensorWorker(cpuVulnWorkerID)
	worker.OneShotSensorType = cpuVulnWorker

	return worker, nil
}
