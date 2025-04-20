// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	lastBootWorkerID     = "boot_time_sensor"
	lastBootWorkerDesc   = "Last boot time"
	lastBootWorkerPrefID = infoWorkerPreferencesID
)

var _ workers.EntityWorker = (*lastBootWorker)(nil)

var ErrInitLastBootWorker = errors.New("could not init last boot worker")

type lastBootWorker struct {
	lastBoot time.Time
	OutCh    chan models.Entity
	prefs    *preferences.CommonWorkerPrefs
	*models.WorkerMetadata
}

func (w *lastBootWorker) PreferencesID() string {
	return lastBootWorkerPrefID
}

func (w *lastBootWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *lastBootWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *lastBootWorker) Execute(ctx context.Context) error {
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
		return fmt.Errorf("could not generate last boot sensor: %w", err)
	}

	w.OutCh <- entity

	return nil
}

func (w *lastBootWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
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

func NewLastBootWorker(ctx context.Context) (workers.EntityWorker, error) {
	lastBoot, found := linux.CtxGetBoottime(ctx)
	if !found {
		return nil, errors.Join(ErrInitLastBootWorker,
			fmt.Errorf("%w: no lastBoot value", linux.ErrInvalidCtx))
	}

	worker := &lastBootWorker{
		WorkerMetadata: models.SetWorkerMetadata(lastBootWorkerID, lastBootWorkerDesc),
		lastBoot:       lastBoot,
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitLastBootWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}
