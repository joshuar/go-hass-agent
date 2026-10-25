// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/pipewire"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*micUsageWorker)(nil)

type micUsageWorker struct {
	*models.WorkerMetadata

	prefs       *workers.CommonWorkerPrefs
	pwEventChan chan pipewire.Event
	sources     *pipewire.NodeTracker
}

func NewMicUsageWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &micUsageWorker{
		WorkerMetadata: models.SetWorkerMetadata("mic_in_use", "Microphone in use"),
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(prefPrefix+"microphone_in_use", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	monitor, found := linux.CtxGetPipewireMonitor(ctx)
	if !found {
		return worker, errors.New("no pipewire monitor in context")
	}
	// Track every audio source so that one source changing state does not
	// mask another that is still running.
	worker.sources = pipewire.NewNodeTracker(pipewire.MediaAudioSource)
	worker.pwEventChan = monitor.AddListener(worker.sources.Filter)

	return worker, nil
}

func (w *micUsageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	outCh := make(chan models.Entity)

	go func() {
		defer close(outCh)

		for range w.pwEventChan {
			inUse := w.sources.AnyIn(pipewire.StateRunning)
			outCh <- models.NewSensor(ctx,
				models.WithName("Microphone In Use"),
				models.WithID("microphone_in_use"),
				models.AsTypeBinarySensor(),
				models.WithIcon(micUseIcon(inUse)),
				models.WithState(inUse),
				models.WithDataSourceAttribute(linux.DataSrcSysFS),
			)
		}
	}()

	return outCh, nil
}

func (w *micUsageWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func micUseIcon(value bool) string {
	switch value {
	case true:
		return "mdi:microphone"
	default:
		return "mdi:microphone-off"
	}
}
