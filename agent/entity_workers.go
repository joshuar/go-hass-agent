// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package agent

import (
	"context"
	"log/slog"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
)

// CreateDeviceworkers.EntityWorkers sets up all device-specific entity
func CreateDeviceEntityWorkers(ctx context.Context, restAPIURL string) []workers.EntityWorker {
	var deviceWorkers []workers.EntityWorker

	// Initialize and add connection latency sensor w.
	if w, err := workers.NewConnectionLatencyWorker(ctx, restAPIURL); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not set up worker.",
			slog.Any("error", err))
	} else {
		deviceWorkers = append(deviceWorkers, w)
	}
	// Initialize and add external IP address sensor workezr.
	if w, err := workers.NewExternalIPWorker(ctx); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not init agent worker.",
			slog.Any("error", err))
	} else {
		deviceWorkers = append(deviceWorkers, w)
	}
	// Initialize and add external version sensor w.
	if w, err := workers.NewVersionWorker(ctx); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not init agent worker.",
			slog.Any("error", err))
	} else {
		deviceWorkers = append(deviceWorkers, w)
	}

	// Initialize and add scripts w.
	if w, err := workers.NewScriptsWorker(ctx); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not init agent worker.",
			slog.Any("error", err))
	} else {
		deviceWorkers = append(deviceWorkers, w)
	}

	return deviceWorkers
}
