package device

import (
	"context"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const (
	monitorDBusMethod          = "org.freedesktop.DBus.Monitoring.BecomeMonitor"
	sessionBus        dbusType = iota
	systemBus
)

type dbusType int
type deviceAPI struct {
	dBusSystem                          *dbus.Conn
	dBusSession                         *dbus.Conn
	DBusSessionEvents, DBusSystemEvents chan *dbus.Signal
	WatchEvents                         chan *DBusWatchData
}

type DBusSignal struct {
	path dbus.ObjectPath
	intr string
}

type DBusWatchData struct {
	bus          dbusType
	signal       DBusSignal
	event        string
	eventHandler func(*dbus.Signal)
}

func (d *deviceAPI) DBusConn(t dbusType) *dbus.Conn {
	switch t {
	case sessionBus:
		return d.dBusSession
	case systemBus:
		return d.dBusSystem
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
			close(d.DBusSessionEvents)
			close(d.DBusSessionEvents)
		case watch := <-d.WatchEvents:
			d.WatchDBusSignal(watch.bus, watch.signal)
			events[watch.bus][watch.event] = watch.eventHandler
			log.Debug().Caller().Msgf("Added watch for %v", watch)
			spew.Dump(events)
		case systemSignal := <-d.DBusSystemEvents:
			if handlerFunc, ok := events[systemBus][systemSignal.Name]; ok {
				handlerFunc(systemSignal)
			}
		case sessionSignal := <-d.DBusSessionEvents:
			log.Debug().Msgf("Recieved session event: %v", sessionSignal.Name)
			if handlerFunc, ok := events[sessionBus][sessionSignal.Name]; ok {
				handlerFunc(sessionSignal)
			}
		}
	}
}

func (d *deviceAPI) WatchDBusSignal(t dbusType, s DBusSignal) error {
	if err := d.DBusConn(t).AddMatchSignal(
		dbus.WithMatchObjectPath(s.path),
		dbus.WithMatchInterface(s.intr),
	); err != nil {
		return err
	} else {
		return nil
	}
}

func NewContextWithDeviceAPI(ctx context.Context) context.Context {
	system, err := dbus.ConnectSystemBus(dbus.WithContext(ctx))
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not connect to system bus: %v", err)
	}
	systemEvents := make(chan *dbus.Signal)
	system.Signal(systemEvents)

	session, err := dbus.ConnectSessionBus(dbus.WithContext(ctx))
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not connect to session bus: %v", err)
	}
	sessionEvents := make(chan *dbus.Signal)
	session.Signal(sessionEvents)

	deviceAPI := &deviceAPI{
		dBusSystem:        system,
		DBusSystemEvents:  systemEvents,
		dBusSession:       session,
		DBusSessionEvents: sessionEvents,
		WatchEvents:       make(chan *DBusWatchData),
	}
	go deviceAPI.monitorDBus(ctx)
	deviceCtx := NewContext(ctx, deviceAPI)
	return deviceCtx
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

func DBusConnectSystem(ctx context.Context) (*dbus.Conn, error) {
	conn, err := dbus.SystemBusPrivate(dbus.WithContext(ctx))
	if err != nil {
		log.Debug().Caller().
			Msg("Failed to connect to private system bus.")
		conn.Close()
		return nil, err
	}

	err = conn.Auth([]dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))})
	if err != nil {
		log.Debug().Caller().
			Msg("Failed to authenticate to private system bus.")
		conn.Close()
		return nil, err
	}

	err = conn.Hello()
	if err != nil {
		log.Debug().Caller().
			Msg("Failed to send Hello call.")
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func DBusConnectSession(ctx context.Context) (*dbus.Conn, error) {
	return dbus.ConnectSessionBus(dbus.WithContext(ctx))
}

func DBusBecomeMonitor(conn *dbus.Conn, rules []string, flag uint) error {
	call := conn.BusObject().Call(monitorDBusMethod, 0, rules, flag)
	return call.Err
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
