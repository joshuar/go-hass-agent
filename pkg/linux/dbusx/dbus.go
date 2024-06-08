// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"context"
	"errors"
	"fmt"
	"os/user"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=dbusType -output busType_strings.go
const (
	SessionBus dbusType = iota // session
	SystemBus                  // system

	PropInterface         = "org.freedesktop.DBus.Properties"
	PropChangedSignal     = PropInterface + ".PropertiesChanged"
	loginBasePath         = "/org/freedesktop/login1"
	loginBaseInterface    = "org.freedesktop.login1"
	loginManagerInterface = loginBaseInterface + ".Manager"
	listSessionsMethod    = loginManagerInterface + ".ListSessions"
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
)

var DbusTypeMap = map[string]dbusType{
	"session": 0,
	"system":  1,
}

type dbusType int

type Trigger struct {
	Signal  string
	Path    string
	Content []any
}

type Watch struct {
	Args          map[int]string
	Path          string
	PathNamespace string
	Interface     string
	Names         []string
	Bus           dbusType
}

func (w *Watch) Parse() []dbus.MatchOption {
	var matchers []dbus.MatchOption

	switch {
	case w.Path != "":
		matchers = append(matchers, dbus.WithMatchObjectPath(dbus.ObjectPath(w.Path)))
	case w.PathNamespace != "":
		matchers = append(matchers, dbus.WithMatchPathNamespace(dbus.ObjectPath(w.PathNamespace)))
	case w.Args != nil:
		for arg, value := range w.Args {
			matchers = append(matchers, dbus.WithMatchArg(arg, value))
		}
	case w.Interface != "":
		matchers = append(matchers, dbus.WithMatchInterface(w.Interface))
	case len(w.Names) != 0:
		for _, name := range w.Names {
			matchers = append(matchers, dbus.WithMatchMember(name))
		}
	}

	return matchers
}

type Properties struct {
	Interface   string
	Changed     map[string]dbus.Variant
	Invalidated []string
}

type Values[T any] struct {
	New T
	Old T
}

type bus struct {
	conn    *dbus.Conn
	busType dbusType
	wg      sync.WaitGroup
}

// newBus sets up D-Bus connections and channels for receiving signals. It
// creates both a system and session bus connection.
//
//nolint:contextcheck
func newBus(ctx context.Context, busType dbusType) (*bus, error) {
	var conn *dbus.Conn

	var err error

	dbusCtx, cancelFunc := context.WithCancel(context.Background())

	switch busType {
	case SessionBus:
		conn, err = dbus.ConnectSessionBus(dbus.WithContext(dbusCtx))
	case SystemBus:
		conn, err = dbus.ConnectSystemBus(dbus.WithContext(dbusCtx))
	}

	if err != nil {
		cancelFunc()

		return nil, fmt.Errorf("could not connect to bus: %w", err)
	}

	bus := &bus{
		conn:    conn,
		busType: busType,
		wg:      sync.WaitGroup{},
	}

	go func() {
		defer conn.Close()
		defer cancelFunc()
		<-ctx.Done()
		bus.wg.Wait()
	}()

	return bus, nil
}

// Call will call the given method, at the given path and interface, with the
// given args on the given bus. If the call fails or cannot be executed, a
// non-nil error will be returned. Call does not return any data. For fetching
// data from the bus, see GetData. For retrieving the value of a property, see
// GetProp.
func Call(ctx context.Context, bus dbusType, path, dest, method string, args ...any) error {
	busObj, ok := getBus(ctx, bus)
	if !ok {
		return ErrNoBusCtx
	}

	obj := busObj.conn.Object(dest, dbus.ObjectPath(path))
	if args != nil {
		return fmt.Errorf("%s: call could not retrieve object (%w)", bus.String(), obj.Call(method, 0, args...).Err)
	}

	err := obj.Call(method, 0).Err
	if err != nil {
		return fmt.Errorf("%s: unable to call method %s (args: %v): %w", bus.String(), method, args, err)
	}

	return obj.Call(method, 0).Err
}

// GetProp retrieves the value of the specified property from D-Bus as the given
// type. If the property cannot be retrieved, a non-nil error is returned.
func GetProp[P any](ctx context.Context, bus dbusType, path, dest, prop string) (P, error) {
	var value P

	busObj, ok := getBus(ctx, bus)
	if !ok {
		return value, ErrNoBusCtx
	}

	log.Trace().Str("path", path).Str("dest", dest).Str("property", prop).Msgf("Requesting property (as %T).", value)
	obj := busObj.conn.Object(dest, dbus.ObjectPath(path))

	res, err := obj.GetProperty(prop)
	if err != nil {
		return value, fmt.Errorf("%s: unable to retrieve property %s from %s: %w", bus.String(), prop, dest, err)
	}

	return VariantToValue[P](res), nil
}

// SetProp sets the specific property to the specified value.
func SetProp[P any](ctx context.Context, bus dbusType, path, dest, prop string, value P) error {
	busObj, ok := getBus(ctx, bus)
	if !ok {
		return ErrNoBusCtx
	}

	v := dbus.MakeVariant(value)
	obj := busObj.conn.Object(dest, dbus.ObjectPath(path))

	err := obj.SetProperty(prop, v)
	if err != nil {
		return fmt.Errorf("%s: unable to set property %s (%s) to %v: %w", bus.String(), prop, dest, value, err)
	}

	return nil
}

