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
	"os"
	"os/exec"
	"slices"
	"sync"
	"time"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/logging"
)

// Monitor handles monitoring pipewire for events and dispatching events to registered listeners as appropriate.
type Monitor struct {
	listeners []*Listener
	stopCh    chan struct{}
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
		stopCh:    make(chan struct{}),
	}

	// Decode pw-dump stdout as json stream.
	dec := json.NewDecoder(stdout)
	go func() {
		for {
			select {
			case <-monitor.stopCh:
				return
			default:
				for {
					_, err := dec.Token()
					if err != nil && errors.Is(err, os.ErrClosed) {
						// Stop decoding if the command stdout is closed.
						return
					}
					if err != nil && !errors.Is(err, io.ErrClosedPipe) {
						// Report any errors in decoding.
						slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "pw-dump: failed to read JSON token.",
							slog.Any("error", err))
					}

					// Read pw-dump output.
					for dec.More() {
						event, ok := eventPool.Get().(*Event)
						if !ok {
							slogctx.FromCtx(ctx).Warn("Unable to allocate event buffer.")
							continue
						}
						func(e *Event) {
							defer func() {
								e = &Event{}
								eventPool.Put(e)
							}()
							if err = dec.Decode(e); err == io.EOF {
								return
							} else if err != nil {
								slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "Error decoding pw-dump output.",
									slog.Any("error", err))
							} else if event == nil {
								return
							}
							event.CapturedAt = time.Now().UTC()
							// Filter the event through all listeners and send the event to whichever listeners want it.
							for listener := range slices.Values(monitor.listeners) {
								if listener.filterFunc(event) {
									go func() {
										listener.eventCh <- *event
									}()
								}
							}
						}(event)
					}

					_, err = dec.Token()
					if err != nil && !errors.Is(err, io.ErrClosedPipe) {
						// Report any errors in decoding.
						slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "pw-dump: failed to read JSON token.",
							slog.Any("error", err))
					}
				}
			}
		}
	}()

	// Wait for pw-dump to finish and log any errors.
	go func() {
		if err = cmd.Wait(); err != nil {
			slogctx.FromCtx(ctx).Warn("pw-dump exited with error.",
				slog.Any("err", err))
		}
		close(monitor.stopCh)
		slogctx.FromCtx(ctx).Debug("Stopped monitoring pipewire for events.")
	}()

	// Close all monitors when we the context is cancelled.
	go func() {
		<-ctx.Done()
		for listener := range slices.Values(monitor.listeners) {
			close(listener.eventCh)
		}
	}()

	slogctx.FromCtx(ctx).Debug("Monitoring pipewire for events.")

	return monitor, nil
}

func (m *Monitor) AddListener(filterFunc func(*Event) bool) chan Event {
	eventCh := make(chan Event)
	m.listeners = append(m.listeners, &Listener{
		filterFunc: filterFunc,
		eventCh:    eventCh,
	})
	return eventCh
}

// Listener contains the data for goroutine that wants to listen for pipewire events.
type Listener struct {
	filterFunc func(*Event) bool
	eventCh    chan Event
}

var eventPool = sync.Pool{
	New: func() any {
		var buf Event
		return &buf
	},
}
