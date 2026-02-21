// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/pipewire"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*micUsageWorker)(nil)

type micUsageWorker struct {
	*models.WorkerMetadata

	prefs       *workers.CommonWorkerPrefs
	pwEventChan chan pipewire.Event
	inUse       bool
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
	worker.pwEventChan = monitor.AddListener(micPipewireEventFilter)

	return worker, nil
}

func (w *micUsageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	outCh := make(chan models.Entity)

	go func() {
		defer close(outCh)

		for event := range w.pwEventChan {
			w.parsePWState(*event.Info.State)
			outCh <- sensor.NewSensor(ctx,
				sensor.WithName("Microphone In Use"),
				sensor.WithID("microphone_in_use"),
				sensor.AsTypeBinarySensor(),
				sensor.WithIcon(micUseIcon(w.inUse)),
				sensor.WithState(w.inUse),
				sensor.WithDataSourceAttribute(linux.DataSrcSysFS),
			)
		}
	}()

	return outCh, nil
}

func (w *micUsageWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

// parsePWState parses a pipewire state value into the appropriate boolean value.
func (w *micUsageWorker) parsePWState(state pipewire.State) {
	switch state {
	case pipewire.StateRunning:
		w.inUse = true
	case pipewire.StateIdle, pipewire.StateSuspended:
		fallthrough
	default:
		w.inUse = false
	}
}

// micPipewireEventFilter filters the pipewire events. For mic monitoring, we are only
// interested in events of type EventNode that have the audio source media type.
func micPipewireEventFilter(e *pipewire.Event) bool {
	if e.Type == pipewire.EventNode || e.IsRemovalEvent() {
		// Parse props.
		props, err := e.NodeProps()
		if err != nil {
			return false
		}
		// Filter for audio stream events.
		return props.MediaClass == pipewire.MediaAudioSource
	}

	return false
}

func micUseIcon(value bool) string {
	switch value {
	case true:
		return "mdi:microphone"
	default:
		return "mdi:microphone-off"
	}
}
