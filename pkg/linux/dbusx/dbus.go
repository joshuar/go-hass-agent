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

var ErrNoBus = errors.New("no D-Bus connection")

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
	Path      string
	Interface string
	Name      string
}

type Properties struct {
	Interface   string
	Changed     map[string]dbus.Variant
	Invalidated []string
}

type Bus struct {
	conn    *dbus.Conn
	busType dbusType
	wg      sync.WaitGroup
}

// NewBus sets up D-Bus connections and channels for receiving signals. It
// creates both a system and session bus connection.
func NewBus(ctx context.Context, t dbusType) (*Bus, error) {
	var conn *dbus.Conn
	var err error
	dbusCtx, cancelFunc := context.WithCancel(context.Background())
	switch t {
	case SessionBus:
		conn, err = dbus.ConnectSessionBus(dbus.WithContext(dbusCtx))
	case SystemBus:
		conn, err = dbus.ConnectSystemBus(dbus.WithContext(dbusCtx))
	}
	if err != nil {
		cancelFunc()
		return nil, err
	}
	b := &Bus{
		conn:    conn,
		busType: t,
	}
	go func() {
		defer conn.Close()
		defer cancelFunc()
		<-ctx.Done()
		b.wg.Wait()
	}()
	return b, nil
}

// busRequest contains properties for building different types of D-Bus requests.
type busRequest struct {
	bus          *Bus
	eventHandler func(*dbus.Signal)
	path         dbus.ObjectPath
	event        string
	dest         string
	match        []dbus.MatchOption
}

// NewBusRequest creates a new busRequest builder on the specified bus type
// (either a system or session bus). If it cannot connect to the specified bus,
// it will still return a busRequest object. Further builder functions should
// check whether there is a valid bus if appropriate.
func NewBusRequest(ctx context.Context, busType dbusType) *busRequest {
	if bus, ok := getBus(ctx, busType); !ok {
		log.Debug().Msg("No D-Bus connection present in context.")
		return &busRequest{}
	} else {
		return &busRequest{
			bus: bus,
		}
	}
}

// Path defines the D-Bus path on which a request will operate.
func (r *busRequest) Path(p dbus.ObjectPath) *busRequest {
	r.path = p
	return r
}

// Match defines D-Bus routing match rules on which a request will operate.
func (r *busRequest) Match(m []dbus.MatchOption) *busRequest {
	r.match = m
	return r
}

// Event defines an event on which a D-Bus request should match.
func (r *busRequest) Event(e string) *busRequest {
	r.event = e
	return r
}

// Handler defines a function that will handle a matched D-Bus signal.
func (r *busRequest) Handler(h func(*dbus.Signal)) *busRequest {
	r.eventHandler = h
	return r
}

// Destination defines the location/interface on a given D-Bus path for a request
// to operate.
func (r *busRequest) Destination(d string) *busRequest {
	r.dest = d
	return r
}

// GetProp uses the given request builder to fetch the specified property from
// D-Bus as the given type. If the property cannot be retrieved, a non-nil error
// is returned.
func GetProp[P any](req *busRequest, prop string) (P, error) {
	var value P
	if req == nil || req.bus == nil {
		return value, ErrNoBus
	}
	obj := req.bus.conn.Object(req.dest, req.path)
	res, err := obj.GetProperty(prop)
	if err != nil {
		return value, fmt.Errorf("unable to retrieve property %s from %s: %w", prop, req.dest, err)
	}
	return VariantToValue[P](res), nil
}

// SetProp sets the specific property to the specified value.
func SetProp[P any](req *busRequest, prop string, value P) error {
	if req == nil || req.bus == nil {
		return ErrNoBus
	}
	v := dbus.MakeVariant(value)
	obj := req.bus.conn.Object(req.dest, req.path)
	return obj.SetProperty(prop, v)
}

// GetData uses the given request builder to fetch D-Bus data from the given
// method, as the provided type. If there is an error or the result cannot be
// stored in the given type, it will return an non-nil error.
func GetData[D any](req *busRequest, method string, args ...any) (D, error) {
	var data D
	if req == nil || req.bus == nil {
		return data, ErrNoBus
	}
	obj := req.bus.conn.Object(req.dest, req.path)
	var err error
	if args != nil {
		err = obj.Call(method, 0, args...).Store(&data)
	} else {
		err = obj.Call(method, 0).Store(&data)
	}
	return data, err
}

// Call executes the given method in the builder and returns the error state.
func (r *busRequest) Call(method string, args ...any) error {
	if r.bus == nil {
		return ErrNoBus
	}
	obj := r.bus.conn.Object(r.dest, r.path)
	if args != nil {
		return r.busRequestError("call could not retrieve object", obj.Call(method, 0, args...).Err)
	}
	return obj.Call(method, 0).Err
}

