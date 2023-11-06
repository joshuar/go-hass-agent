// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const (
	SessionBus dbusType = iota // session
	SystemBus                  // system
)

type dbusType int

type Bus struct {
	conn    *dbus.Conn
	busType dbusType
	wg      sync.WaitGroup
}

// NewBus sets up DBus connections and channels for receiving signals. It creates both a system and session bus connection.
func NewBus(ctx context.Context, t dbusType) *Bus {
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
		log.Error().Err(err).Msg("Could not connect to bus.")
		cancelFunc()
		return nil
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
	return b
}

// busRequest contains properties for building different types of DBus requests
type busRequest struct {
	bus          *Bus
	eventHandler func(*dbus.Signal)
	path         dbus.ObjectPath
	event        string
	dest         string
	match        []dbus.MatchOption
}

func NewBusRequest(ctx context.Context, busType dbusType) *busRequest {
	if bus, ok := getBus(ctx, busType); !ok {
		log.Warn().Msg("No D-Bus connection present in context.")
		return &busRequest{}
	} else {
		return &busRequest{
			bus: bus,
		}
	}
}

// Path defines the DBus path on which a request will operate
func (r *busRequest) Path(p dbus.ObjectPath) *busRequest {
	r.path = p
	return r
}

// Match defines DBus routing match rules on which a request will operate
func (r *busRequest) Match(m []dbus.MatchOption) *busRequest {
	r.match = m
	return r
}

// Event defines an event on which a DBus request should match
func (r *busRequest) Event(e string) *busRequest {
	r.event = e
	return r
}

// Handler defines a function that will handle a matched DBus signal
func (r *busRequest) Handler(h func(*dbus.Signal)) *busRequest {
	r.eventHandler = h
	return r
}

// Destination defines the location/interface on a given DBus path for a request
// to operate
func (r *busRequest) Destination(d string) *busRequest {
	r.dest = d
	return r
}

// GetProp fetches the specified property from DBus with the options specified
// in the builder
func (r *busRequest) GetProp(prop string) (dbus.Variant, error) {
	if r.bus != nil {
		obj := r.bus.conn.Object(r.dest, r.path)
		res, err := obj.GetProperty(prop)
		if err != nil {
			log.Warn().Err(err).
				Msgf("Unable to retrieve property %s (%s)", prop, r.dest)
			return dbus.MakeVariant(""), err
		}
		return res, nil
	} else {
		return dbus.MakeVariant(""), errors.New("no bus connection")
	}
}

// SetProp sets the specific property to the specified value
func (r *busRequest) SetProp(prop string, value dbus.Variant) error {
	if r.bus != nil {
		obj := r.bus.conn.Object(r.dest, r.path)
		return obj.SetProperty(prop, value)
	}
	return errors.New("no bus connection")
}

// GetData fetches DBus data from the given method in the builder
func (r *busRequest) GetData(method string, args ...interface{}) *dbusData {
	d := new(dbusData)
	if r.bus != nil {
		obj := r.bus.conn.Object(r.dest, r.path)
		var err error
		if args != nil {
			err = obj.Call(method, 0, args...).Store(&d.data)
		} else {
			err = obj.Call(method, 0).Store(&d.data)
		}
		if err != nil {
			log.Warn().Err(err).
				Msgf("Unable to execute %s on %s (args: %s)", method, r.dest, args)
		}
		return d
	} else {
		log.Error().Msg("No bus connection.")
		return d
	}
}

// Call executes the given method in the builder and returns the error state
func (r *busRequest) Call(method string, args ...interface{}) error {
	if r.bus != nil {
		obj := r.bus.conn.Object(r.dest, r.path)
		if args != nil {
			return obj.Call(method, 0, args...).Err
		} else {
			return obj.Call(method, 0).Err
		}
	} else {
		return errors.New("no bus connection")
	}
}

func (r *busRequest) AddWatch(ctx context.Context) error {
	if r.bus == nil {
		return errors.New("no bus connection")
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
				if err := r.bus.conn.RemoveMatchSignal(r.match...); err != nil {
					log.Warn().Err(err).
						Str("path", string(r.path)).
						Str("dest", r.dest).
						Str("event", r.event).
						Msg("Failed to remove D-Bus watch.")
					return
				}
				log.Debug().
					Str("path", string(r.path)).
					Str("dest", r.dest).
					Str("event", r.event).
					Msgf("Stopped D-Bus watch.")
				return
			case signal := <-signalCh:
				if strings.Contains(string(signal.Path), string(r.path)) {
					r.eventHandler(signal)
				}
			}
		}
	}()
	log.Debug().
		Str("path", string(r.path)).
		Str("dest", r.dest).
		Str("event", r.event).
		Msgf("Added D-Bus watch.")
	go func() {
		wg.Wait()
		r.bus.wg.Done()
	}()
	return nil
}

func (r *busRequest) RemoveWatch(ctx context.Context) error {
	if r.bus == nil {
		return errors.New("no bus connection")
	}
	if err := r.bus.conn.RemoveMatchSignalContext(ctx, r.match...); err != nil {
		return err
	}
	log.Trace().
		Str("path", string(r.path)).
		Str("dest", r.dest).
		Str("event", r.event).
		Msgf("Removed D-Bus signal.")
	return nil
}

type dbusData struct {
	data interface{}
}

// AsVariantMap formats DBus data as a map[string]dbus.Variant
func (d *dbusData) AsVariantMap() map[string]dbus.Variant {
	if d.data != nil {
		wanted := make(map[string]dbus.Variant)
		for k, v := range d.data.(map[string]interface{}) {
			wanted[k] = dbus.MakeVariant(v)
		}
		return wanted
	} else {
		return nil
	}
}

// AsStringMap formats DBus data as a map[string]string
func (d *dbusData) AsStringMap() map[string]string {
	if d.data != nil {
		return d.data.(map[string]string)
	} else {
		return nil
	}
}

// AsObjectPathList formats DBus data as a []dbus.ObjectPath
func (d *dbusData) AsObjectPathList() []dbus.ObjectPath {
	if d.data != nil {
		return d.data.([]dbus.ObjectPath)
	} else {
		return nil
	}
}

// AsStringList formats DBus data as a []string
func (d *dbusData) AsStringList() []string {
	if d.data != nil {
		return d.data.([]string)
	} else {
		return nil
	}
}

// AsObjectPath formats DBus data as a dbus.ObjectPath
func (d *dbusData) AsObjectPath() dbus.ObjectPath {
	if d.data != nil {
		return d.data.(dbus.ObjectPath)
	} else {
		return ""
	}
}

// AsRawInterface formats DBus data as a plain interface{}
func (d *dbusData) AsRawInterface() interface{} {
	if d.data != nil {
		return d.data
	} else {
		return nil
	}
}

// variantToValue converts a dbus.Variant type into the specified Go native
// type.
func variantToValue[S any](variant dbus.Variant) S {
	var value S
	err := variant.Store(&value)
	if err != nil {
		log.Warn().Err(err).
			Msgf("Unable to convert dbus variant %v to type %T.", variant, value)
		return value
	}
	return value
}

// findPortal is a helper function to work out which portal interface should be
// used for getting information on running apps.
func findPortal() string {
	switch os.Getenv("XDG_CURRENT_DESKTOP") {
	case "KDE":
		return "org.freedesktop.impl.portal.desktop.kde"
	case "GNOME":
		return "org.freedesktop.impl.portal.desktop.kde"
	default:
		return ""
	}
}
