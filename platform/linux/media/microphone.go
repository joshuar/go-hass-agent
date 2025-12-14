// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"errors"
	"fmt"

	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*micUsageWorker)(nil)

type micUsageWorker struct {
	*models.WorkerMetadata

	prefs       *workers.CommonWorkerPrefs
	pwEventChan chan pwmonitor.Event
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
	worker.pwEventChan = monitor.AddListener(ctx, micPipewireEventFilter)

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
func (w *micUsageWorker) parsePWState(state pwmonitor.State) {
	switch state {
	case pwmonitor.StateRunning:
		w.inUse = true
	case pwmonitor.StateIdle, pwmonitor.StateSuspended:
		fallthrough
	default:
		w.inUse = false
	}
}

// micPipewireEventFilter filters the pipewire events. For mic monitoring, we are only
// interested in events of type EventNode that have the audio source media type.
func micPipewireEventFilter(e *pwmonitor.Event) bool {
	if e.Type == pwmonitor.EventNode || e.IsRemovalEvent() {
		// Parse props.
		props, err := e.NodeProps()
		if err != nil {
			return false
		}
		// Filter for audio stream events.
		return props.MediaClass == pwmonitor.MediaAudioSource
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
