// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"os"
	"os/user"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jaypipes/ghw"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/whichdistro"
)

type Device struct {
	appName       string
	appVersion    string
	hostname      string
	deviceID      string
	hwVendor      string
	hwModel       string
	distro        string
	distroVersion string
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
	return l.hwVendor
}

func (l *Device) Model() string {
	return l.hwModel
}

func (l *Device) OsName() string {
	return l.distro
}

func (l *Device) OsVersion() string {
	return l.distroVersion
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

func NewDevice(name, version string) *Device {
	dev := &Device{
		appName:    name,
		appVersion: version,
		deviceID:   getDeviceID(),
		hostname:   getHostname(),
	}

	osReleaseInfo, err := whichdistro.GetOSRelease()
	if err != nil {
		log.Warn().Err(err).Msg("Could not read /etc/os-release. Contact your distro vendor to implement this file.")
		dev.distro = "Unknown Distro"
		dev.distroVersion = "Unknown Version"
	} else {
		dev.distro = osReleaseInfo["ID"]
		dev.distroVersion = osReleaseInfo["VERSION_ID"]
	}

	dev.hwModel, dev.hwVendor = getHWProductInfo()

	return dev
}

func MQTTDevice() *mqtthass.Device {
	dev := NewDevice(preferences.AppName, preferences.AppVersion)
	return &mqtthass.Device{
		Name:         dev.DeviceName(),
		URL:          preferences.AppURL,
		SWVersion:    dev.OsVersion(),
		Manufacturer: dev.Manufacturer(),
		Model:        dev.Model(),
		Identifiers:  []string{dev.DeviceID()},
	}
}

// getDeviceID create a new device ID. It will be a randomly generated UUIDv4.
func getDeviceID() string {
	deviceID, err := uuid.NewV4()
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve a machine ID")
		return "unknown"
	}
	return deviceID.String()
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

func getHWProductInfo() (model, vendor string) {
	product, err := ghw.Product(ghw.WithDisableWarnings())
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve hardware information.")
		return "Unknown Product", "Unknown Vendor"
	}
	return product.Name, product.Vendor
}

// FindPortal is a helper function to work out which portal interface should be
// used for getting information on running apps.
func FindPortal() string {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	switch {
	case strings.Contains(desktop, "KDE"):
		return "org.freedesktop.impl.portal.desktop.kde"
	case strings.Contains(desktop, "GNOME"):
		return "org.freedesktop.impl.portal.desktop.gtk"
	default:
		return ""
	}
}
