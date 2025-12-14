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

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	lastBootWorkerPrefID = infoWorkerPreferencesID
)

var _ workers.EntityWorker = (*lastBootWorker)(nil)

type lastBootWorker struct {
	*models.WorkerMetadata

	lastBoot time.Time
	OutCh    chan models.Entity
	prefs    *workers.CommonWorkerPrefs
}

func NewLastBootWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &lastBootWorker{
		WorkerMetadata: models.SetWorkerMetadata("last_boottime", "Last boot time"),
	}

	var found bool

	worker.lastBoot, found = linux.CtxGetBoottime(ctx)
	if !found {
		return worker, errors.New("no last boot info in context")
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(lastBootWorkerPrefID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
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

func (w *lastBootWorker) Execute(ctx context.Context) error {
	w.OutCh <- sensor.NewSensor(ctx,
		sensor.WithName("Last Reboot"),
		sensor.WithID("last_reboot"),
		sensor.AsDiagnostic(),
		sensor.WithDeviceClass(class.SensorClassTimestamp),
		sensor.WithIcon("mdi:restart"),
		sensor.WithState(w.lastBoot.Format(time.RFC3339)),
		sensor.WithDataSourceAttribute(linux.ProcFSRoot),
	)
	return nil
}

func (w *lastBootWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
