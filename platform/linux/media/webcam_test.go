// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"testing"
	"time"

	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/pipewire"
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

func TestEmitIfChanged_onlyOnValueChange(t *testing.T) {
	t.Parallel()

	worker := &webcamUsageWorker{
		devices: newDeviceUsageTracker(),
		sources: pipewire.NewNodeTracker(pipewire.MediaVideoSource),
	}
	outCh := make(chan models.Entity, 8)
	ctx := context.Background()

	// First idle emission should send.
	worker.emitIfChanged(ctx, outCh)
	if got := drainCount(outCh); got != 1 {
		t.Fatalf("initial idle: want 1 emission, got %d", got)
	}

	// Duplicate idle must not send.
	worker.emitIfChanged(ctx, outCh)
	worker.emitIfChanged(ctx, outCh)
	if got := drainCount(outCh); got != 0 {
		t.Fatalf("duplicate idle: want 0 emissions, got %d", got)
	}

	// Transition to in-use should send once.
	worker.parseInotifyEvent(webcam.Event{Device: "/dev/video0", Status: webcam.StatusActive})
	worker.emitIfChanged(ctx, outCh)
	if got := drainCount(outCh); got != 1 {
		t.Fatalf("became active: want 1 emission, got %d", got)
	}
	worker.emitIfChanged(ctx, outCh)
	if got := drainCount(outCh); got != 0 {
		t.Fatalf("duplicate active: want 0 emissions, got %d", got)
	}

	// Sibling idle while still active must not send.
	worker.parseInotifyEvent(webcam.Event{Device: "/dev/video1", Status: webcam.StatusIdle})
	worker.emitIfChanged(ctx, outCh)
	if got := drainCount(outCh); got != 0 {
		t.Fatalf("sibling idle while active: want 0 emissions, got %d", got)
	}

	// Real idle transition should send once.
	worker.parseInotifyEvent(webcam.Event{Device: "/dev/video0", Status: webcam.StatusIdle})
	worker.emitIfChanged(ctx, outCh)
	if got := drainCount(outCh); got != 1 {
		t.Fatalf("became idle: want 1 emission, got %d", got)
	}
	worker.emitIfChanged(ctx, outCh)
	if got := drainCount(outCh); got != 0 {
		t.Fatalf("duplicate idle after close: want 0 emissions, got %d", got)
	}
}

func drainCount(ch <-chan models.Entity) int {
	n := 0
	for {
		select {
		case <-ch:
			n++
		case <-time.After(20 * time.Millisecond):
			return n
		}
	}
}
