// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/event"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	loginBasePath        = "/org/freedesktop/login1"
	loginBaseInterface   = "org.freedesktop.login1"
	managerInterface     = loginBaseInterface + ".Manager"
	sessionAddedSignal   = "SessionNew"
	sessionRemovedSignal = "SessionRemoved"
	listSessionsMethod   = managerInterface + ".ListSessions"

	usersSensorUnits = "users"
	usersSensorIcon  = "mdi:account"

	userSessionsPreferencesID = sensorsPrefPrefix + "users"

	sessionStartedEventName = "session_started"
	sessionStoppedEventName = "session_stopped"
)

var (
	_ workers.EntityWorker = (*UserSessionSensorWorker)(nil)
	_ workers.EntityWorker = (*UserSessionEventsWorker)(nil)
)

type UserSessionSensorWorker struct {
	*models.WorkerMetadata

	bus   *dbusx.Bus
	prefs *UserSessionsPrefs
}

func NewUserSessionSensorWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &UserSessionSensorWorker{
		WorkerMetadata: models.SetWorkerMetadata("user_sessions", "User sessions"),
	}

	var ok bool

	worker.bus, ok = linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, fmt.Errorf("get system bus: %w", linux.ErrNoSystemBus)
	}

	defaultPrefs := &UserSessionsPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(userSessionsPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

func (w *UserSessionSensorWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sessionAddedSignal, sessionRemovedSignal),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("watch user sessions: %w", err)
	}
	sensorCh := make(chan models.Entity)

	sendUpdate := func() {
		users, err := w.getUsers()
		if err != nil {
			slogctx.FromCtx(ctx).Debug("Failed to get list of user sessions.", slog.Any("error", err))
		} else {
			sensorCh <- newUsersSensor(ctx, users)
		}
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case <-triggerCh:
				go sendUpdate()
			}
		}
	}()

	// Send an initial sensor update.
	go sendUpdate()

	return sensorCh, nil
}

func (w *UserSessionSensorWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *UserSessionSensorWorker) getUsers() ([]string, error) {
	userData, err := dbusx.GetData[[][]any](w.bus, loginBasePath, loginBaseInterface, listSessionsMethod)
	if err != nil {
		return nil, fmt.Errorf("get users from D-Bus: %w", err)
	}

	var users []string

	for _, u := range userData {
		if user, ok := u[2].(string); ok {
			users = append(users, user)
		}
	}

	return users, nil
}

func newUsersSensor(ctx context.Context, users []string) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName("Current Users"),
		sensor.WithID("current_users"),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits(usersSensorUnits),
		sensor.WithIcon(usersSensorIcon),
		sensor.WithState(len(users)),
		sensor.WithDataSourceAttribute(linux.DataSrcDBus),
		sensor.WithAttribute("usernames", users),
	)
}

type UserSessionEventsWorker struct {
	*sessionTracker
	*models.WorkerMetadata

	prefs *UserSessionsPrefs
}

func NewUserSessionEventsWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &UserSessionEventsWorker{
		WorkerMetadata: models.SetWorkerMetadata("user_session_events", "User session events"),
	}

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, fmt.Errorf("get system bus: %w", linux.ErrNoSystemBus)
	}

	worker.sessionTracker = &sessionTracker{
		bus:      bus,
		sessions: make(map[string]map[string]any),
	}

	defaultPrefs := &UserSessionsPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(userSessionsPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	currentSessions, err := dbusx.GetData[[][]any](bus, loginBasePath, loginBaseInterface, listSessionsMethod)
	if err != nil {
		return worker, fmt.Errorf("get user sessions from D-Bus: %w", err)
	}

	for session := range slices.Values(currentSessions) {
		s, ok := session[4].(string)
		if ok {
			worker.trackSession(s)
		}
	}

	return worker, nil
}

type sessionTracker struct {
	bus      *dbusx.Bus
	sessions map[string]map[string]any
	mu       sync.Mutex
}

func (t *sessionTracker) trackSession(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sessions[path] = t.getSessionDetails(path)
}

func (t *sessionTracker) unTrackSession(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, path)
}

func (t *sessionTracker) getSessionProp(path, prop string) (dbus.Variant, error) {
	var value dbus.Variant
	value, err := dbusx.NewProperty[dbus.Variant](t.bus,
		path,
		loginBaseInterface,
		loginBaseInterface+".Session."+prop).Get()
	if err != nil {
		return dbus.MakeVariant("Unknown"),
			fmt.Errorf("could not retrieve session property %s (session %s): %w", prop, path, err)
	}

	return value, nil
}

func (t *sessionTracker) getSessionDetails(path string) map[string]any {
	sessionDetails := make(map[string]any)

	sessionDetails["user"] = sessionProp[string](t.getSessionProp, path, "Name")
	sessionDetails["remote"] = sessionProp[bool](t.getSessionProp, path, "Remote")

	if _, ok := sessionDetails["remote"].(bool); ok {
		sessionDetails["remote_host"] = sessionProp[string](t.getSessionProp, path, "RemoteHost")
		sessionDetails["remote_user"] = sessionProp[string](t.getSessionProp, path, "RemoteUser")
	}

	sessionDetails["desktop"] = sessionProp[string](t.getSessionProp, path, "Desktop")
	sessionDetails["service"] = sessionProp[string](t.getSessionProp, path, "Service")
	sessionDetails["type"] = sessionProp[string](t.getSessionProp, path, "Type")

	return sessionDetails
}

//nolint:gocognit
func (w *UserSessionEventsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sessionAddedSignal, sessionRemovedSignal),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("watch user sessions: %w", err)
	}

	eventCh := make(chan models.Entity)

	go func() {
		defer close(eventCh)

		for {
			select {
			case <-ctx.Done():
				return
			case trigger := <-triggerCh:
				if len(trigger.Content) != 2 {
					continue
				}
				// If the trigger does not contain a session path, ignore.
				path, ok := trigger.Content[1].(dbus.ObjectPath)
				if !ok {
					continue
				}
				// Send the appropriate event type.
				switch {
				case strings.Contains(trigger.Signal, sessionAddedSignal):
					// Add the session to the tracker.
					w.trackSession(string(path))
					// Send the session added event.
					entity, err := event.NewEvent(sessionStartedEventName, w.sessions[string(path)])
					if err != nil {
						slogctx.FromCtx(ctx).Warn("Could not generate users event.", slog.Any("error", err))
					} else {
						eventCh <- entity
					}
				case strings.Contains(trigger.Signal, sessionRemovedSignal):
					// Send the session removed event.
					entity, err := event.NewEvent(sessionStoppedEventName, w.sessions[string(path)])
					if err != nil {
						slogctx.FromCtx(ctx).Warn("Could not generate users event.", slog.Any("error", err))
					} else {
						eventCh <- entity
					}
					// Remove the session from the tracker.
					w.unTrackSession(string(path))
				}
			}
		}
	}()

	return eventCh, nil
}

func (w *UserSessionEventsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func sessionProp[T any](getFunc func(string, string) (dbus.Variant, error), path, prop string) T {
	var (
		err     error
		value   T
		variant dbus.Variant
	)

	if variant, err = getFunc(path, prop); err != nil {
		return value
	}

	value, _ = dbusx.VariantToValue[T](variant)

	return value
}
