// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"sync/atomic"
	"time"

	"github.com/holoplot/go-evdev"
	slogctx "github.com/veqryn/slog-context"

	"kernel.org/pub/linux/libs/security/libcap/cap"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	activityWorkerID                 = "activity"
	activityWorkerDefaultIdleTimeout = 5 * time.Second
)

type activityWorkerPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	IdleTimeout string `toml:"idle_timeout"`
}

type activityWorker struct {
	*models.WorkerMetadata

	prefs        *activityWorkerPrefs `toml:",squash"`
	inputDevices []*evdev.InputDevice
	activity     chan bool
}

// NewUserActivitySensor creates a worker that detects when the user is using the device, through input events.
func NewUserActivitySensor(ctx context.Context) (workers.EntityWorker, error) {
	worker := &activityWorker{
		WorkerMetadata: models.SetWorkerMetadata(activityWorkerID, "User Activity"),
		activity:       make(chan bool),
	}

	// Check for required capabilities.
	group, err := user.LookupGroup("input")
	if err != nil {
		return worker, fmt.Errorf("lookup group: %w", err)
	}
	capabilities := &linux.Checks{
		Groups:       []user.Group{*group},
		Capabilities: []cap.Value{cap.SETUID, cap.SETGID},
	}
	passed, err := capabilities.Passed()
	if err != nil || !passed {
		return worker, fmt.Errorf("check capabilities: %w", err)
	}

	// Load worker preferences.
	defaultPrefs := &activityWorkerPrefs{
		IdleTimeout: activityWorkerDefaultIdleTimeout.String(),
	}
	worker.prefs, err = workers.LoadWorkerPreferences(sensorsPrefPrefix+"app_sensors", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	// Get input devices.
	// TODO: need to detect additions/removals.
	worker.inputDevices, err = initInputDevices(ctx)
	if err != nil {
		return worker, fmt.Errorf("init input devices: %w", err)
	}

	return worker, nil
}

func (w *activityWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	idleTimeout, err := time.ParseDuration(w.prefs.IdleTimeout)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Unable to parse idle timeout in preferences, using default of 5s.")
		idleTimeout = activityWorkerDefaultIdleTimeout
	}

	sensorCh := make(chan models.Entity)
	var activityDetected atomic.Bool

	// Start monitoring input devices.
	w.monitorInputDevices(ctx)

	// Handle user activity events.
	go func() {
		slogctx.FromCtx(ctx).Debug("Started monitoring user activity.")
		for {
			select {
			case <-ctx.Done():
				slogctx.FromCtx(ctx).Debug("Stopped monitoring user activity.")
				return
			case <-w.activity:
				if !activityDetected.Load() {
					activityDetected.Store(true)
					sensorCh <- sensor.NewSensor(ctx,
						sensor.WithName("User Activity"),
						sensor.WithID("user_activity"),
						sensor.AsTypeBinarySensor(),
						sensor.WithIcon("mdi:bell-ring"),
						sensor.WithState(true),
					)
				}
			case <-time.After(idleTimeout):
				if activityDetected.Load() {
					activityDetected.Store(false)
					sensorCh <- sensor.NewSensor(ctx,
						sensor.WithName("User Activity"),
						sensor.WithID("user_activity"),
						sensor.AsTypeBinarySensor(),
						sensor.WithIcon("mdi:bell-off"),
						sensor.WithState(false),
					)
				}
			}
		}
	}()

	return sensorCh, nil
}

func (w *activityWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *activityWorker) monitorInputDevices(ctx context.Context) {
	for _, device := range w.inputDevices {
		dev := device // Capture for goroutine
		go func() {
			defer dev.Close()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Read input event
					event, err := dev.ReadOne()
					if err != nil {
						// Check if it's just an error from closed device
						if !errors.Is(err, os.ErrClosed) {
							time.Sleep(inputReadRetryDelay)
						}
						continue
					}

					// Ignore sync events (they're just markers, not actual input)
					if event.Type == evdev.EV_SYN {
						continue
					}

					w.activity <- true
				}
			}
		}()
	}
}
