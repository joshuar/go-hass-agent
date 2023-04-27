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

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const (
	sessionBus dbusType = iota
	systemBus
)

type dbusType int

type bus struct {
	conn    *dbus.Conn
	events  chan *dbus.Signal
	busType dbusType
}

type DeviceAPI struct {
	dBusSystem, dBusSession *bus
	WatchEvents             chan *DBusWatchRequest
}

// DBusWatchRequest contains all the information required to set-up a DBus match
// signal watcher.
type DBusWatchRequest struct {
	bus          dbusType
	path         dbus.ObjectPath
	match        []dbus.MatchOption
	event        string
	eventHandler func(*dbus.Signal)
}

// NewBus sets up DBus connections and channels for receiving signals. It creates both a system and session bus connection.
func NewBus(ctx context.Context, t dbusType) *bus {
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
		events := make(chan *dbus.Signal)
		conn.Signal(events)
		return &bus{
			conn:    conn,
			events:  events,
			busType: t,
		}
	}
}

func NewDeviceAPI(ctx context.Context) *DeviceAPI {
	api := &DeviceAPI{
		dBusSystem:  NewBus(ctx, systemBus),
		dBusSession: NewBus(ctx, sessionBus),
		WatchEvents: make(chan *DBusWatchRequest),
	}
	if api.dBusSystem == nil && api.dBusSession == nil {
		return nil
	} else {
		go api.monitorDBus(ctx)
		return api
	}
}

func (d *DeviceAPI) bus(t dbusType) *dbus.Conn {
	switch t {
	case sessionBus:
		if d.dBusSession != nil {
			return d.dBusSession.conn
		} else {
			return nil
		}
	case systemBus:
		if d.dBusSystem != nil {
			return d.dBusSystem.conn
		} else {
			return nil
		}
	default:
		log.Warn().Msg("Could not discern DBus bus type.")
		return nil
	}
}

// monitorDBus listens for DBus watch requests and ensures the appropriate
// signal watches are created. It will also dispatch to a handler function when
// a signal is matched.
func (d *DeviceAPI) monitorDBus(ctx context.Context) {
	events := make(map[dbusType]map[string]func(*dbus.Signal))
	watches := make(map[dbusType]*DBusWatchRequest)
	defer close(d.WatchEvents)
	// For each bus signal handler that exists, try to match first on an exact
	// path match, then try a substr match. Whichever matches, run the handler
	// function associated with it.
	if d.dBusSession != nil {
		events[sessionBus] = make(map[string]func(*dbus.Signal))
		defer d.dBusSession.conn.RemoveSignal(d.dBusSession.events)
		go func() {
			for sessionSignal := range d.dBusSession.events {
				if handlerFunc, ok := events[sessionBus][string(sessionSignal.Path)]; ok {
					handlerFunc(sessionSignal)
				} else {
					for matchPath, handlerFunc := range events[systemBus] {
						if strings.Contains(string(sessionSignal.Path), matchPath) {
							handlerFunc(sessionSignal)
						}
					}
				}
			}
		}()
	}
	if d.dBusSystem != nil {
		events[systemBus] = make(map[string]func(*dbus.Signal))
		defer d.dBusSystem.conn.RemoveSignal(d.dBusSystem.events)
		go func() {
			for systemSignal := range d.dBusSystem.events {
				if handlerFunc, ok := events[systemBus][string(systemSignal.Path)]; ok {
					handlerFunc(systemSignal)
				} else {
					for matchPath, handlerFunc := range events[systemBus] {
						if strings.Contains(string(systemSignal.Path), matchPath) {
							handlerFunc(systemSignal)
						}
					}
				}
			}
		}()
	}
	for {
		select {
		// When the context is finished/cancelled, try to clean up gracefully.
		case <-ctx.Done():
			log.Debug().Caller().Msg("Stopping DBus Monitor.")
			close(d.WatchEvents)
			d.dBusSession.conn.RemoveSignal(d.dBusSession.events)
			d.dBusSystem.conn.RemoveSignal(d.dBusSystem.events)
			return
		// When a new watch request is received, send it to the right DBus
		// connection and record it so it can be matched to a handler.
		case watch := <-d.WatchEvents:
			err := d.AddDBusWatch(watch.bus, watch.match)
			if err != nil {
				log.Debug().Err(err).Caller().
					Msgf("Could not add watch for %v.", watch.event)
			} else {
				events[watch.bus][string(watch.path)] = watch.eventHandler
				watches[watch.bus] = watch
			}
		}
	}
}

