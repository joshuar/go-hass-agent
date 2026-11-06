// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/pipewire"
	"github.com/joshuar/go-hass-agent/pkg/linux/webcam"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*webcamUsageWorker)(nil)

// webcamStateDebounce ignores brief open/close probes (PipeWire/portal
// enumeration) that would otherwise flap the Home Assistant binary sensor.
const webcamStateDebounce = time.Second

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

	// lastSent is the last state emitted to Home Assistant.
	lastSentMu sync.Mutex
	lastSent   *bool

	// Debounce: only emit after the aggregated state has been stable.
	debounceMu sync.Mutex
	pending    *bool
	timer      *time.Timer
	outCh      chan<- models.Entity
	emitCtx    context.Context
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
	w.emitCtx = ctx
	w.outCh = outCh

	// Monitor webcam events through pipewire.
	go func() {
		defer close(outCh)

		for range w.pwEventChan {
			w.scheduleEmit()
		}
	}()

	// Monitor webcam events through inotify/device files.
	go func() {
		for event := range w.inMonitor.Run(ctx) {
			w.parseInotifyEvent(event)
			w.scheduleEmit()
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
	pwInUse := w.sources != nil && w.sources.AnyIn(pipewire.StateRunning)
	return pwInUse || w.devices.any()
}

// scheduleEmit records the latest aggregated state and (re)starts a debounce
// timer. Brief probe open/close pairs settle before the timer fires and are
// never sent to Home Assistant.
func (w *webcamUsageWorker) scheduleEmit() {
	inUse := w.currentInUse()

	w.debounceMu.Lock()
	defer w.debounceMu.Unlock()

	pending := inUse
	w.pending = &pending
	if w.timer == nil {
		w.timer = time.AfterFunc(webcamStateDebounce, w.flushDebounced)
		return
	}
	w.timer.Reset(webcamStateDebounce)
}

func (w *webcamUsageWorker) flushDebounced() {
	w.debounceMu.Lock()
	pending := w.pending
	outCh := w.outCh
	ctx := w.emitCtx
	w.debounceMu.Unlock()

	if pending == nil || outCh == nil || ctx == nil {
		return
	}
	// Re-sample after the quiet period so we emit the live state, not a
	// stale intermediate value from the start of the burst.
	inUse := w.currentInUse()

	w.lastSentMu.Lock()
	if w.lastSent == nil && !inUse {
		// Default/idle with no prior send: remember idle without a webhook.
		// Brief probe bursts that settle idle must not spam Home Assistant.
		idle := false
		w.lastSent = &idle
		w.lastSentMu.Unlock()
		return
	}
	if w.lastSent != nil && *w.lastSent == inUse {
		w.lastSentMu.Unlock()
		return
	}
	last := inUse
	w.lastSent = &last
	w.lastSentMu.Unlock()

	select {
	case <-ctx.Done():
		return
	case outCh <- models.NewSensor(ctx,
		models.WithName("Webcam In Use"),
		models.WithID("webcam_in_use"),
		models.AsTypeBinarySensor(),
		models.WithIcon(webcamUseIcon(inUse)),
		models.WithState(inUse),
		models.WithDataSourceAttribute(linux.DataSrcSysFS),
	):
	}
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
