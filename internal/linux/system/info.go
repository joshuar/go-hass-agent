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
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	infoWorkerID            = "system_info"
	infoWorkerPreferencesID = sensorsPrefPrefix + "info_sensors"
)

var ErrInitInfoWorker = errors.New("could not init system info worker")

type infoWorker struct{}

func (w *infoWorker) PreferencesID() string {
	return infoWorkerPreferencesID
}

func (w *infoWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *infoWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	var sensors []models.Entity
	var errs error

	// Get distribution name and version.
	distro, version, err := device.GetOSDetails()
	if err != nil {
		logging.FromContext(ctx).
			With(slog.String("worker", infoWorkerID)).
			Warn("Could not retrieve distro details.", slog.Any("error", err))
	} else {
		var (
			entity models.Entity
			err    error
		)

		entity, err = sensor.NewSensor(ctx,
			sensor.WithName("Distribution Name"),
			sensor.WithID("distribution_name"),
			sensor.AsDiagnostic(),
			sensor.WithIcon("mdi:linux"),
			sensor.WithState(distro),
			sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
		)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("could not generate distribution name sensor: %w", err))
		} else {
			sensors = append(sensors, entity)
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
			errs = errors.Join(errs, fmt.Errorf("could not generate distribution version sensor: %w", err))
		} else {
			sensors = append(sensors, entity)
		}

	}

	// Get kernel version.
	kernelVersion, err := device.GetKernelVersion()
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
			errs = errors.Join(errs, fmt.Errorf("could not generate kernel version sensor: %w", err))
		} else {
			sensors = append(sensors, entity)
		}
	}

	return sensors, errs
}

func NewInfoWorker(_ context.Context) (*linux.OneShotSensorWorker, error) {
	infoWorker := &infoWorker{}

	prefs, err := preferences.LoadWorker(infoWorker)
	if err != nil {
		return nil, errors.Join(ErrInitInfoWorker, err)
	}

	//nolint:nilnil
	if prefs.IsDisabled() {
		return nil, nil
	}

	worker := linux.NewOneShotSensorWorker(infoWorkerID)
	worker.OneShotSensorType = infoWorker

	return worker, nil
}
