// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	cpuVulnWorkerID      = "cpu_vulnerabilities"
	cpuVulnPreferencesID = cpuVulnWorkerID
	cpuVulnPath          = "devices/system/cpu/vulnerabilities"
)

var (
	ErrNewVulnSensor  = errors.New("could not create vulnerabilities sensor")
	ErrInitVulnWorker = errors.New("could not init vulnerabilities worker")
)

type cpuVulnWorker struct {
	path string
}

func (w *cpuVulnWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	var (
		cpuVulnerabilitiesFound bool
		err                     error
	)

	vulnerabilities, err := filepath.Glob(w.path + "/*")
	if err != nil {
		return nil, errors.Join(ErrNewVulnSensor, err)
	}

	attrs := make(map[string]any)

	for _, vulnerability := range vulnerabilities {
		var data []byte

		data, err = os.ReadFile(vulnerability)
		if err != nil {
			continue
		}

		name := filepath.Base(vulnerability)
		details := strings.TrimSpace(string(data))

		if strings.Contains(details, "Vulnerable") {
			cpuVulnerabilitiesFound = true
		}

		attrs[name] = details
	}

	cpuVulnSensor, err := sensor.NewSensor(ctx,
		sensor.WithName("CPU Vulnerabilities"),
		sensor.WithID("cpu_vulnerabilities"),
		sensor.AsTypeBinarySensor(),
		sensor.WithDeviceClass(class.BinaryClassProblem),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:security"),
		sensor.WithState(cpuVulnerabilitiesFound),
		sensor.WithAttributes(attrs),
	)
	if err != nil {
		return nil, errors.Join(ErrNewVulnSensor, err)
	}

	return []models.Entity{cpuVulnSensor}, nil
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
		return nil, errors.Join(ErrInitVulnWorker, err)
	}

	//nolint:nilnil
	if prefs.Disabled {
		return nil, nil
	}

	worker := linux.NewOneShotSensorWorker(cpuVulnWorkerID)
	worker.OneShotSensorType = cpuVulnWorker

	return worker, nil
}
