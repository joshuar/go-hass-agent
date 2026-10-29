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

const (
	// restartMinDelay is the initial delay before restarting pw-dump after it
	// exits.
	restartMinDelay = time.Second
	// restartMaxDelay is the maximum delay between attempts to restart
	// pw-dump. A pw-dump that ran for longer than this resets the delay.
	restartMaxDelay = time.Minute
)

// Monitor handles monitoring pipewire for events and dispatching events to registered listeners as appropriate.
type Monitor struct {
	mu        sync.RWMutex
	listeners []*Listener
}

// pwDump is a running pw-dump process.
type pwDump struct {
	cmd    *exec.Cmd
	stdout io.Reader
	cancel context.CancelFunc
}

// startPWDump starts pw-dump in monitor mode.
func startPWDump(ctx context.Context) (*pwDump, error) {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, "pw-dump", "--monitor", "--no-colors")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error starting pw-dump: %w", err)
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("error starting pw-dump: %w", err)
	}

	return &pwDump{cmd: cmd, stdout: stdout, cancel: cancel}, nil
}

// NewMonitor creates a new pipewire monitor. If pw-dump exits, for example
// because pipewire was restarted, it is restarted until the context is
// cancelled.
func NewMonitor(ctx context.Context) (*Monitor, error) {
	dump, err := startPWDump(ctx)
	if err != nil {
		return nil, err
	}

	monitor := &Monitor{
		listeners: make([]*Listener, 0),
	}

	go monitor.run(ctx, dump)

	// Close all monitors when we the context is cancelled.
	go func() {
		<-ctx.Done()
		monitor.mu.RLock()
		defer monitor.mu.RUnlock()
		for listener := range slices.Values(monitor.listeners) {
			close(listener.eventCh)
		}
	}()

	slogctx.FromCtx(ctx).Debug("Monitoring pipewire for events.")

	return monitor, nil
}

// run dispatches events from pw-dump to the listeners, restarting pw-dump with
// an increasing delay whenever it exits.
func (m *Monitor) run(ctx context.Context, dump *pwDump) {
	delay := restartMinDelay

	for {
		started := time.Now()
		m.monitor(ctx, dump)

		if time.Since(started) > restartMaxDelay {
			delay = restartMinDelay
		}

		for {
			select {
			case <-ctx.Done():
				slogctx.FromCtx(ctx).Debug("Stopped monitoring pipewire for events.")
				return
			case <-time.After(delay):
			}

			delay = min(2*delay, restartMaxDelay)

			var err error
			if dump, err = startPWDump(ctx); err == nil {
				break
			}
			slogctx.FromCtx(ctx).Warn("Could not restart pw-dump.",
				slog.Any("error", err))
		}

		slogctx.FromCtx(ctx).Debug("Restarted pw-dump.")
	}
}

// monitor dispatches events from the given pw-dump until its output ends or
// can no longer be decoded, then stops it. As pw-dump will report all objects
// again when restarted, removal events for all objects it reported are
// dispatched, so that listeners do not track stale objects.
func (m *Monitor) monitor(ctx context.Context, dump *pwDump) {
	ids := m.decode(ctx, dump.stdout)

	// Stop pw-dump in case decoding failed while it is still running.
	dump.cancel()
	if err := dump.cmd.Wait(); err != nil && ctx.Err() == nil {
		slogctx.FromCtx(ctx).Warn("pw-dump exited with error.",
			slog.Any("err", err))
	} else {
		slogctx.FromCtx(ctx).Debug("pw-dump exited.")
	}

	// Listener channels are closed when the context is cancelled.
	if ctx.Err() != nil {
		return
	}
	for id := range ids {
		m.dispatch(&Event{ID: id})
	}
}

// decode reads the pw-dump output and dispatches its events to the listeners.
// It returns the ids of the objects that pw-dump reported and did not remove,
// once the output ends or cannot be decoded.
func (m *Monitor) decode(ctx context.Context, r io.Reader) map[int]struct{} {
	ids := make(map[int]struct{})
	dec := json.NewDecoder(r)

	for {
		// pw-dump outputs each batch of events as a JSON array. Errors from
		// the decoder are sticky, so any error ends decoding.
		if _, err := dec.Token(); err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
				slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "pw-dump: failed to read JSON token.",
					slog.Any("error", err))
			}

			return ids
		}

		for dec.More() {
			var event Event
			if err := dec.Decode(&event); err != nil {
				slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "Error decoding pw-dump output.",
					slog.Any("error", err))

				return ids
			}
			event.CapturedAt = time.Now().UTC()

			if event.IsRemovalEvent() {
				delete(ids, event.ID)
			} else if event.ID != 0 {
				ids[event.ID] = struct{}{}
			}

			m.dispatch(&event)
		}

		// Read the closing bracket of the array.
		if _, err := dec.Token(); err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
				slogctx.FromCtx(ctx).Log(ctx, logging.LevelTrace, "pw-dump: failed to read JSON token.",
					slog.Any("error", err))
			}

			return ids
		}
	}
}

// dispatch filters the event through all listeners and sends the event to
// whichever listeners want it.
func (m *Monitor) dispatch(event *Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for listener := range slices.Values(m.listeners) {
		if listener.filterFunc(event) {
			go func() {
				listener.eventCh <- *event
			}()
		}
	}
}

func (m *Monitor) AddListener(filterFunc func(*Event) bool) chan Event {
	m.mu.Lock()
	defer m.mu.Unlock()

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
