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

var _ workers.EntityWorker = (*webcamUsageWorker)(nil)

var (
	ErrInitWebcamUsageWorker = errors.New("could not init webcam usage worker")
	ErrNewWebcamUsageSensor  = errors.New("could not create webcam usage sensor")
)

const (
	webcamUsageWorkerID   = "webcam_usage_sensor"
	webcamUsageWorkerDesc = "Webcam usage detection"
)

type webcamUsageWorker struct {
	prefs       *workers.CommonWorkerPrefs
	pwEventChan chan pwmonitor.Event
	inUse       bool
	*models.WorkerMetadata
}

func NewWebcamUsageWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &webcamUsageWorker{
		WorkerMetadata: models.SetWorkerMetadata(webcamUsageWorkerID, webcamUsageWorkerDesc),
	}

	// Get worker preferences.
	defaultPrefs := &workers.CommonWorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(webcamUsagePrefID, defaultPrefs)
	if err != nil {
		return nil, errors.Join(ErrInitWebcamUsageWorker, err)
	}

	// Set up pipewire listener.
	monitor, found := linux.CtxGetPipewireMonitor(ctx)
	if !found {
		return nil, fmt.Errorf("%w: no pipewire monitor in context", ErrInitMicUsageWorker)
	}
	worker.pwEventChan = monitor.AddListener(ctx, webcamPipewireEventFilter)

	return worker, nil
}

func (w *webcamUsageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	outCh := make(chan models.Entity)
	go func() {
		defer close(outCh)

		for event := range w.pwEventChan {
			w.parsePWState(*event.Info.State)
			outCh <- sensor.NewSensor(ctx,
				sensor.WithName("Webcam In Use"),
				sensor.WithID("webcam_in_use"),
				sensor.AsTypeBinarySensor(),
				sensor.WithIcon(webcamUseIcon(w.inUse)),
				sensor.WithState(w.inUse),
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
func (w *webcamUsageWorker) parsePWState(state pwmonitor.State) {
	switch state {
	case pwmonitor.StateRunning, pwmonitor.StateIdle:
		w.inUse = true
	case pwmonitor.StateSuspended:
		fallthrough
	default:
		w.inUse = false
	}
}

// webcamPipewireEventFilter filters the pipewire events. For webcam monitoring, we are only
// interested in events of type EventNode that have the "media.class" property
// of "Video/Source".
func webcamPipewireEventFilter(e *pwmonitor.Event) bool {
	if e.Type == pwmonitor.EventNode || e.IsRemovalEvent() {
		// Parse props.
		props, err := e.NodeProps()
		if err != nil {
			return false
		}
		// Filter for v4l2 node events.
		if props.MediaClass == "Video/Source" {
			return true
		}
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
