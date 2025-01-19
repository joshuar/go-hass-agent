// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
//go:generate go run golang.org/x/tools/cmd/stringer -type=dbusType -output busType_strings.go
package dbusx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/user"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
)

const (
	SessionBus dbusType = iota // session
	SystemBus                  // system

	PropInterface         = "org.freedesktop.DBus.Properties"
	PropChangedSignal     = PropInterface + ".PropertiesChanged"
	loginBasePath         = "/org/freedesktop/login1"
	loginBaseInterface    = "org.freedesktop.login1"
	loginManagerInterface = loginBaseInterface + ".Manager"
	listSessionsMethod    = loginManagerInterface + ".ListSessions"
	listUsersMethod       = loginManagerInterface + ".ListUsers"
)

var (
	ErrNoBus          = errors.New("no D-Bus connection")
	ErrNoBusCtx       = errors.New("no D-Bus connection in context")
	ErrNotPropChanged = errors.New("signal contents do not appear to represent changed properties")
	ErrParseInterface = errors.New("could not parse interface name")
	ErrParseNewProps  = errors.New("could not parse changed properties")
	ErrParseOldProps  = errors.New("could not parse invalidated properties")
	ErrNotValChanged  = errors.New("signal contents do not appear to represent a value change")
	ErrParseNewVal    = errors.New("could not parse new value")
	ErrParseOldVal    = errors.New("could not parse old value")
	ErrNoSessionPath  = errors.New("could not determine session path")
	ErrInvalidPath    = errors.New("invalid D-Bus path")
	ErrUnknownBus     = errors.New("unknown bus")
)

var DbusTypeMap = map[string]dbusType{
	"session": 0,
	"system":  1,
}

type dbusType int

// Values represents a property value that changed. It contains the new and old
// values.
type Values[T any] struct {
	New T
	Old T
}

// Bus represents a particular D-Bus connection to either the system or session
// bus.
type Bus struct {
	conn     *dbus.Conn
	traceLog func(msg string, args ...any)
	busType  dbusType
	logger   *slog.Logger
}

func (b *Bus) getObject(intr, path string) dbus.BusObject {
	return b.conn.Object(intr, dbus.ObjectPath(path))
}

// NewBus creates a D-Bus connection to the requested bus. If a connection
// cannot be established, an error is returned.
//
//nolint:sloglint
func NewBus(ctx context.Context, busType dbusType) (*Bus, error) {
	var (
		conn *dbus.Conn
		err  error
	)

	// Connect to the requested bus.
	switch busType {
	case SessionBus:
		conn, err = dbus.ConnectSessionBus(dbus.WithContext(ctx))
	case SystemBus:
		conn, err = dbus.ConnectSystemBus(dbus.WithContext(ctx))
	default:
		return nil, ErrUnknownBus
	}
	// If the connection fails, we bails.
	if err != nil {
		return nil, fmt.Errorf("could not connect to bus: %w", err)
	}

	logger := logging.FromContext(ctx).
		WithGroup("dbus").
		With(
			slog.String("bus", busType.String()),
		)

	// Set up our bus object.
	bus := &Bus{
		conn:    conn,
		busType: busType,
		logger:  logger,
		traceLog: func(msg string, args ...any) {
			logger.Log(ctx, logging.LevelTrace, msg, args...)
		},
	}

	// Start a goroutine to close the connection when the context is canceled
	// (i.e. agent shutdown).
	go func() {
		defer conn.Close()
		<-ctx.Done()
	}()

	return bus, nil
}

// GetData fetches data using the given method from D-Bus, as the provided type.
// If there is an error or the result cannot be stored in the given type, it
// will return an non-nil error. To execute a method, see Call. To get the value
// of a property, see GetProp.
func GetData[D any](bus *Bus, path, dest, method string, args ...any) (D, error) {
	var (
		data D
		err  error
	)

	bus.traceLog("Getting data.", slog.String("path", path), slog.String("dest", dest), slog.String("method", method))

	obj := bus.getObject(dest, path)

	if args != nil {
		err = obj.Call(method, 0, args...).Store(&data)
	} else {
		err = obj.Call(method, 0).Store(&data)
	}

	if err != nil {
		return data, fmt.Errorf("%s: unable to get data %s from %s: %w", bus.busType.String(), method, dest, err)
	}

	return data, nil
}

// GetSessionPath retrieves the user's session path from D-Bus. This is used by
// various sensors to retrieve other data.
func (b *Bus) GetSessionPath() (string, error) {
	var userPath dbus.ObjectPath
	// Get the user running the agent.
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("unable to determine user: %w", err)
	}
	// Fetch all logged-in users from D-Bus.
	users, err := GetData[[][]any](b, loginBasePath, loginBaseInterface, listUsersMethod)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve session path: %w", err)
	}
	// Find the systemd-logind D-Bus user path for the user.
	for _, s := range users {
		if thisUser, ok := s[1].(string); ok && thisUser == usr.Username {
			if p, ok := s[2].(dbus.ObjectPath); ok {
				userPath = p
				b.logger.Debug("Retrieved user path.", slog.String("path", string(userPath)))
			}
		}
	}
	// If we didn't find a path, bail.
	if string(userPath) == "" {
		return "", ErrNoSessionPath
	}
	// Fetch the desktop session for the user, which will contain the correct
	// systemd-logind D-Bus session path for the user.
	displayDetails, err := NewProperty[[]any](b, string(userPath), loginBaseInterface, loginBaseInterface+".User.Display").Get()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve session path: %w", err)
	}
	// Extract/validate the session path.
	sessionPath, ok := displayDetails[1].(dbus.ObjectPath)
	if !ok || string(sessionPath) == "" {
		return "", ErrNoSessionPath
	}
	// We have a valid session path.
	b.logger.Debug("Retrieved session path.", slog.String("path", string(sessionPath)))
	// Return the session path as a string.
	return string(sessionPath), nil
}

// ParseValueChange treats the given signal body as matching a value change of a
// property from an old value to a new value. It will parse the signal body into
// a Value object with old/new values of the given type. If there was a problem
// parsing the signal body, an error will be returned with details of the
// problem.
//
//nolint:mnd
func ParseValueChange[T any](valueChanged []any) (*Values[T], error) {
	values := &Values[T]{}

	var ok bool

	if len(valueChanged) != 2 {
		return nil, ErrNotValChanged
	}

	values.New, ok = valueChanged[0].(T)
	if !ok {
		return nil, ErrParseNewVal
	}

	values.Old, ok = valueChanged[1].(T)
	if !ok {
		return nil, ErrParseOldVal
	}

	return values, nil
}

// VariantToValue converts a dbus.Variant value into the specified Go type. If
// the value is nil or it cannot be converted, then the return value will be the
// default value of the specified type.
func VariantToValue[S any](variant dbus.Variant) (S, error) {
	var value S

	err := variant.Store(&value)
	if err != nil {
		return value, fmt.Errorf("unable to convert D-Bus variant %v to type %T: %w", variant, value, err)
	}

	return value, nil
}
