// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"testing"
	"time"

	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/webcam"
)

func TestDeviceUsageTracker_multiNode(t *testing.T) {
	t.Parallel()

	tracker := newDeviceUsageTracker()
	worker := &webcamUsageWorker{devices: tracker}

	worker.parseInotifyEvent(webcam.Event{Device: "/dev/video0", Status: webcam.StatusActive})
	if !tracker.any() {
		t.Fatal("expected in use after video0 active")
	}

	// An idle event from a sibling node must not clear the overall state.
	worker.parseInotifyEvent(webcam.Event{Device: "/dev/video1", Status: webcam.StatusIdle})
	if !tracker.any() {
		t.Fatal("expected still in use while video0 is active")
	}

	worker.parseInotifyEvent(webcam.Event{Device: "/dev/video0", Status: webcam.StatusIdle})
	if tracker.any() {
		t.Fatal("expected idle after all devices idle")
	}
}

func TestWebcamDebounceDropsProbeFlapping(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outCh := make(chan models.Entity, 16)
	worker := &webcamUsageWorker{
		devices: newDeviceUsageTracker(),
		emitCtx: ctx,
		outCh:   outCh,
	}

	// Simulate a probe burst: open+close within the debounce window.
	worker.parseInotifyEvent(webcam.Event{Device: "/dev/video0", Status: webcam.StatusActive})
	worker.scheduleEmit()
	worker.parseInotifyEvent(webcam.Event{Device: "/dev/video0", Status: webcam.StatusIdle})
	worker.scheduleEmit()

	select {
	case <-outCh:
		t.Fatal("emitted during debounce window")
	case <-time.After(200 * time.Millisecond):
	}

	// After quiet period with idle state, and never having sent before, no update.
	deadline := time.After(webcamStateDebounce + 500*time.Millisecond)
	var got int
	for {
		select {
		case <-outCh:
			got++
		case <-deadline:
			if got != 0 {
				t.Fatalf("expected no HA updates for brief probe, got %d", got)
			}
			return
		}
	}
}

func TestWebcamDebounceEmitsAfterSustainedUse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outCh := make(chan models.Entity, 16)
	worker := &webcamUsageWorker{
		devices: newDeviceUsageTracker(),
		emitCtx: ctx,
		outCh:   outCh,
	}

	worker.parseInotifyEvent(webcam.Event{Device: "/dev/video0", Status: webcam.StatusActive})
	worker.scheduleEmit()

	select {
	case entity := <-outCh:
		sensor, err := entity.AsSensor()
		if err != nil {
			t.Fatalf("AsSensor: %v", err)
		}
		if sensor.State != true {
			t.Fatalf("expected true after sustained use, got %v", sensor.State)
		}
	case <-time.After(webcamStateDebounce + 500*time.Millisecond):
		t.Fatal("timed out waiting for sustained-use emit")
	}
}
