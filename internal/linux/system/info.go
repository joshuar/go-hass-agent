// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/device/info"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	infoWorkerID            = "system_info"
	infoWorkerDesc          = "General system information"
	infoWorkerPreferencesID = sensorsPrefPrefix + "info_sensors"
)

var _ workers.EntityWorker = (*infoWorker)(nil)

var (
	ErrNewInfoSensor  = errors.New("could not create info sensor")
	ErrInitInfoWorker = errors.New("could not init system info worker")
)

type infoWorker struct {
	OutCh chan models.Entity
	prefs *preferences.CommonWorkerPrefs
	*models.WorkerMetadata
}

func (w *infoWorker) PreferencesID() string {
	return infoWorkerPreferencesID
}

func (w *infoWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *infoWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *infoWorker) Execute(ctx context.Context) error {
	var warnings error

	// Get distribution name and version.
	distro, version, err := info.GetOSDetails()
	if err != nil {
		logging.FromContext(ctx).
			With(slog.String("worker", infoWorkerID)).
			Warn("Could not retrieve distro details.", slog.Any("error", err))
	} else {
		var entity models.Entity

		entity, err = sensor.NewSensor(ctx,
			sensor.WithName("Distribution Name"),
			sensor.WithID("distribution_name"),
			sensor.AsDiagnostic(),
			sensor.WithIcon("mdi:linux"),
			sensor.WithState(distro),
			sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
		)
		if err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not generate distribution name sensor: %w", err))
		} else {
			w.OutCh <- entity
		}

		entity, err = sensor.NewSensor(ctx,
			sensor.WithName("Distribution Version"),
			sensor.WithID("distribution_version"),
			sensor.AsDiagnostic(),
			sensor.WithIcon("mdi:numeric"),
			sensor.WithState(version),
			sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
		)
		if err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not generate distribution version sensor: %w", err))
		} else {
			w.OutCh <- entity
		}
	}

	// Get kernel version.
	kernelVersion, err := info.GetKernelVersion()
	if err != nil {
		logging.FromContext(ctx).
			With(slog.String("worker", infoWorkerID)).
			Warn("Could not retrieve kernel version.", slog.Any("error", err))
	} else {
		entity, err := sensor.NewSensor(ctx,
			sensor.WithName("Kernel Version"),
			sensor.WithID("kernel_version"),
			sensor.AsDiagnostic(),
			sensor.WithIcon("mdi:chip"),
			sensor.WithState(kernelVersion),
			sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
		)
		if err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not generate kernel version sensor: %w", err))
		} else {
			w.OutCh <- entity
		}
	}

	return warnings
}

func (w *infoWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	go func() {
		defer close(w.OutCh)
		if err := w.Execute(ctx); err != nil {
			logging.FromContext(ctx).Warn("Failed to send info details",
				slog.Any("error", err))
		}
	}()
	return w.OutCh, nil
}

func NewInfoWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &infoWorker{
		WorkerMetadata: models.SetWorkerMetadata(infoWorkerID, infoWorkerDesc),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitInfoWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
