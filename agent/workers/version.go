// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package workers

import (
	"context"
	"errors"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/sensor"
)

const (
	versionWorkerID   = "agent_version"
	versionWorkerDesc = "Go Hass Agent version"
)

var _ EntityWorker = (*Version)(nil)

var ErrVersion = errors.New("version worker error")

type Version struct {
	*models.WorkerMetadata

	prefs *CommonWorkerPrefs
}

func (w *Version) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *Version) Start(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)

	go func() {
		defer close(sensorCh)

		sensorCh <- sensor.NewSensor(ctx,
			sensor.WithName("Go Hass Agent Version"),
			sensor.WithID("agent_version"),
			sensor.AsDiagnostic(),
			sensor.WithIcon("mdi:face-agent"),
			sensor.WithState(config.AppVersion),
		)
	}()

	return sensorCh, nil
}

func NewVersionWorker(_ context.Context) (EntityWorker, error) {
	worker := &Version{
		WorkerMetadata: models.SetWorkerMetadata(versionWorkerID, versionWorkerDesc),
	}

	defaultPrefs := &CommonWorkerPrefs{}
	var err error
	worker.prefs, err = LoadWorkerPreferences("sensors.agent.version", defaultPrefs)
	if err != nil {
		return nil, errors.Join(ErrVersion, err)
	}

	return worker, nil
}
