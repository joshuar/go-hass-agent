// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver

package device

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/iancoleman/strcase"
	"github.com/jaypipes/ghw"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	unknownVendor        = "Unknown Vendor"
	unknownModel         = "Unknown Model"
	unknownDistro        = "Unknown Distro"
	unknownDistroVersion = "Unknown Version"
)

var ErrUnsupportedHardware = errors.New("unsupported hardware")

type Device struct {
	appName    string
	appVersion string
	hostname   string
	deviceID   string
	hwVendor   string
	hwModel    string
	osName     string
	osVersion  string
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
	return l.osName
}

func (l *Device) OsVersion() string {
	return l.osVersion
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

//nolint:exhaustruct
func New(name, version string) *Device {
	dev := &Device{
		appName:    name,
		appVersion: version,
		deviceID:   getDeviceID(),
		hostname:   getHostname(),
	}
	dev.osName, dev.osVersion = getOSID()
	dev.hwModel, dev.hwVendor = getHWProductInfo()

	return dev
}

//nolint:exhaustruct
func MQTTDeviceInfo(ctx context.Context) *mqtthass.Device {
	prefs, err := preferences.ContextGetPrefs(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve preferences.")
	}

	hostname, _, _ := strings.Cut(getHostname(), ".")
	_, version := getOSID()
	model, manufacturer := getHWProductInfo()

	return &mqtthass.Device{
		Name:         hostname,
		URL:          preferences.AppURL,
		SWVersion:    version,
		Manufacturer: model,
		Model:        manufacturer,
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
