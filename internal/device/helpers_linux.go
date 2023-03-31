package device

import (
	"os"
	"strconv"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

func ConnectSystemDBus() (*dbus.Conn, error) {

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		log.Debug().Caller().
			Msg("Failed to connect to private system bus.")
		return nil, err
	}

	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}

	err = conn.Auth(methods)
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

func ConnectSessionBus() (*dbus.Conn, error) {
	return dbus.ConnectSessionBus()
}
