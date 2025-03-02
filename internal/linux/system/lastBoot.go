// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	lastBootWorkerID     = "boot_time_sensor"
	lastBootWorkerPrefID = infoWorkerPreferencesID
)

var ErrInitLastBootWorker = errors.New("could not init last boot worker")

type lastBootWorker struct {
	lastBoot time.Time
}

func (w *lastBootWorker) PreferencesID() string {
	return lastBootWorkerPrefID
}

func (w *lastBootWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *lastBootWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	entity, err := sensor.NewSensor(ctx,
		sensor.WithName("Last Reboot"),
		sensor.WithID("last_reboot"),
		sensor.AsDiagnostic(),
		sensor.WithDeviceClass(class.SensorClassTimestamp),
		sensor.WithIcon("mdi:restart"),
		sensor.WithState(w.lastBoot.Format(time.RFC3339)),
		sensor.WithDataSourceAttribute(linux.ProcFSRoot),
	)
	if err != nil {
		return nil, fmt.Errorf("could not generate last boot sensor: %w", err)
	}

	return []models.Entity{entity}, nil
}

func NewLastBootWorker(ctx context.Context) (*linux.OneShotSensorWorker, error) {
	lastBoot, found := linux.CtxGetBoottime(ctx)
	if !found {
		return nil, errors.Join(ErrInitLastBootWorker,
			fmt.Errorf("%w: no lastBoot value", linux.ErrInvalidCtx))
	}

	bootWorker := &lastBootWorker{lastBoot: lastBoot}

	prefs, err := preferences.LoadWorker(bootWorker)
	if err != nil {
		return nil, errors.Join(ErrInitLastBootWorker, err)
	}

	//nolint:nilnil
	if prefs.IsDisabled() {
		return nil, nil
	}

	worker := linux.NewOneShotSensorWorker(lastBootWorkerID)
	worker.OneShotSensorType = bootWorker

	return worker, nil
}