// Watch takes a given busRequest and will create watch for signal events with
// the given watch conditions. It will send any signal events via the returned
// channel. Typically, this is used to react when a certain property or signal
// with a given path and on a given interface, changes. The data returned in the
// channel will contain the signal (or property) that triggered the match, the
// path and the contents (what values actually changed).
func (r *busRequest) Watch(ctx context.Context, watchConditions Watch) (chan Trigger, error) {
	var matchers []dbus.MatchOption
	matchers = append(matchers,
		dbus.WithMatchObjectPath(dbus.ObjectPath(watchConditions.Path)),
		dbus.WithMatchInterface(watchConditions.Interface),
	)
	if err := r.bus.conn.AddMatchSignalContext(ctx, matchers...); err != nil {
		return nil, r.busRequestError("unable to add watch conditions", err)
	}
	signalCh := make(chan *dbus.Signal)
	outCh := make(chan Trigger)
	r.bus.conn.Signal(signalCh)
	r.bus.wg.Add(1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				r.bus.conn.RemoveSignal(signalCh)
				close(signalCh)
				close(outCh)
				return
			case signal := <-signalCh:
				if signal.Path == dbus.ObjectPath(watchConditions.Path) && strings.Contains(signal.Name, watchConditions.Name) {
					outCh <- Trigger{
						Signal:  signal.Name,
						Path:    string(signal.Path),
						Content: signal.Body,
					}
				}
			}
		}
	}()
	go func() {
		wg.Wait()
		r.bus.wg.Done()
	}()
	return outCh, nil
}

func (r *busRequest) AddWatch(ctx context.Context) error {
	if r.bus == nil {
		return ErrNoBus
	}
	if err := r.bus.conn.AddMatchSignalContext(ctx, r.match...); err != nil {
		return err
	}
	signalCh := make(chan *dbus.Signal)
	r.bus.conn.Signal(signalCh)
	r.bus.wg.Add(1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				r.bus.conn.RemoveSignal(signalCh)
				close(signalCh)
				return
			case signal := <-signalCh:
				r.eventHandler(signal)
			}
		}
	}()
	go func() {
		wg.Wait()
		r.bus.wg.Done()
	}()
	return nil
}

func (r *busRequest) RemoveWatch(ctx context.Context) error {
	if r.bus == nil {
		return ErrNoBus
	}
	if err := r.bus.conn.RemoveMatchSignalContext(ctx, r.match...); err != nil {
		return r.busRequestError("unable to remove watch", err)
	}
	log.Trace().
		Str("path", string(r.path)).
		Str("dest", r.dest).
		Str("event", r.event).
		Msgf("Removed D-Bus signal.")
	return nil
}

// busRequestError wraps a lower level error while creating a busRequest into
// something more understandable.
func (r *busRequest) busRequestError(msg string, err error) error {
	return fmt.Errorf("%s: %s (%w)", r.bus.busType.String(), msg, err)
}

func GetSessionPath(ctx context.Context) dbus.ObjectPath {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	req := NewBusRequest(ctx, SystemBus).
		Path(loginBasePath).
		Destination(loginBaseInterface)

	sessions, err := GetData[[][]any](req, listSessionsMethod)
	if err != nil {
		return ""
	}
	for _, s := range sessions {
		if thisUser, ok := s[2].(string); ok && thisUser == u.Username {
			if p, ok := s[4].(dbus.ObjectPath); ok {
				return p
			}
		}
	}
	return ""
}

// ParsePropertiesChanged treats the given signal body as matching the canonical
// org.freedesktop.DBus.PropertiesChanged signature and will parse it into a
// Properties structure that is easier to use. Adapted from
// https://github.com/godbus/dbus/issues/201
func ParsePropertiesChanged(v []any) (*Properties, error) {
	props := &Properties{}
	var ok bool
	if len(v) != 3 {
		return nil, errors.New("signal contents do not appear to represent changed properties")
	}
	props.Interface, ok = v[0].(string)
	if !ok {
		return nil, errors.New("could not parse interface name")
	}
	props.Changed, ok = v[1].(map[string]dbus.Variant)
	if !ok {
		return nil, errors.New("could not parse changed properteis")
	}
	props.Invalidated, ok = v[2].([]string)
	if !ok {
		return nil, errors.New("could not parse invalidated properties")
	}
	return props, nil
}

// VariantToValue converts a dbus.Variant interface{} value into the specified
// Go native type. If the value is nil, then the return value will be the
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
