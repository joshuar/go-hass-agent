package device

import (
	"context"
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const (
	monitorDBusMethod          = "org.freedesktop.DBus.Monitoring.BecomeMonitor"
	sessionBus        dbusType = iota
	systemBus
)

type dbusType int
type bus struct {
	conn    *dbus.Conn
	events  chan *dbus.Signal
	busType dbusType
}

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
		log.Debug().Caller().
			Msgf("Could not connect to bus: %v", err)
		return nil
	}
	events := make(chan *dbus.Signal)
	conn.Signal(events)
	return &bus{
		conn:    conn,
		events:  events,
		busType: t,
	}
}

type DBusSignalMatch struct {
	path dbus.ObjectPath
	intr string
}

type DBusWatchRequest struct {
	bus          dbusType
	match        DBusSignalMatch
	event        string
	eventHandler func(*dbus.Signal)
}

type deviceAPI struct {
	dBusSystem, dBusSession *bus
	WatchEvents             chan *DBusWatchRequest
}

func NewContextWithDeviceAPI(ctx context.Context) context.Context {
	deviceAPI := &deviceAPI{
		dBusSystem:  NewBus(ctx, systemBus),
		dBusSession: NewBus(ctx, sessionBus),
		WatchEvents: make(chan *DBusWatchRequest),
	}
	go deviceAPI.monitorDBus(ctx)
	deviceCtx := NewContext(ctx, deviceAPI)
	return deviceCtx
}

func (d *deviceAPI) bus(t dbusType) *dbus.Conn {
	switch t {
	case sessionBus:
		return d.dBusSession.conn
	case systemBus:
		return d.dBusSystem.conn
	default:
		return nil
	}
}

func (d *deviceAPI) monitorDBus(ctx context.Context) {
	events := make(map[dbusType]map[string]func(*dbus.Signal))
	events[sessionBus] = make(map[string]func(*dbus.Signal))
	events[systemBus] = make(map[string]func(*dbus.Signal))
	for {
		select {
		case <-ctx.Done():
			close(d.WatchEvents)
			close(d.dBusSession.events)
			close(d.dBusSystem.events)
		case watch := <-d.WatchEvents:
			d.AddDBusWatch(watch.bus, watch.match)
			events[watch.bus][watch.event] = watch.eventHandler
			log.Debug().Caller().Msgf("Added watch for %v on %v", watch.event, watch.match.path)
		case systemSignal := <-d.dBusSystem.events:
			log.Debug().Msgf("Recieved system event: %v", systemSignal.Name)
			if handlerFunc, ok := events[systemBus][systemSignal.Name]; ok {
				handlerFunc(systemSignal)
			}
		case sessionSignal := <-d.dBusSession.events:
			log.Debug().Msgf("Recieved session event: %v", sessionSignal.Name)
			if handlerFunc, ok := events[sessionBus][sessionSignal.Name]; ok {
				handlerFunc(sessionSignal)
			}
		}
	}
}

// watchDBusSignal will add a matcher to the specified bus monitoring for
// the specified path and interface.
func (d *deviceAPI) AddDBusWatch(t dbusType, m DBusSignalMatch) error {
	if err := d.bus(t).AddMatchSignal(
		dbus.WithMatchObjectPath(m.path),
		dbus.WithMatchInterface(m.intr),
	); err != nil {
		return err
	} else {
		return nil
	}
}

func (d *deviceAPI) GetDBusProp(t dbusType, dest string, path dbus.ObjectPath, prop string) (dbus.Variant, error) {
	obj := d.bus(t).Object(dest, path)
	res, err := obj.GetProperty(prop)
	if err != nil {
		return dbus.MakeVariant(""), err
	}
	return res, nil
}

func (d *deviceAPI) GetDBusData(t dbusType, dest string, path dbus.ObjectPath, method string, args ...interface{}) (interface{}, error) {
	obj := d.bus(t).Object(dest, path)
	var data interface{}
	var err error
	if args != nil {
		err = obj.Call(method, 0, args).Store(&data)
	} else {
		err = obj.Call(method, 0).Store(&data)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return data, nil
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// configKey is the key for device.deviceAPI values in Contexts. It is
// unexported; clients use device.NewContext and device.FromContext
// instead of using this key directly.
var configKey key

// NewContext returns a new Context that carries value d.
func NewContext(ctx context.Context, d *deviceAPI) context.Context {
	return context.WithValue(ctx, configKey, d)
}

// FromContext returns the deviceAPI value stored in ctx, if any.
func FromContext(ctx context.Context) (*deviceAPI, bool) {
	c, ok := ctx.Value(configKey).(*deviceAPI)
	return c, ok
}

func FindPortal() string {
	switch os.Getenv("XDG_CURRENT_DESKTOP") {
	case "KDE":
		return "org.freedesktop.impl.portal.desktop.kde"
	case "GNOME":
		return "org.freedesktop.impl.portal.desktop.kde"
	default:
		log.Warn().Msg("Unsupported desktop/window environment. No app logging available.")
		return ""
	}
}
