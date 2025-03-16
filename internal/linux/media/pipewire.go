// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"

	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
)

// monitorPipewire starts a listener for pipewire events, filters events by the
// given eventFilter function and returns the filtered events on the channel.
func monitorPipewire(ctx context.Context, filterFunc func(*pwmonitor.Event) bool) (chan pwmonitor.Event, error) {
	// Set up pw-dump command.
	cmd := exec.CommandContext(ctx, "pw-dump", "--monitor", "--no-colors")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error starting pw-dump: %w", err)
	}
	// Start pw-dump.
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting pw-dump: %w", err)
	}
	// Decode pw-dump stdout as json stream.
	outCh := make(chan pwmonitor.Event)
	dec := json.NewDecoder(stdout)
	go func() {
		for {
			_, err := dec.Token()
			if err != nil && !errors.Is(err, io.ErrClosedPipe) {
				logging.FromContext(ctx).Debug("pw-dump: failed to read JSON token.",
					slog.Any("error", err))
			}

			// Read pw-dump output.
			for dec.More() {
				var event pwmonitor.Event
				if err := dec.Decode(&event); err == io.EOF {
					break
				} else if err != nil {
					logging.FromContext(ctx).Debug("Error decoding pw-dump output.",
						slog.Any("error", err))
				}
				if filterFunc(&event) {
					outCh <- event
				}
			}

			_, err = dec.Token()
			if err != nil && !errors.Is(err, io.ErrClosedPipe) {
				logging.FromContext(ctx).Debug("pw-dump: failed to read JSON token.",
					slog.Any("error", err))
			}
		}
	}()

	go func() {
		defer close(outCh)
		if err := cmd.Wait(); err != nil {
			logging.FromContext(ctx).Debug("Error running pw-dump.",
				slog.Any("error", err))
		}
	}()

	return outCh, nil
}
