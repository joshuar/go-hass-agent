// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package agentsensor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	versionWorkerID = "agent_version"
)

var ErrInitVersionWorker = errors.New("could not init version worker")

func newVersionSensor(ctx context.Context) (models.Entity, error) {
	return sensor.NewSensor(ctx,
		sensor.WithName("Go Hass Agent Version"),
		sensor.WithID("agent_version"),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:face-agent"),
		sensor.WithState(preferences.AppVersion()),
	)
}

type VersionWorker struct {
	prefs *preferences.CommonWorkerPrefs
}

func (w *VersionWorker) PreferencesID() string {
	return preferences.SensorsPrefPrefix + "agent" + preferences.PathDelim + "version"
}

func (w *VersionWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *VersionWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *VersionWorker) ID() string { return versionWorkerID }

func (w *VersionWorker) Stop() error { return nil }

func (w *VersionWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)

	go func() {
		defer close(sensorCh)

		entity, err := newVersionSensor(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Failed to create version sensor entity.",
				slog.Any("error", err))
			return
		}

		sensorCh <- entity
	}()

	return sensorCh, nil
}

func (w *VersionWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	entity, err := newVersionSensor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create version sensor entity: %w", err)
	}

	return []models.Entity{entity}, nil
}

func NewVersionWorker(_ context.Context) (*VersionWorker, error) {
	worker := &VersionWorker{}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitVersionWorker, err)
	}

	worker.prefs = prefs

	return worker, nil
}
