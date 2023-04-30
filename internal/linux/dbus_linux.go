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
	sessionBus dbusType = iota
	systemBus
)

type dbusType int

type bus struct {
	mu        sync.RWMutex
	conn      *dbus.Conn
	events    chan *dbus.Signal
	eventList map[string]func(*dbus.Signal)
	busType   dbusType
}

// newBus sets up DBus connections and channels for receiving signals. It creates both a system and session bus connection.
func newBus(ctx context.Context, t dbusType) *bus {
	var conn *dbus.Conn
	var err error
	switch t {
	case sessionBus:
		conn, err = dbus.ConnectSessionBus(dbus.WithContext(ctx))
	case systemBus:
		conn, err = dbus.ConnectSystemBus(dbus.WithContext(ctx))
	}
	if err != nil {
		log.Error().Stack().Err(err).
			Msg("Could not connect to bus")
		return nil
	} else {
		bus := &bus{
			conn:      conn,
			events:    make(chan *dbus.Signal),
			eventList: make(map[string]func(*dbus.Signal)),
			busType:   t,
		}
		conn.Signal(bus.events)
		go func() {
			defer bus.conn.RemoveSignal(bus.events)
			for signal := range bus.events {
				// bus.mu.RLock()
				if handlerFunc, ok := bus.eventList[string(signal.Path)]; ok {
					handlerFunc(signal)
				} else {
					for matchPath, handlerFunc := range bus.eventList {
						if strings.Contains(string(signal.Path), matchPath) {
							handlerFunc(signal)
						}
					}
				}
				// bus.mu.Unlock()
			}
		}()
		return bus
	}
}

// busRequest contains properties for building different types of DBus requests
type busRequest struct {
	bus          *bus
	path         dbus.ObjectPath
	match        []dbus.MatchOption
	event        string
	eventHandler func(*dbus.Signal)
	dest         string
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
			log.Debug().Caller().Err(err).
				Msgf("Unable to retrieve property %s (%s)", prop, r.dest)
			return dbus.MakeVariant(""), err
		}
		return res, nil
	} else {
		return dbus.MakeVariant(""), errors.New("no bus connection")
	}
}

// GetData fetches DBus data from the given method in the builder
func (r *busRequest) GetData(method string, args ...interface{}) *dbusData {
	if r.bus != nil {
		d := &dbusData{}
		obj := r.bus.conn.Object(r.dest, r.path)
		var err error
		if args != nil {
			err = obj.Call(method, 0, args...).Store(&d.data)
		} else {
			err = obj.Call(method, 0).Store(&d.data)
		}
		if err != nil {
			log.Debug().Err(err).Caller().
				Msgf("Unable to execute %s on %s (args: %s)", method, r.dest, args)
			return nil
		}
		return d
	} else {
		log.Debug().Caller().
			Msgf("no bus connection")
		return nil
	}
}

// AddWatch adds a DBus watch to the bus with the given options in the builder
func (r *busRequest) AddWatch() error {
	if err := r.bus.conn.AddMatchSignal(r.match...); err != nil {
		return err
	} else {
		log.Debug().Caller().
			Msgf("Adding watch on %s for %s", r.path, r.event)
		r.bus.mu.Lock()
		r.bus.eventList[string(r.path)] = r.eventHandler
		r.bus.mu.Unlock()
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

type DeviceAPI struct {
	dBusSystem, dBusSession *bus
}

// NewDeviceAPI sets up a DeviceAPI struct with appropriate DBus API endpoints
func NewDeviceAPI(ctx context.Context) *DeviceAPI {
	api := &DeviceAPI{
		dBusSystem:  newBus(ctx, systemBus),
		dBusSession: newBus(ctx, sessionBus),
	}
	if api.dBusSystem == nil && api.dBusSession == nil {
		return nil
	} else {
		// go api.monitorDBus(ctx)
		return api
	}
}

// SessionBusRequest creates a request builder for the session bus
func (d *DeviceAPI) SessionBusRequest() *busRequest {
	return &busRequest{
		bus: d.dBusSession,
	}
}

// SystemBusRequest creates a request builder for the system bus
func (d *DeviceAPI) SystemBusRequest() *busRequest {
	return &busRequest{
		bus: d.dBusSystem,
	}
}

// variantToValue converts a dbus.Variant type into the specified Go native
// type.
func variantToValue[S any](variant dbus.Variant) S {
	var value S
	err := variant.Store(&value)
	if err != nil {
		log.Debug().Err(err).
			Msgf("Unable to convert dbus variant to type.")
		return value
	}
	return value
}

// FindPortal is a helper function to work out which portal interface should be
// used for getting information on running apps.
func FindPortal() string {
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
	deviceAPI, err := FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to DBus.")
		return "localhost"
	}
	var dBusDest = "org.freedesktop.hostname1"
	var dBusPath = "/org/freedesktop/hostname1"
	hostnameFromDBus, err := deviceAPI.SystemBusRequest().
		Path(dbus.ObjectPath(dBusPath)).
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
	deviceAPI, err := FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to DBus.")
		return "", ""
	}
	var dBusDest = "org.freedesktop.hostname1"
	var dBusPath = "/org/freedesktop/hostname1"
	hwVendorFromDBus, err := deviceAPI.SystemBusRequest().
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
	hwModelFromDBus, err := deviceAPI.SystemBusRequest().
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

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// configKey is the key for DeviceAPI values in Contexts. It is
// unexported; clients use linux.NewContext and linux.FromContext
// instead of using this key directly.
var configKey key

// StoreAPIInContext returns a new Context that embeds a DeviceAPI.
func StoreAPIInContext(ctx context.Context, c *DeviceAPI) context.Context {
	return context.WithValue(ctx, configKey, c)
}

// FetchAPIFromContext returns the DeviceAPI stored in ctx, or an error if there
// is none
func FetchAPIFromContext(ctx context.Context) (*DeviceAPI, error) {
	if c, ok := ctx.Value(configKey).(*DeviceAPI); !ok {
		return nil, errors.New("no API in context")
	} else {
		return c, nil
	}
}
