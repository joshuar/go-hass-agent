// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"encoding/json"
	"os"
	"os/user"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
)

type Device struct {
	appName    string
	appVersion string
	hostname   string
	deviceID   string
}

// Setup returns a new Context that contains the D-Bus API.
func (l *Device) Setup(ctx context.Context) context.Context {
	return dbushelpers.Setup(ctx)
}

func (l *Device) AppName() string {
	return l.appName
}

func (l *Device) AppVersion() string {
	return l.appVersion
}

func (l *Device) AppID() string {
	// Use the current user's username to construct an app ID.
	currentUser, err := user.Current()
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve current user details.")
		return l.appName + "-unknown"
	}
	return l.appName + "-" + currentUser.Username
}

func (l *Device) DeviceName() string {
	shortHostname, _, _ := strings.Cut(l.hostname, ".")
	return shortHostname
}

func (l *Device) DeviceID() string {
	return l.deviceID
}

func (l *Device) Manufacturer() string {
	return getHWVendor()
}

func (l *Device) Model() string {
	return getHWModel()
}

func (l *Device) OsName() string {
	_, osRelease, _, err := host.PlatformInformation()
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve distribution details.")
		return "Unknown OS"
	}
	return osRelease
}

func (l *Device) OsVersion() string {
	_, _, osVersion, err := host.PlatformInformation()
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve version details.")
		return "Unknown Version"
	}
	return osVersion
}

func (l *Device) SupportsEncryption() bool {
	return false
}

func (l *Device) AppData() any {
	return &struct {
		PushWebsocket bool `json:"push_websocket_channel"`
	}{
		PushWebsocket: true,
	}
}

func (l *Device) MarshalJSON() ([]byte, error) {
	return json.Marshal(&api.RegistrationRequest{
		DeviceID:           l.DeviceID(),
		AppID:              l.AppID(),
		AppName:            l.AppName(),
		AppVersion:         l.AppVersion(),
		DeviceName:         l.DeviceName(),
		Manufacturer:       l.Manufacturer(),
		Model:              l.Model(),
		OsName:             l.OsName(),
		OsVersion:          l.OsVersion(),
		SupportsEncryption: l.SupportsEncryption(),
		AppData:            l.AppData(),
	})
}

func NewDevice(name, version string) *Device {
	var deviceID string
	var err error
	deviceID, err = host.HostID()
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve a machine ID")
		deviceID = "unknown"
	}
	return &Device{
		appName:    name,
		appVersion: version,
		deviceID:   deviceID,
		hostname:   getHostname(),
	}
}

// getHostname retrieves the hostname of the device running the agent, or
// localhost if that doesn't work.
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve hostname. Using 'localhost'.")
		return "localhost"
	}
	return hostname
}

// getHWVendor will try to retrieve the vendor from the sysfs filesystem. It
// will return "Unknown Vendor" if unsuccessful.
// Reference: https://github.com/ansible/ansible/blob/devel/lib/ansible/module_utils/facts/hardware/linux.py
func getHWVendor() string {
	hwVendor, err := os.ReadFile("/sys/devices/virtual/dmi/id/board_vendor")
	if err != nil {
		return "Unknown Vendor"
	}
	return strings.TrimSpace(string(hwVendor))
}

// getHWModel will try to retrieve the hardware model from the sysfs filesystem. It
// will return "Unknown Model" if unsuccessful.
// Reference: https://github.com/ansible/ansible/blob/devel/lib/ansible/module_utils/facts/hardware/linux.py
func getHWModel() string {
	hwModel, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name")
	if err != nil {
		return "Unknown Model"
	}
	return strings.TrimSpace(string(hwModel))
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
