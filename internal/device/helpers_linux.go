package device

import (
	"context"
	"fmt"
	"os"

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
	*SensorInfo
	dBusSystem, dBusSession *bus
	WatchEvents             chan *DBusWatchRequest
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
	watches := make(map[dbusType]*DBusWatchRequest)
	defer close(d.WatchEvents)
	defer d.dBusSession.conn.RemoveSignal(d.dBusSession.events)
	defer d.dBusSystem.conn.RemoveSignal(d.dBusSystem.events)
	for {
		select {
		case <-ctx.Done():
			log.Debug().Caller().Msg("Stopping DBus Monitor.")
			for bus, request := range watches {
				d.RemoveDBusWatch(bus, request)
			}
			close(d.WatchEvents)
			d.dBusSession.conn.RemoveSignal(d.dBusSession.events)
			d.dBusSystem.conn.RemoveSignal(d.dBusSystem.events)
			return
		case watch := <-d.WatchEvents:
			d.AddDBusWatch(watch.bus, watch.match)
			events[watch.bus][watch.event] = watch.eventHandler
			watches[watch.bus] = watch
			log.Debug().Caller().Msgf("Added watch for %v on %v", watch.event, watch.match.path)
		case systemSignal := <-d.dBusSystem.events:
			if handlerFunc, ok := events[systemBus][systemSignal.Name]; ok {
				handlerFunc(systemSignal)
			}
		case sessionSignal := <-d.dBusSession.events:
			if handlerFunc, ok := events[sessionBus][sessionSignal.Name]; ok {
				handlerFunc(sessionSignal)
			}
		}
	}
}

// AddDBusWatch will add a matcher to the specified bus monitoring for
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

// RemoveDBusWatch will remove a matcher from the specified bus to stop it
// generating signals for that match.
func (d *deviceAPI) RemoveDBusWatch(t dbusType, w *DBusWatchRequest) error {
	if err := d.bus(t).RemoveMatchSignal(
		dbus.WithMatchObjectPath(w.match.path),
		dbus.WithMatchInterface(w.match.intr),
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

func SetupContext(ctx context.Context) context.Context {
	deviceAPI := &deviceAPI{
		SensorInfo:  NewSensorInfo(),
		dBusSystem:  NewBus(ctx, systemBus),
		dBusSession: NewBus(ctx, sessionBus),
		WatchEvents: make(chan *DBusWatchRequest),
	}
	go deviceAPI.monitorDBus(ctx)
	deviceAPI.SensorInfo.Add("Battery", BatteryUpdater)
	deviceAPI.SensorInfo.Add("Apps", AppUpdater)
	deviceCtx := NewContext(ctx, deviceAPI)
	return deviceCtx
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
		log.Warn().Msg("Unsupported desktop/window environment. No app logging available.")
		return ""
	}
}
