// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/holoplot/go-evdev"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	lastActivePollInterval  = 10 * time.Second
	lastActivePollJitter    = time.Second
	lastActivePreferencesID = sensorsPrefPrefix + "last_active"

	mprisDBusPath        = "/org/mpris/MediaPlayer2"
	mprisDBusNamespace   = "org.mpris.MediaPlayer2"
	mprisPlayerInterface = "org.mpris.MediaPlayer2.Player"
)

var (
	ErrInitLastActiveWorker = errors.New("could not init last active worker")
	ErrNoInputDevices       = errors.New("no input devices found")
)

var (
	_ quartz.Job                  = (*lastActiveWorker)(nil)
	_ workers.PollingEntityWorker = (*lastActiveWorker)(nil)
)

// LastActivePrefs are the preferences for the last active sensor worker.
type LastActivePrefs struct {
	workers.CommonWorkerPrefs `toml:",squash"`

	UpdateInterval string `toml:"update_interval"`
}

// lastActiveWorker tracks the last time the system was actively used based on
// input device activity (keyboard/mouse) and media playback.
//
// The worker monitors:
// - Input events from /dev/input/event* devices using evdev (requires read permissions)
// - MPRIS-compatible media players via D-Bus for active playback
//
// System Requirements:
// - Read access to /dev/input/event* devices (typically via 'input' group membership)
// - D-Bus session bus access for MPRIS media player detection
type lastActiveWorker struct {
	*workers.PollingEntityWorkerData
	*models.WorkerMetadata

	prefs            *LastActivePrefs
	lastActivityTime time.Time
	mu               sync.RWMutex
	inputDevices     []*evdev.InputDevice
	bus              *dbusx.Bus
}

// NewLastActiveWorker creates a new worker to track the last active time of the system.
func NewLastActiveWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &lastActiveWorker{
		WorkerMetadata:          models.SetWorkerMetadata("last_active", "Last active time tracking"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		lastActivityTime:        time.Now(),
		inputDevices:            make([]*evdev.InputDevice, 0),
	}

	// Load preferences
	defaultPrefs := &LastActivePrefs{
		UpdateInterval: lastActivePollInterval.String(),
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(lastActivePreferencesID, defaultPrefs)
	if err != nil {
		return worker, errors.Join(ErrInitLastActiveWorker, err)
	}

	// Set up polling trigger
	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = lastActivePollInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, lastActivePollJitter)

	// Get session bus for MPRIS
	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		slogctx.FromCtx(ctx).Warn("Could not get session bus for MPRIS monitoring, media playback detection will be unavailable.")
	} else {
		worker.bus = bus
	}

	// Initialize input device monitoring
	if err := worker.initInputDevices(ctx); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not initialize input device monitoring.",
			slog.Any("error", err))
	}

	return worker, nil
}

// initInputDevices discovers and opens input devices for monitoring.
// It looks for keyboard and mouse devices in /dev/input/event*.
func (w *lastActiveWorker) initInputDevices(ctx context.Context) error {
	inputDir := filepath.Join(linux.DevFSRoot, "input")
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("could not read input directory: %w", err)
	}

	deviceCount := 0
	for _, entry := range entries {
		// Only process event devices
		if !strings.HasPrefix(entry.Name(), "event") {
			continue
		}

		devicePath := filepath.Join(inputDir, entry.Name())
		device, err := evdev.Open(devicePath)
		if err != nil {
			// Permission errors are expected for devices we can't access
			slogctx.FromCtx(ctx).Debug("Could not open input device.",
				slog.String("device", devicePath),
				slog.Any("error", err))
			continue
		}

		// Check if this is a keyboard or mouse device
		// Keyboards have key capabilities, mice have relative or absolute positioning
		capableTypes := device.CapableTypes()

		hasKeys := false
		hasPointer := false

		// Check for keyboard capabilities (EV_KEY events)
		for _, evType := range capableTypes {
			if evType == evdev.EV_KEY {
				// Check if there are actual key events (keyboards have many, mice have few)
				keyCaps := device.CapableEvents(evdev.EV_KEY)
				if len(keyCaps) > 10 {
					hasKeys = true
				}
			}
			// Check for pointer capabilities (relative or absolute motion)
			if evType == evdev.EV_REL || evType == evdev.EV_ABS {
				hasPointer = true
			}
		}

		if hasKeys || hasPointer {
			w.inputDevices = append(w.inputDevices, device)
			deviceCount++
			name, _ := device.Name()
			slogctx.FromCtx(ctx).Debug("Monitoring input device.",
				slog.String("device", devicePath),
				slog.String("name", name))
		} else {
			device.Close()
		}
	}

	if deviceCount == 0 {
		return ErrNoInputDevices
	}

	slogctx.FromCtx(ctx).Info("Initialized input device monitoring.",
		slog.Int("device_count", deviceCount))

	return nil
}