// GetData fetches data using the given method from D-Bus, as the provided type.
// If there is an error or the result cannot be stored in the given type, it
// will return an non-nil error. To execute a method, see Call. To get the value
// of a property, see GetProp.
func GetData[D any](ctx context.Context, bus dbusType, path, dest, method string, args ...any) (D, error) {
	var data D

	var err error

	busObj, ok := getBus(ctx, bus)
	if !ok {
		return data, ErrNoBusCtx
	}

	obj := busObj.conn.Object(dest, dbus.ObjectPath(path))

	if args != nil {
		err = obj.Call(method, 0, args...).Store(&data)
	} else {
		err = obj.Call(method, 0).Store(&data)
	}

	if err != nil {
		return data, fmt.Errorf("%s: unable to get data %s from %s: %w", bus.String(), method, dest, err)
	}

	return data, nil
}

// WatchBus will set up a channel on which D-Bus messages matching the given
// rules can be monitored. Typically, this is used to react when a certain
// property or signal with a given path and on a given interface, changes. The
// data returned in the channel will contain the signal (or property) that
// triggered the match, the path and the contents (what values actually
// changed).
//
//nolint:cyclop
func WatchBus(ctx context.Context, conditions *Watch) (chan Trigger, error) {
	var wg sync.WaitGroup

	bus, ok := getBus(ctx, conditions.Bus)
	if !ok {
		return nil, ErrNoBusCtx
	}

	matchers := conditions.Parse()
	if err := bus.conn.AddMatchSignalContext(ctx, matchers...); err != nil {
		return nil, fmt.Errorf("unable to add watch conditions (%w)", err)
	}

	signalCh := make(chan *dbus.Signal)
	outCh := make(chan Trigger)

	bus.conn.Signal(signalCh)
	bus.wg.Add(1)

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				bus.conn.RemoveSignal(signalCh)
				close(outCh)

				return
			case signal := <-signalCh:
				// If the signal is empty, ignore.
				if signal == nil {
					continue
				}
				// If this signal is not for our path, ignore.
				if conditions.Path != "" {
					if string(signal.Path) != conditions.Path {
						continue
					}
				}

				if conditions.PathNamespace != "" {
					if !strings.HasPrefix(string(signal.Path), conditions.PathNamespace) {
						continue
					}
				}
				// We have a match! Send the signal details back to the client
				// for further processing.
				log.Trace().
					Str("path", conditions.Path).
					Str("interface", conditions.Interface).
					Strs("names", conditions.Names).
					Interface("signal", signal).
					Msg("Dispatching D-Bus trigger.")
				outCh <- Trigger{
					Signal:  signal.Name,
					Path:    string(signal.Path),
					Content: signal.Body,
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		bus.wg.Done()
	}()

	return outCh, nil
}

func GetSessionPath(ctx context.Context) dbus.ObjectPath {
	usr, err := user.Current()
	if err != nil {
		return ""
	}

	sessions, err := GetData[[][]any](ctx, SystemBus, loginBasePath, loginBaseInterface, listSessionsMethod)
	if err != nil {
		return ""
	}

	for _, s := range sessions {
		if thisUser, ok := s[2].(string); ok && thisUser == usr.Username {
			if p, ok := s[4].(dbus.ObjectPath); ok {
				return p
			}
		}
	}

	return ""
}

// ParsePropertiesChanged treats the given signal body as matching the canonical
// org.freedesktop.DBus.PropertiesChanged signature and will parse it into a
// Properties structure that is easier to use. If the signal body cannot be
// parsed an error will be returned with details of the problem. Adapted from
// https://github.com/godbus/dbus/issues/201
//
//nolint:exhaustruct,mnd
func ParsePropertiesChanged(propsChanged []any) (*Properties, error) {
	props := &Properties{}

	var ok bool

	if len(propsChanged) != 3 {
		return nil, ErrNotPropChanged
	}

	props.Interface, ok = propsChanged[0].(string)
	if !ok {
		return nil, ErrParseInterface
	}

	props.Changed, ok = propsChanged[1].(map[string]dbus.Variant)
	if !ok {
		return nil, ErrParseNewProps
	}

	props.Invalidated, ok = propsChanged[2].([]string)
	if !ok {
		return nil, ErrParseOldProps
	}

	return props, nil
}

// ParseValueChange treats the given signal body as matching a value change of a
// property from an old value to a new value. It will parse the signal body into
// a Value object with old/new values of the given type. If there was a problem
// parsing the signal body, an error will be returned with details of the
// problem.
//
//nolint:exhaustruct,mnd
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
func VariantToValue[S any](variant dbus.Variant) S {
	var value S

	err := variant.Store(&value)
	if err != nil {
		log.Debug().Err(err).
			Msgf("Unable to convert dbus variant %v to type %T.", variant, value)

		return value
	}

	return value
}
