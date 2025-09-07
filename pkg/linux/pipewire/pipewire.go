// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package pipewire

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"slices"

	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/logging"
)

// Monitor handles monitoring pipewire for events and dispatching events to registered listeners as appropriate.
type Monitor struct {
	listeners []*Listener
}

// NewMonitor creates a new pipewire monitor.
//
//nolint:gocognit
func NewMonitor(ctx context.Context) (*Monitor, error) {
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
	// Create monitor
	monitor := &Monitor{
		listeners: make([]*Listener, 0),
	}

	// Decode pw-dump stdout as json stream.
	dec := json.NewDecoder(stdout)
	go func() {
		for {
			_, err := dec.Token()
			if err != nil && !errors.Is(err, io.ErrClosedPipe) {
				slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "pw-dump: failed to read JSON token.",
					slog.Any("error", err))
			}

			// Read pw-dump output.
			for dec.More() {
				var event pwmonitor.Event
				if err = dec.Decode(&event); err == io.EOF {
					break
				} else if err != nil {
					slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "Error decoding pw-dump output.",
						slog.Any("error", err))
				}
				// Filter the event through all listeners and send the event to whichever listeners want it.
				for listener := range slices.Values(monitor.listeners) {
					if listener.filterFunc(&event) {
						go func() {
							listener.eventCh <- event
						}()
					}
				}
			}

			_, err = dec.Token()
			if err != nil && !errors.Is(err, io.ErrClosedPipe) {
				slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "pw-dump: failed to read JSON token.",
					slog.Any("error", err))
			}
		}
	}()

	go func() {
		<-ctx.Done()
		for listener := range slices.Values(monitor.listeners) {
			close(listener.eventCh)
		}
	}()

	return monitor, nil
}

func (m *Monitor) AddListener(ctx context.Context, filterFunc func(*pwmonitor.Event) bool) chan pwmonitor.Event {
	eventCh := make(chan pwmonitor.Event)
	m.listeners = append(m.listeners, &Listener{
		filterFunc: filterFunc,
		eventCh:    eventCh,
	})
	return eventCh
}

// Listener contains the data for goroutine that wants to listen for pipewire events.
type Listener struct {
	filterFunc func(*pwmonitor.Event) bool
	eventCh    chan pwmonitor.Event
}
