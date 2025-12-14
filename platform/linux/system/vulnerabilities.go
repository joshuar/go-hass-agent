// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const cpuVulnPath = "devices/system/cpu/vulnerabilities"

var _ workers.EntityWorker = (*cpuVulnWorker)(nil)

type cpuVulnWorker struct {
	*models.WorkerMetadata

	path  string
	prefs *workers.CommonWorkerPrefs
	OutCh chan models.Entity
}

func NewCPUVulnerabilityWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &cpuVulnWorker{
		WorkerMetadata: models.SetWorkerMetadata("cpu_vulnerabilities", "Check CPU vulnerabilities"),
		path:           filepath.Join(linux.SysFSRoot, cpuVulnPath),
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences("cpu_vulnerabilities", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
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

func (w *cpuVulnWorker) Execute(ctx context.Context) error {
	var (
		cpuVulnerabilitiesFound bool
		err                     error
	)

	vulnerabilities, err := filepath.Glob(w.path + "/*")
	if err != nil {
		return fmt.Errorf("get vulnerabilities: %w", err)
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

	w.OutCh <- sensor.NewSensor(ctx,
		sensor.WithName("CPU Vulnerabilities"),
		sensor.WithID("cpu_vulnerabilities"),
		sensor.AsTypeBinarySensor(),
		sensor.WithDeviceClass(class.BinaryClassProblem),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:security"),
		sensor.WithState(cpuVulnerabilitiesFound),
		sensor.WithAttributes(attrs),
	)

	return nil
}

func (w *cpuVulnWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
