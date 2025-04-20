// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	cpuVulnWorkerID      = "cpu_vulnerabilities"
	cpuVulnWorkerDesc    = "Potential CPU vulnerabilities reported by the kernel"
	cpuVulnPreferencesID = cpuVulnWorkerID
	cpuVulnPath          = "devices/system/cpu/vulnerabilities"
)

var _ workers.EntityWorker = (*cpuVulnWorker)(nil)

var (
	ErrNewVulnSensor  = errors.New("could not create vulnerabilities sensor")
	ErrInitVulnWorker = errors.New("could not init vulnerabilities worker")
)

type cpuVulnWorker struct {
	path  string
	prefs *preferences.CommonWorkerPrefs
	OutCh chan models.Entity
	*models.WorkerMetadata
}

func (w *cpuVulnWorker) Execute(ctx context.Context) error {
	var (
		cpuVulnerabilitiesFound bool
		err                     error
	)

	vulnerabilities, err := filepath.Glob(w.path + "/*")
	if err != nil {
		return errors.Join(ErrNewVulnSensor, err)
	}

	attrs := make(map[string]any)

	for vulnerability := range slices.Values(vulnerabilities) {
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
		return errors.Join(ErrNewVulnSensor, err)
	}

	w.OutCh <- cpuVulnSensor

	return nil
}

func (w *cpuVulnWorker) PreferencesID() string {
	return infoWorkerPreferencesID
}

func (w *cpuVulnWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *cpuVulnWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *cpuVulnWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	go func() {
		defer close(w.OutCh)
		if err := w.Execute(ctx); err != nil {
			slogctx.FromCtx(ctx).Warn("Failed to send cpu vulnerability details",
				slog.Any("error", err))
		}
	}()
	return w.OutCh, nil
}

func NewCPUVulnerabilityWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &cpuVulnWorker{
		WorkerMetadata: models.SetWorkerMetadata(cpuVulnWorkerID, cpuVulnWorkerDesc),
		path:           filepath.Join(linux.SysFSRoot, cpuVulnPath),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitVulnWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
