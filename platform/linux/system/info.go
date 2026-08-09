// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"fmt"
	"log/slog"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/device"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	infoWorkerPreferencesID = sensorsPrefPrefix + "info_sensors"
)

var _ workers.EntityWorker = (*infoWorker)(nil)

type infoWorker struct {
	*models.WorkerMetadata

	OutCh chan models.Entity
	prefs *workers.CommonWorkerPrefs
}

func NewInfoWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &infoWorker{
		WorkerMetadata: models.SetWorkerMetadata("system_info", "System information"),
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(infoWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

func (w *infoWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	go func() {
		defer close(w.OutCh)
		if err := w.Execute(ctx); err != nil {
			slogctx.FromCtx(ctx).Warn("Failed to send info details",
				slog.Any("error", err))
		}
	}()
	return w.OutCh, nil
}

func (w *infoWorker) Execute(ctx context.Context) error {
	var warnings error

	// Get distribution name and version.
	distro, version, err := device.GetOSDetails()
	if err != nil {
		return fmt.Errorf("could not retrieve distro details: %w", err)
	}

	w.OutCh <- models.NewSensor(ctx,
		models.WithName("Distribution Name"),
		models.WithID("distribution_name"),
		models.AsDiagnostic(),
		models.WithIcon("mdi:linux"),
		models.WithState(distro),
		models.WithDataSourceAttribute(linux.DataSrcProcFS),
	)

	w.OutCh <- models.NewSensor(ctx,
		models.WithName("Distribution Version"),
		models.WithID("distribution_version"),
		models.AsDiagnostic(),
		models.WithIcon("mdi:numeric"),
		models.WithState(version),
		models.WithDataSourceAttribute(linux.DataSrcProcFS),
	)

	// Get kernel version.
	kernelVersion, err := device.GetKernelVersion()
	if err != nil {
		return fmt.Errorf("could not retrieve kernel version: %w", err)
	}

	w.OutCh <- models.NewSensor(ctx,
		models.WithName("Kernel Version"),
		models.WithID("kernel_version"),
		models.AsDiagnostic(),
		models.WithIcon("mdi:chip"),
		models.WithState(kernelVersion),
		models.WithDataSourceAttribute(linux.DataSrcProcFS),
	)

	return warnings
}

func (w *infoWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