// monitorInputDevices watches for activity on all configured input devices.
// It runs in a separate goroutine and updates lastActivityTime when events are detected.
func (w *lastActiveWorker) monitorInputDevices(ctx context.Context) {
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
					ev, err := dev.ReadOne()
					if err != nil {
						// Check if it's just an error from closed device
						if !errors.Is(err, os.ErrClosed) {
							time.Sleep(100 * time.Millisecond)
						}
						continue
					}

					// Ignore sync events (they're just markers, not actual input)
					if ev.Type == evdev.EV_SYN {
						continue
					}

					// Any non-sync event indicates activity
					w.mu.Lock()
					w.lastActivityTime = time.Now()
					w.mu.Unlock()
				}
			}
		}()
	}
}

// isMediaPlaying checks if any MPRIS-compatible media player is currently playing.
func (w *lastActiveWorker) isMediaPlaying(ctx context.Context) bool {
	if w.bus == nil {
		return false
	}

	// Get list of all MPRIS player names on the session bus
	names, err := w.bus.ListNames()
	if err != nil {
		slogctx.FromCtx(ctx).Debug("Could not list D-Bus names.", slog.Any("error", err))
		return false
	}

	// Check each MPRIS player
	for _, name := range names {
		if !strings.HasPrefix(name, mprisDBusNamespace) {
			continue
		}

		// Get PlaybackStatus property for this specific player
		// The service name (destination) should be the player's unique bus name
		prop := dbusx.NewProperty[string](w.bus, mprisDBusPath, name, mprisPlayerInterface+".PlaybackStatus")
		status, err := prop.Get()
		if err != nil {
			continue
		}

		// If any player is playing, consider system active
		if status == "Playing" {
			return true
		}
	}

	return false
}

// getLastActiveTime returns the time of the last detected activity.
func (w *lastActiveWorker) getLastActiveTime(ctx context.Context) time.Time {
	w.mu.RLock()
	lastActive := w.lastActivityTime
	w.mu.RUnlock()

	// Check if media is playing - if so, update last active to now
	if w.isMediaPlaying(ctx) {
		w.mu.Lock()
		w.lastActivityTime = time.Now()
		lastActive = w.lastActivityTime
		w.mu.Unlock()
	}

	return lastActive
}

func (w *lastActiveWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)

	// Start monitoring input devices
	w.monitorInputDevices(ctx)

	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start last active worker: %w", err)
	}
	return w.OutCh, nil
}

func (w *lastActiveWorker) Execute(ctx context.Context) error {
	lastActive := w.getLastActiveTime(ctx)
	timeSinceActive := time.Since(lastActive)

	w.OutCh <- sensor.NewSensor(ctx,
		sensor.WithName("Last Active"),
		sensor.WithID("last_active"),
		sensor.AsDiagnostic(),
		sensor.WithDeviceClass(class.SensorClassTimestamp),
		sensor.WithIcon("mdi:account-clock"),
		sensor.WithState(lastActive.Format(time.RFC3339)),
		sensor.WithAttribute("seconds_since_active", int(timeSinceActive.Seconds())),
		sensor.WithAttribute("minutes_since_active", int(timeSinceActive.Minutes())),
		sensor.WithDataSourceAttribute("evdev+mpris"),
	)
	return nil
}

func (w *lastActiveWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
