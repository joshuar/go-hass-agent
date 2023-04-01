package device

import (
	"context"
	"os"
	"strconv"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const (
	monitorDBusMethod = "org.freedesktop.DBus.Monitoring.BecomeMonitor"
)

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
