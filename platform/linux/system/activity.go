// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"slices"
	"sync/atomic"
	"time"

	"github.com/holoplot/go-evdev"
	slogctx "github.com/veqryn/slog-context"

	"kernel.org/pub/linux/libs/security/libcap/cap"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	activityWorkerID                 = "activity"
	activityWorkerDefaultIdleTimeout = 5 * time.Second

	sleepSignal    = "PrepareForSleep"
	shutdownSignal = "PrepareForShutdown"
)

// newActivityState generates the User Activity sensor with the given state.
func newActivityState(ctx context.Context, state bool) models.Entity {
	icon := "mdi:bell-off"
	if state {
		icon = "mdi:bell-ring"
	}

	return models.NewSensor(ctx,
		models.WithName("User Activity"),
		models.WithID("user_activity"),
		models.AsTypeBinarySensor(),
		models.WithIcon(icon),
		models.WithState(state),
	)
}

// resetOnPowerSignal switches the sensor off when the system announces it is
// going to sleep or shutting down, and records that input events are to be
// ignored until it resumes.
func resetOnPowerSignal(
	ctx context.Context,
	event dbusx.Trigger,
	sensorCh chan<- models.Entity,
	activityDetected, goingDown *atomic.Bool,
) {
	if len(event.Content) == 0 {
		return
	}

	going, ok := event.Content[0].(bool)
	if !ok {
		return
	}

	// Both signals are sent again with false on resume, which lifts the
	// suppression rather than switching the sensor off.
	goingDown.Store(going)

	if !going {
		return
	}

	// Switch the sensor off while the agent can still send it.
	if activityDetected.Load() {
		activityDetected.Store(false)
		sensorCh <- newActivityState(ctx, false)
	}
}

type activityWorkerPrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	IdleTimeout string `toml:"idle_timeout"`
}

type activityWorker struct {
	*models.WorkerMetadata

	prefs        *activityWorkerPrefs `toml:",squash"`
	bus          *dbusx.Bus
	inputDevices []*evdev.InputDevice
	activity     chan bool
}

// NewUserActivitySensor creates a worker that detects when the user is using the device, through input events.
func NewUserActivitySensor(ctx context.Context) (workers.EntityWorker, error) {
	worker := &activityWorker{
		WorkerMetadata: models.SetWorkerMetadata(activityWorkerID, "User Activity"),
		activity:       make(chan bool, 1),
	}

	// Load worker preferences.
	defaultPrefs := &activityWorkerPrefs{
		IdleTimeout: activityWorkerDefaultIdleTimeout.String(),
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(sensorsPrefPrefix+"app_sensors", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	if worker.prefs.IsDisabled() {
		return worker, nil
	}

	// Get the system bus, used to detect the system going to sleep or shutting
	// down. This is not fatal: without it, the sensor just cannot switch itself
	// off before the system goes down.
	var ok bool
	if worker.bus, ok = linux.CtxGetSystemBus(ctx); !ok {
		slogctx.FromCtx(ctx).Debug("No system bus, user activity will not be reset on sleep/shutdown.")
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
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
	}()

	// Start monitoring input devices.
	w.monitorInputDevices(ctx)

	// Handle user activity events.
	go w.monitorActivity(ctx, sensorCh, w.watchPowerSignals(ctx), idleTimeout)

	return sensorCh, nil
}

func (w *activityWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

// watchPowerSignals watches for the system going to sleep or shutting down. A
// nil channel is returned when the signals cannot be watched, in which case
// activity cannot be switched off before the system goes down.
func (w *activityWorker) watchPowerSignals(ctx context.Context) <-chan dbusx.Trigger {
	if w.bus == nil {
		return nil
	}

	powerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sleepSignal, shutdownSignal),
	).Start(ctx, w.bus)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not watch for sleep/shutdown, user activity will not be reset.",
			slog.Any("error", err))

		return nil
	}

	return powerCh
}

// monitorActivity reports user activity from input events, and switches it off
// when the system announces it is going to sleep or shutting down. Without
// that, activity detected as the system goes down is left switched on in Home
// Assistant until the user is active again.
func (w *activityWorker) monitorActivity(
	ctx context.Context,
	sensorCh chan<- models.Entity,
	powerCh <-chan dbusx.Trigger,
	idleTimeout time.Duration,
) {
	var activityDetected atomic.Bool
	// goingDown is set while the system is on its way to sleep or shutdown.
	// Input events are ignored until it resumes, so an event already queued
	// when the system went down cannot switch the sensor back on.
	var goingDown atomic.Bool

	slogctx.FromCtx(ctx).Debug("Started monitoring user activity.")

	for {
		select {
		case <-ctx.Done():
			slogctx.FromCtx(ctx).Debug("Stopped monitoring user activity.")

			return
		case <-w.activity:
			if goingDown.Load() {
				continue
			}

			if !activityDetected.Load() {
				activityDetected.Store(true)
				sensorCh <- newActivityState(ctx, true)
			}
		case <-time.After(idleTimeout):
			if activityDetected.Load() {
				activityDetected.Store(false)
				sensorCh <- newActivityState(ctx, false)
			}
		case event := <-powerCh:
			resetOnPowerSignal(ctx, event, sensorCh, &activityDetected, &goingDown)
		}
	}
}

func (w *activityWorker) monitorInputDevices(ctx context.Context) {
	for device := range slices.Values(w.inputDevices) {
		go func() {
			name, _ := device.Name()
			slogctx.FromCtx(ctx).Debug("Monitoring input device.",
				slog.String("device", device.Path()),
				slog.String("name", name),
			)
			for {
				select {
				case <-ctx.Done():
					defer device.Close()
					slogctx.FromCtx(ctx).Debug("Stopped monitoring input device.",
						slog.String("device", device.Path()),
						slog.String("name", name),
					)
					return
				default:
					// Read input event
					event, err := device.ReadOne()
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
