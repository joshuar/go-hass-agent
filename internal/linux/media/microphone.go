// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"errors"
	"fmt"

	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

var _ workers.EntityWorker = (*micUsageWorker)(nil)

var (
	ErrInitMicUsageWorker = errors.New("could not init mic usage worker")
	ErrNewMicUsageSensor  = errors.New("could not create mic usage sensor")
)

const (
	micUsageWorkerID   = "microphone_usage_sensor"
	micUsageWorkerDesc = "Microphone usage detection"
)

type micUsageWorker struct {
	prefs       *preferences.CommonWorkerPrefs
	pwEventChan chan pwmonitor.Event
	inUse       bool
	*models.WorkerMetadata
}

func NewMicUsageWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &micUsageWorker{
		WorkerMetadata: models.SetWorkerMetadata(micUsageWorkerID, micUsageWorkerDesc),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitMicUsageWorker, err)
	}
	worker.prefs = prefs

	monitor, found := linux.CtxGetPipewireMonitor(ctx)
	if !found {
		return nil, fmt.Errorf("%w: no pipewire monitor in context", ErrInitMicUsageWorker)
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

func (w *micUsageWorker) PreferencesID() string {
	return micUsagePrefID
}

func (w *micUsageWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
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
