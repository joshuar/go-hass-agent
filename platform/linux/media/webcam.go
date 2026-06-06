// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/pipewire"
	"github.com/joshuar/go-hass-agent/pkg/linux/webcam"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*webcamUsageWorker)(nil)

type webcamUsageWorker struct {
	*models.WorkerMetadata

	prefs       *workers.CommonWorkerPrefs
	pwEventChan chan pipewire.Event
	inMonitor   *webcam.Monitor
	inUse       atomic.Bool
}

func NewWebcamUsageWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &webcamUsageWorker{
		WorkerMetadata: models.SetWorkerMetadata("webcam_in_use", "Webcam in use"),
	}

	// Get worker preferences.
	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(prefPrefix+"webcam_in_use", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	// Set up pipewire listener.
	pwMonitor, found := linux.CtxGetPipewireMonitor(ctx)
	if !found {
		return worker, errors.New("no pipewire monitor in context")
	}
	worker.pwEventChan = pwMonitor.AddListener(webcamPipewireEventFilter)

	// Set up inotify listener.
	worker.inMonitor = webcam.NewMonitor()

	return worker, nil
}

func (w *webcamUsageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	outCh := make(chan models.Entity)

	// Monitor webcam events through pipewire.
	go func() {
		defer close(outCh)

		for event := range w.pwEventChan {
			w.parsePWState(*event.Info.State)
			outCh <- sensor.NewSensor(ctx,
				sensor.WithName("Webcam In Use"),
				sensor.WithID("webcam_in_use"),
				sensor.AsTypeBinarySensor(),
				sensor.WithIcon(webcamUseIcon(w.inUse.Load())),
				sensor.WithState(w.inUse.Load()),
				sensor.WithDataSourceAttribute(linux.DataSrcSysFS),
			)
		}
	}()

	// Monitor webcam events through inotify/device files.
	go func() {
		for event := range w.inMonitor.Run(ctx) {
			w.parseInotifyEvent(event)
			outCh <- sensor.NewSensor(ctx,
				sensor.WithName("Webcam In Use"),
				sensor.WithID("webcam_in_use"),
				sensor.AsTypeBinarySensor(),
				sensor.WithIcon(webcamUseIcon(w.inUse.Load())),
				sensor.WithState(w.inUse.Load()),
				sensor.WithDataSourceAttribute(linux.DataSrcSysFS),
			)
		}
	}()

	return outCh, nil
}

func (w *webcamUsageWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

// parsePWState parses a pipewire state value into the appropriate boolean value.
func (w *webcamUsageWorker) parsePWState(state pipewire.State) {
	switch state {
	case pipewire.StateRunning, pipewire.StateIdle:
		w.inUse.Store(true)
	case pipewire.StateSuspended:
		fallthrough
	default:
		w.inUse.Store(false)
	}
}

func (w *webcamUsageWorker) parseInotifyEvent(event webcam.Event) {
	switch event.Status {
	case webcam.StatusActive:
		w.inUse.Store(true)
	case webcam.StatusIdle:
		fallthrough
	default:
		w.inUse.Store(false)
	}
}

// webcamPipewireEventFilter filters the pipewire events. For webcam monitoring, we are only
// interested in events of type EventNode that have the "media.class" property
// of "Video/Source".
func webcamPipewireEventFilter(e *pipewire.Event) bool {
	if e.Type == pipewire.InterfaceNodeEvent || e.IsRemovalEvent() {
		// Parse props.
		props, err := e.NodeProps()
		if err != nil {
			return false
		}
		return props.MediaClass == pipewire.MediaVideoSource
	}

	return false
}

func webcamUseIcon(value bool) string {
	switch value {
	case true:
		return "mdi:webcam"
	default:
		return "mdi:webcam-off"
	}
}
