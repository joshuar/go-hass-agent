// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/pipewire"
	"github.com/joshuar/go-hass-agent/pkg/linux/webcam"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*webcamUsageWorker)(nil)

// deviceUsageTracker keeps per-device webcam usage from the V4L2/inotify
// monitor. A single boolean is wrong on multi-node cameras: an IDLE event from
// one /dev/video* node would clear "in use" while another node is still active.
type deviceUsageTracker struct {
	mu      sync.Mutex
	devices map[string]bool
}

func newDeviceUsageTracker() *deviceUsageTracker {
	return &deviceUsageTracker{devices: make(map[string]bool)}
}

func (t *deviceUsageTracker) set(device string, inUse bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if device == "" {
		return
	}
	if inUse {
		t.devices[device] = true
		return
	}
	delete(t.devices, device)
}

func (t *deviceUsageTracker) any() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return len(t.devices) > 0
}

type webcamUsageWorker struct {
	*models.WorkerMetadata

	prefs       *workers.CommonWorkerPrefs
	pwEventChan chan pipewire.Event
	inMonitor   *webcam.Monitor
	sources     *pipewire.NodeTracker
	devices     *deviceUsageTracker

	// lastSent guards duplicate Home Assistant updates when underlying
	// PipeWire/V4L2 events leave the aggregated in-use boolean unchanged.
	lastSentMu sync.Mutex
	lastSent   *bool
}

func NewWebcamUsageWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &webcamUsageWorker{
		WorkerMetadata: models.SetWorkerMetadata("webcam_in_use", "Webcam in use"),
		devices:        newDeviceUsageTracker(),
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
			w.emitIfChanged(ctx, outCh)
		}
	}()

	// Monitor webcam events through inotify/device files.
	go func() {
		for event := range w.inMonitor.Run(ctx) {
			w.parseInotifyEvent(event)
			w.emitIfChanged(ctx, outCh)
		}
	}()

	return outCh, nil
}

func (w *webcamUsageWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

// currentInUse reports the webcam as in use when either pipewire or the device
// file monitor sees it in use.
func (w *webcamUsageWorker) currentInUse() bool {
	// Match microphone: only treat PipeWire "running" as in use. "idle" is
	// common for unused camera nodes and caused false positives / flapping.
	return w.sources.AnyIn(pipewire.StateRunning) || w.devices.any()
}

// emitIfChanged sends a sensor update only when the aggregated in-use state
// changes. Duplicate events with the same boolean are dropped so Home Assistant
// is not spammed with unchanged webhook updates.
func (w *webcamUsageWorker) emitIfChanged(ctx context.Context, outCh chan<- models.Entity) {
	inUse := w.currentInUse()

	w.lastSentMu.Lock()
	if w.lastSent != nil && *w.lastSent == inUse {
		w.lastSentMu.Unlock()
		return
	}
	last := inUse
	w.lastSent = &last
	w.lastSentMu.Unlock()

	outCh <- models.NewSensor(ctx,
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
		w.devices.set(event.Device, true)
	case webcam.StatusIdle:
		fallthrough
	default:
		w.devices.set(event.Device, false)
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
