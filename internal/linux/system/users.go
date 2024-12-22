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
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/event"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
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

	userSessionSensorWorkerID = "user_session_sensor_worker"
	userSessionEventWorkerID  = "user_session_event_worker"

	sessionStartedEventName = "session_started"
	sessionStoppedEventName = "session_stopped"
)

func newUsersSensor(users []string) sensor.Entity {
	return sensor.NewSensor(
		sensor.WithName("Current Users"),
		sensor.WithID("current_users"),
		sensor.WithStateClass(types.StateClassMeasurement),
		sensor.WithUnits(usersSensorUnits),
		sensor.WithState(
			sensor.WithIcon(usersSensorIcon),
			sensor.WithValue(len(users)),
			sensor.WithDataSourceAttribute(linux.DataSrcDbus),
			sensor.WithAttribute("usernames", users),
		),
	)
}

type UserSessionSensorWorker struct {
	getUsers  func() ([]string, error)
	triggerCh chan dbusx.Trigger
	linux.EventSensorWorker
}

func (w *UserSessionSensorWorker) Events(ctx context.Context) (chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)

	sendUpdate := func() {
		users, err := w.getUsers()
		if err != nil {
			slog.With(slog.String("worker", userSessionSensorWorkerID)).Debug("Failed to get list of user sessions.", slog.Any("error", err))
		} else {
			sensorCh <- newUsersSensor(users)
		}
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.triggerCh:
				go sendUpdate()
			}
		}
	}()

	// Send an initial sensor update.
	go sendUpdate()

	return sensorCh, nil
}

func (w *UserSessionSensorWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	users, err := w.getUsers()

	return []sensor.Entity{newUsersSensor(users)}, err
}

func NewUserSessionSensorWorker(ctx context.Context) (*UserSessionSensorWorker, error) {
	worker := &UserSessionSensorWorker{}
	worker.WorkerID = userSessionSensorWorkerID

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sessionAddedSignal, sessionRemovedSignal),
	).Start(ctx, bus)
	if err != nil {
		return nil, fmt.Errorf("unable to set-up D-Bus watch for user sessions: %w", err)
	}

	worker.triggerCh = triggerCh

	worker.getUsers = func() ([]string, error) {
		userData, err := dbusx.GetData[[][]any](bus, loginBasePath, loginBaseInterface, listSessionsMethod)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve users from D-Bus: %w", err)
		}

		var users []string

		for _, u := range userData {
			if user, ok := u[2].(string); ok {
				users = append(users, user)
			}
		}

		return users, nil
	}

	return worker, nil
}

type UserSessionEventsWorker struct {
	triggerCh chan dbusx.Trigger
	tracker   sessionTracker
	linux.EventWorker
}

type sessionTracker struct {
	getSessionProp func(path, prop string) (dbus.Variant, error)
	sessions       map[string]map[string]any
	mu             sync.Mutex
}

func (t *sessionTracker) addSession(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sessions[path] = t.getSessionDetails(path)
}

func (t *sessionTracker) removeSession(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, path)
}

func (t *sessionTracker) getSessionDetails(path string) map[string]any {
	sessionDetails := make(map[string]any)

	sessionDetails["user"] = sessionProp[string](t.getSessionProp, path, "Name")
	sessionDetails["remote"] = sessionProp[bool](t.getSessionProp, path, "Remote")

	if sessionDetails["remote"].(bool) {
		sessionDetails["remote_host"] = sessionProp[string](t.getSessionProp, path, "RemoteHost")
		sessionDetails["remote_user"] = sessionProp[string](t.getSessionProp, path, "RemoteUser")
	}

	sessionDetails["desktop"] = sessionProp[string](t.getSessionProp, path, "Desktop")
	sessionDetails["service"] = sessionProp[string](t.getSessionProp, path, "Service")
	sessionDetails["type"] = sessionProp[string](t.getSessionProp, path, "Type")

	return sessionDetails
}

func (w *UserSessionEventsWorker) Events(ctx context.Context) (<-chan event.Event, error) {
	eventCh := make(chan event.Event)

	go func() {
		defer close(eventCh)

		for {
			select {
			case <-ctx.Done():
				return
			case trigger := <-w.triggerCh:
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
					w.tracker.addSession(string(path))
					eventCh <- event.Event{
						EventType: sessionStartedEventName,
						EventData: w.tracker.sessions[string(path)],
					}
				case strings.Contains(trigger.Signal, sessionRemovedSignal):
					eventCh <- event.Event{
						EventType: sessionStoppedEventName,
						EventData: w.tracker.sessions[string(path)],
					}
					w.tracker.removeSession(string(path))
				}
			}
		}
	}()

	return eventCh, nil
}

func NewUserSessionEventsWorker(ctx context.Context) (*linux.EventWorker, error) {
	worker := linux.NewEventWorker(userSessionEventWorkerID)

	bus, ok := linux.CtxGetSystemBus(ctx)
	if !ok {
		return worker, linux.ErrNoSystemBus
	}

	eventWorker := &UserSessionEventsWorker{
		tracker: sessionTracker{
			sessions: make(map[string]map[string]any),
			getSessionProp: func(path, prop string) (dbus.Variant, error) {
				value, err := dbusx.NewProperty[dbus.Variant](bus,
					path,
					loginBaseInterface,
					loginBaseInterface+".Session."+prop).Get()
				if err != nil {
					return dbus.MakeVariant(sensor.StateUnknown),
						fmt.Errorf("could not retrieve session property %s (session %s): %w", prop, path, err)
				}

				return value, nil
			},
		},
	}

	currentSessions, err := dbusx.GetData[[][]any](bus, loginBasePath, loginBaseInterface, listSessionsMethod)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve sessions from D-Bus: %w", err)
	}

	for _, session := range currentSessions {
		eventWorker.tracker.addSession(string(session[4].(dbus.ObjectPath)))
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(loginBasePath),
		dbusx.MatchInterface(managerInterface),
		dbusx.MatchMembers(sessionAddedSignal, sessionRemovedSignal),
	).Start(ctx, bus)
	if err != nil {
		return nil, fmt.Errorf("unable to set-up D-Bus watch for user sessions: %w", err)
	}

	eventWorker.triggerCh = triggerCh

	worker.EventType = eventWorker

	return worker, nil
}

//nolint:errcheck
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
