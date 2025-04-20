// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package device

import (
	"context"
	"log/slog"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/device/worker"
	"github.com/joshuar/go-hass-agent/internal/device/worker/scripts"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

// CreateDeviceEntityWorkers sets up all device-specific entity workers.
func CreateDeviceEntityWorkers(ctx context.Context) []workers.EntityWorker {
	var deviceWorkers []workers.EntityWorker

	// Initialize and add connection latency sensor w.
	if w, err := worker.NewConnectionLatencyWorker(ctx); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not set up worker.",
			slog.String("id", w.ID()),
			slog.Any("error", err))
	} else {
		deviceWorkers = append(deviceWorkers, w)
	}
	// Initialize and add external IP address sensor workezr.
	if w, err := worker.NewExternalIPWorker(ctx); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not init agent worker.",
			slog.String("id", w.ID()),
			slog.Any("error", err))
	} else {
		deviceWorkers = append(deviceWorkers, w)
	}
	// Initialize and add external version sensor w.
	if w, err := worker.NewVersionWorker(ctx); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not init agent worker.",
			slog.String("id", w.ID()),
			slog.Any("error", err))
	} else {
		deviceWorkers = append(deviceWorkers, w)
	}

	// Initialize and add scripts w.
	if w, err := scripts.NewScriptsWorker(ctx); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not init agent worker.",
			slog.String("id", w.ID()),
			slog.Any("error", err))
	} else {
		deviceWorkers = append(deviceWorkers, w)
	}

	return deviceWorkers
}
