// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"log/slog"

	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
)

// monitorPipewire starts a listener for pipewire events, filters events by the
// given eventFilter function and returns the filtered events on the eventCh channel.
func monitorPipewire(ctx context.Context, eventCh chan []*pwmonitor.Event, eventFilter func(*pwmonitor.Event) bool) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := pwmonitor.Monitor(ctx, eventCh, eventFilter); err != nil {
				logging.FromContext(ctx).Debug("Pipewire monitor shutdown unexpectedly. Restarting.",
					slog.Any("error", err))
			}
		}
	}
}
