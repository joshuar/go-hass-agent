// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver
package device

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jaypipes/ghw"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	unknownVendor        = "Unknown Vendor"
	unknownModel         = "Unknown Model"
	unknownDistro        = "Unknown Distro"
	unknownDistroVersion = "Unknown Version"
	UnknownValue         = "unknown"
)

var ErrUnsupportedHardware = errors.New("unsupported hardware")

type Device struct {
	ID                 string  `toml:"id" json:"device_id" validate:"required,ascii"`
	AppID              string  `toml:"-" json:"app_id"`
	AppName            string  `toml:"-" json:"app_name"`
	AppVersion         string  `toml:"-" json:"app_version"`
	Name               string  `toml:"name" json:"device_name" validate:"required,ascii"`
	Manufacturer       string  `toml:"-" json:"manufacturer"`
	Model              string  `toml:"-" json:"model"`
	OsName             string  `toml:"-" json:"os_name"`
	OsVersion          string  `toml:"-" json:"os_version"`
	AppData            AppData `toml:"-" json:"app_data,omitempty"`
	SupportsEncryption bool    `toml:"-" json:"supports_encryption"`
}

type AppData struct {
	PushWebsocket bool `toml:"-" json:"push_websocket_channel"`
}

func NewDevice() *Device {
	dev := &Device{
		AppName:    preferences.AppName,
		AppVersion: preferences.AppVersion(),
		AppID:      preferences.AppID(),
	}

	// Retrieve the name as the device name.
	name, err := getHostname(true)
	if err != nil {
		slog.Warn("Unable to determine device hostname.", slog.Any("error", err))
	}

	dev.Name = name

	// Generate a new unique Device ID
	id, err := newDeviceID()
	if err != nil {
		slog.Warn("Unable to generate a device ID.", slog.Any("error", err))
	}

	dev.ID = id

	// Retrieve the OS name and version.
	osName, osVersion, err := getOSID()
	if err != nil {
		slog.Warn("Unable to determine OS details.", slog.Any("error", err))
	}

	dev.OsName = osName
	dev.OsVersion = osVersion

	// Retrieve the hardware model and manufacturer.
	model, manufacturer, err := getHWProductInfo()
	if err != nil {
		slog.Warn("Unable to determine device hardware details.", slog.Any("error", err))
	}

	dev.Model = model
	dev.Manufacturer = manufacturer

	return dev
}

// newDeviceID create a new device ID. It will be a randomly generated UUIDv4.
func newDeviceID() (string, error) {
	deviceID, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("could not retrieve a machine ID: %w", err)
	}

	return deviceID.String(), nil
}

// Chassis will return the chassis type of the machine, such as "desktop" or
// "laptop". If this cannot be retrieved, it will return "unknown".
func Chassis() (string, error) {
	chassisInfo, err := ghw.Chassis(ghw.WithDisableWarnings())
	if err != nil || chassisInfo == nil {
		return "", fmt.Errorf("could not determine chassis type: %w", err)
	}

	return chassisInfo.Type, nil
}

// getHostname retrieves the hostname of the device running the agent, or
// localhost if that doesn't work.
//
//revive:disable:flag-parameter
func getHostname(short bool) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost", fmt.Errorf("could not retrieve hostname: %w", err)
	}

	if short {
		shortHostname, _, _ := strings.Cut(hostname, ".")

		return shortHostname, nil
	}

	return hostname, nil
}

// getHWProductInfo retrieves the model and vendor of the machine. If these
// cannot be retrieved or cannot be found, they will be set to default unknown
// strings.
func getHWProductInfo() (model, vendor string, err error) {
	product, err := ghw.Product(ghw.WithDisableWarnings())
	if err != nil || product == nil {
		return unknownModel, unknownVendor, fmt.Errorf("could not retrieve hardware information: %w", err)
	}

	return product.Name, product.Vendor, nil
}
