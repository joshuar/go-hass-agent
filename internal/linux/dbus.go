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
	"github.com/joshuar/go-hass-agent/internal/api"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=dbusType -output dbusTypesStringer.go -linecomment
const (
	SessionBus dbusType = iota // session
	SystemBus                  // system
)

type dbusType int

type Bus struct {
	conn           *dbus.Conn
	signals        chan *dbus.Signal
	signalMatchers map[string]func(*dbus.Signal)
	matchRequests  chan signalMatcher
	busType        dbusType
	mu             sync.RWMutex
}

func (bus *Bus) signalHandler(ctx context.Context) {
	bus.conn.Signal(bus.signals)
	defer bus.conn.RemoveSignal(bus.signals)
	for {
		select {
		case <-ctx.Done():
			return
		case signal := <-bus.signals:
			// bus.mu.RLock()
			// defer bus.mu.Unlock()
			for matchPath, handlerFunc := range bus.signalMatchers {
				if strings.Contains(string(signal.Path), matchPath) {
					handlerFunc(signal)
				}
			}
		case request := <-bus.matchRequests:
			// bus.mu.Lock()
			// defer bus.mu.Unlock()
			bus.signalMatchers[request.match] = request.handler
		}
	}
}

// NewBus sets up DBus connections and channels for receiving signals. It creates both a system and session bus connection.
func NewBus(ctx context.Context, t dbusType) *Bus {
	var conn *dbus.Conn
	var err error
	switch t {
	case SessionBus:
		conn, err = dbus.ConnectSessionBus(dbus.WithContext(ctx))
	case SystemBus:
		conn, err = dbus.ConnectSystemBus(dbus.WithContext(ctx))
	}
	if err != nil {
		log.Error().Err(err).
			Msgf("Could not connect to %s bus.", t.String())
		return nil
	} else {
		bus := &Bus{
			conn:           conn,
			signals:        make(chan *dbus.Signal),
			signalMatchers: make(map[string]func(*dbus.Signal)),
			matchRequests:  make(chan signalMatcher),
			busType:        t,
		}
		go bus.signalHandler(ctx)
		return bus
	}
}

type signalMatcher struct {
	handler func(*dbus.Signal)
	match   string
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
	deviceAPI, err := api.FetchAPIFromContext(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Could not retrieve device API from context.")
		return nil
	}
	dbusAPI := api.GetAPIEndpoint[*Bus](deviceAPI, busType.String())
	return &busRequest{
		bus: dbusAPI,
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

// AddWatch adds a DBus watch to the bus with the given options in the builder
func (r *busRequest) AddWatch(ctx context.Context) error {
	if r.bus == nil {
		return errors.New("no bus connection")
	}
	if err := r.bus.conn.AddMatchSignalContext(ctx, r.match...); err != nil {
		return err
	} else {
		log.Trace().Caller().
			Msgf("Adding watch on %s for %s", r.path, r.event)
		r.bus.matchRequests <- signalMatcher{
			match:   string(r.path),
			handler: r.eventHandler,
		}
	}
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
		log.Warn().Msg("Unsupported desktop/window environment.")
		return ""
	}
}

// GetHostname will try to fetch the hostname of the device from DBus. Failing
// that, it will default to using "localhost"
func GetHostname(ctx context.Context) string {
	var dBusDest = "org.freedesktop.hostname1"
	hostnameFromDBus, err := NewBusRequest(ctx, SystemBus).
		Path(dbus.ObjectPath("/org/freedesktop/hostname1")).
		Destination(dBusDest).
		GetProp(dBusDest + ".Hostname")
	if err != nil {
		return "localhost"
	} else {
		return string(variantToValue[[]uint8](hostnameFromDBus))
	}
}

// GetHardwareDetails will try to get a hardware vendor and model from DBus.
// Failing that, it will try to read them from the /sys filesystem. If that
// fails, it returns empty strings for these values
func GetHardwareDetails(ctx context.Context) (string, string) {
	var vendor, model string
	var dBusDest = "org.freedesktop.hostname1"
	var dBusPath = "/org/freedesktop/hostname1"
	hwVendorFromDBus, err := NewBusRequest(ctx, SystemBus).
		Path(dbus.ObjectPath(dBusPath)).
		Destination(dBusDest).
		GetProp(dBusDest + ".HardwareVendor")
	if err != nil {
		hwVendor, err := os.ReadFile("/sys/devices/virtual/dmi/id/board_vendor")
		if err != nil {
			vendor = "Unknown Vendor"
		} else {
			vendor = strings.TrimSpace(string(hwVendor))
		}
	} else {
		vendor = string(variantToValue[[]uint8](hwVendorFromDBus))
	}
	hwModelFromDBus, err := NewBusRequest(ctx, SystemBus).
		Path(dbus.ObjectPath(dBusPath)).
		Destination(dBusDest).
		GetProp(dBusDest + ".HardwareVendor")
	if err != nil {
		hwModel, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name")
		if err != nil {
			model = "Unknown Vendor"
		} else {
			model = strings.TrimSpace(string(hwModel))
		}
	} else {
		model = string(variantToValue[[]uint8](hwModelFromDBus))
	}
	return vendor, model
}
