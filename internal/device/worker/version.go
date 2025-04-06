// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	versionWorkerID   = "agent_version"
	versionWorkerDesc = "Go Hass Agent version"
)

var _ workers.EntityWorker = (*Version)(nil)

var ErrVersion = errors.New("version worker error")

type Version struct {
	prefs *preferences.CommonWorkerPrefs
	*models.WorkerMetadata
}

func (w *Version) PreferencesID() string {
	return preferences.SensorsPrefPrefix + "agent" + preferences.PathDelim + "version"
}

func (w *Version) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *Version) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *Version) Start(ctx context.Context) (<-chan models.Entity, error) {
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

func newVersionSensor(ctx context.Context) (models.Entity, error) {
	sensor, err := sensor.NewSensor(ctx,
		sensor.WithName("Go Hass Agent Version"),
		sensor.WithID("agent_version"),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:face-agent"),
		sensor.WithState(preferences.AppVersion()),
	)
	if err != nil {
		return sensor, fmt.Errorf("%w: %w", ErrVersion, err)
	}
	return sensor, nil
}

func NewVersionWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &Version{
		WorkerMetadata: models.SetWorkerMetadata(versionWorkerID, versionWorkerDesc),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrVersion, err)
	}

	worker.prefs = prefs

	return worker, nil
}