// AddDBusWatch will add a matcher to the specified bus monitoring for the
// specified path and interface. For adding dbus.MatchOptions, see the available
// ones here:
// https://dbus.freedesktop.org/doc/dbus-specification.html#message-bus-routing-match-rules
func (d *DeviceAPI) AddDBusWatch(t dbusType, matches []dbus.MatchOption) error {
	if err := d.bus(t).AddMatchSignal(matches...); err != nil {
		return err
	} else {
		return nil
	}
}

// RemoveDBusWatch will remove a matcher from the specified bus to stop it
// generating signals for that match.
func (d *DeviceAPI) RemoveDBusWatch(t dbusType, path dbus.ObjectPath, intr string) error {
	if err := d.bus(t).RemoveMatchSignal(
		dbus.WithMatchObjectPath(path),
		dbus.WithMatchInterface(intr),
	); err != nil {
		return err
	} else {
		return nil
	}
}

func (d *DeviceAPI) GetDBusObject(t dbusType, dest string, path dbus.ObjectPath) dbus.BusObject {
	if d.bus(t) != nil {
		return d.bus(t).Object(dest, path)
	} else {
		return nil
	}
}

// GetDBusProp will retrieve the specified property value from the given path
// and destination.
func (d *DeviceAPI) GetDBusProp(t dbusType, dest string, path dbus.ObjectPath, prop string) (dbus.Variant, error) {
	if d.bus(t) != nil {
		obj := d.bus(t).Object(dest, path)
		res, err := obj.GetProperty(prop)
		if err != nil {
			// log.Debug().Caller().Err(err).
			// 	Msgf("Unable to retrieve property %s (%s)", prop, dest)
			return dbus.MakeVariant(""), err
		}
		return res, nil
	} else {
		return dbus.MakeVariant(""), errors.New("no bus connection")
	}
}

func (d *DeviceAPI) GetDBusDataAsMap(t dbusType, dest string, path dbus.ObjectPath, method string, args ...interface{}) (map[string]dbus.Variant, error) {
	if d.bus(t) != nil {
		obj := d.bus(t).Object(dest, path)
		var data map[string]dbus.Variant
		var err error
		if args != nil {
			err = obj.Call(method, 0, args...).Store(&data)
		} else {
			err = obj.Call(method, 0).Store(&data)
		}
		if err != nil {
			log.Error().Err(err).
				Msgf("Unable to execute %s on %s (args: %s)", method, dest, args)
			return nil, err
		}
		return data, nil
	} else {
		return nil, errors.New("no bus connection")
	}
}

func (d *DeviceAPI) GetDBusDataAsList(t dbusType, dest string, path dbus.ObjectPath, method string, args ...interface{}) ([]string, error) {
	if d.bus(t) != nil {
		obj := d.bus(t).Object(dest, path)
		var data []string
		var err error
		if args != nil {
			err = obj.Call(method, 0, args...).Store(&data)
		} else {
			err = obj.Call(method, 0).Store(&data)
		}
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		return nil, errors.New("no bus connection")
	}
}

func (d *DeviceAPI) GetDBusData(t dbusType, dest string, path dbus.ObjectPath, method string, args ...interface{}) (interface{}, error) {
	if d.bus(t) != nil {
		obj := d.bus(t).Object(dest, path)
		var data interface{}
		var err error
		if args != nil {
			err = obj.Call(method, 0, args...).Store(&data)
		} else {
			err = obj.Call(method, 0).Store(&data)
		}
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		return nil, errors.New("no bus connection")
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

func GetHostname(ctx context.Context) string {
	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor network.")
		return "localhost"
	}
	var dBusDest = "org.freedesktop.hostname1"
	var dBusPath = "/org/freedesktop/hostname1"
	hostnameFromDBus, err := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		dbus.ObjectPath(dBusPath),
		dBusDest+".Hostname")
	if err != nil {
		return "localhost"
	} else {
		return string(variantToValue[[]uint8](hostnameFromDBus))
	}
}

func GetHardwareDetails(ctx context.Context) (string, string) {
	var vendor, model string
	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor network.")
		return "", ""
	}
	var dBusDest = "org.freedesktop.hostname1"
	var dBusPath = "/org/freedesktop/hostname1"
	hwVendorFromDBus, err := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		dbus.ObjectPath(dBusPath),
		dBusDest+".HardwareVendor")
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
	hwModelFromDBus, err := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		dbus.ObjectPath(dBusPath),
		dBusDest+".HardwareModel")
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

// NewContext returns a new Context that carries value c.
func NewContext(ctx context.Context, c *DeviceAPI) context.Context {
	return context.WithValue(ctx, configKey, c)
}

// FromContext returns the value stored in ctx, if any.
func FromContext(ctx context.Context) (*DeviceAPI, bool) {
	c, ok := ctx.Value(configKey).(*DeviceAPI)
	return c, ok
}
