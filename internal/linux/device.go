// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver

package linux

import (
	"os"
	"strings"
	"syscall"

	"github.com/gofrs/uuid/v5"
	"github.com/iancoleman/strcase"
	"github.com/jaypipes/ghw"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/whichdistro"
)

const (
	unknownVendor        = "Unknown Vendor"
	unknownModel         = "Unknown Model"
	unknownDistro        = "Unknown Distro"
	unknownDistroVersion = "Unknown Version"
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
	return strcase.ToSnake(l.appName)
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
	dev.distro, dev.distroVersion = GetDistroID()
	dev.hwModel, dev.hwVendor = getHWProductInfo()
	return dev
}

func MQTTDevice() *mqtthass.Device {
	dev := NewDevice(preferences.AppName, preferences.AppVersion)
	prefs, err := preferences.Load()
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve preferences.")
	}
	return &mqtthass.Device{
		Name:         dev.DeviceName(),
		URL:          preferences.AppURL,
		SWVersion:    dev.OsVersion(),
		Manufacturer: dev.Manufacturer(),
		Model:        dev.Model(),
		Identifiers:  []string{prefs.DeviceID},
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

// getHWProductInfo retrieves the model and vendor of the machine. If these
// cannot be retrieved or cannot be found, they will be set to default unknown
// strings.
func getHWProductInfo() (model, vendor string) {
	product, err := ghw.Product(ghw.WithDisableWarnings())
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve hardware information.")
		return unknownModel, unknownVendor
	}
	return product.Name, product.Vendor
}

// Chassis will return the chassis type of the machine, such as "desktop" or
// "laptop". If this cannot be retrieved, it will return "unknown".
func Chassis() string {
	chassisInfo, err := ghw.Chassis(ghw.WithDisableWarnings())
	if err != nil {
		log.Warn().Err(err).Msg("Could not determine chassis type.")
		return "unknown"
	}
	return chassisInfo.Type
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

// GetDistroID will retrieve the distribution ID and version ID. These are
// suitable for usage as part of identifiers and variables. See also
// GetDistroDetails.
func GetDistroID() (id, versionid string) {
	var distroName, distroVersion string
	osReleaseInfo, err := whichdistro.GetOSRelease()
	if err != nil {
		log.Warn().Err(err).Msg("Could not read /etc/os-release. Contact your distro vendor to implement this file.")
		return unknownDistro, unknownDistroVersion
	}
	if v, ok := osReleaseInfo.GetValue("ID"); !ok {
		distroName = unknownDistro
	} else {
		distroName = v
	}
	if v, ok := osReleaseInfo.GetValue("VERSION_ID"); !ok {
		distroVersion = unknownDistroVersion
	} else {
		distroVersion = v
	}
	return distroName, distroVersion
}

// GetDistroDetails will retrieve the distribution name and version. The values
// are pretty-printed and may not be suitable for usage as identifiers and
// variables. See also GetDistroID.
func GetDistroDetails() (name, version string) {
	var distroName, distroVersion string
	osReleaseInfo, err := whichdistro.GetOSRelease()
	if err != nil {
		log.Warn().Err(err).Msg("Could not read /etc/os-release. Contact your distro vendor to implement this file.")
		return unknownDistro, unknownDistroVersion
	}
	if v, ok := osReleaseInfo.GetValue("NAME"); !ok {
		distroName = unknownDistro
	} else {
		distroName = v
	}
	if v, ok := osReleaseInfo.GetValue("VERSION"); !ok {
		distroVersion = unknownDistroVersion
	} else {
		distroVersion = v
	}
	return distroName, distroVersion
}

// GetKernelVersion will retrieve the kernel version.
//
//nolint:prealloc
func GetKernelVersion() string {
	var utsname syscall.Utsname
	var versionBytes []byte

	err := syscall.Uname(&utsname)
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve kernel version.")
		return "Unknown"
	}
	for _, v := range utsname.Release {
		if v == 0 {
			continue
		}
		versionBytes = append(versionBytes, uint8(v))
	}
	return string(versionBytes)
}
