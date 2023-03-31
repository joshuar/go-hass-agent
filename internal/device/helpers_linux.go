package device

import (
	"os"
	"strconv"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

func ConnectSystemDBus() *dbus.Conn {

	conn, err := dbus.SystemBusPrivate()
	logging.CheckError(err)

	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}

	err = conn.Auth(methods)
	logging.CheckError(err)

	err = conn.Hello()
	if err != nil {
		conn.Close()
		return nil
	}
	return conn
}
