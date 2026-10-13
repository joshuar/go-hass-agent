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
	sources     *pipewire.NodeTracker
	// deviceInUse is the usage reported by the device file (inotify) monitor.
	deviceInUse atomic.Bool
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
	// Track every video source so that one source changing state does not
	// mask another that is still in use.
	worker.sources = pipewire.NewNodeTracker(pipewire.MediaVideoSource)
	worker.pwEventChan = pwMonitor.AddListener(worker.sources.Filter)

	// Set up inotify listener.
	worker.inMonitor = webcam.NewMonitor()

	return worker, nil
}

func (w *webcamUsageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	outCh := make(chan models.Entity)

	// Monitor webcam events through pipewire.
	go func() {
		defer close(outCh)

		for range w.pwEventChan {
			outCh <- w.newSensor(ctx)
		}
	}()

	// Monitor webcam events through inotify/device files.
	go func() {
		for event := range w.inMonitor.Run(ctx) {
			w.parseInotifyEvent(event)
			outCh <- w.newSensor(ctx)
		}
	}()

	return outCh, nil
}

func (w *webcamUsageWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

// newSensor reports the webcam as in use when either pipewire or the device
// file monitor sees it in use.
func (w *webcamUsageWorker) newSensor(ctx context.Context) models.Entity {
	inUse := w.sources.AnyIn(pipewire.StateRunning, pipewire.StateIdle) || w.deviceInUse.Load()

	return models.NewSensor(ctx,
		models.WithName("Webcam In Use"),
		models.WithID("webcam_in_use"),
		models.AsTypeBinarySensor(),
		models.WithIcon(webcamUseIcon(inUse)),
		models.WithState(inUse),
		models.WithDataSourceAttribute(linux.DataSrcSysFS),
	)
}

func (w *webcamUsageWorker) parseInotifyEvent(event webcam.Event) {
	switch event.Status {
	case webcam.StatusActive:
		w.deviceInUse.Store(true)
	case webcam.StatusIdle:
		fallthrough
	default:
		w.deviceInUse.Store(false)
	}
}

func webcamUseIcon(value bool) string {
	switch value {
	case true:
		return "mdi:webcam"
	default:
		return "mdi:webcam-off"
	}
}
