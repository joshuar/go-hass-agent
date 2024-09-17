// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
package preferences

import (
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/device"
)

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

func newDevice() (*Device, error) {
	dev := &Device{
		AppName:    AppName,
		AppVersion: AppVersion,
		AppID:      AppID,
	}

	// Retrieve the name as the device name.
	name, err := device.GetHostname(true)
	if err != nil {
		slog.Warn("Unable to determine device hostname.", slog.Any("error", err))
	}

	dev.Name = name

	// Generate a new unique Device ID
	id, err := device.NewDeviceID()
	if err != nil {
		slog.Warn("Unable to generate a device ID.", slog.Any("error", err))
	}

	dev.ID = id

	// Retrieve the OS name and version.
	osName, osVersion, err := device.GetOSID()
	if err != nil {
		slog.Warn("Unable to determine OS details.", slog.Any("error", err))
	}

	dev.OsName = osName
	dev.OsVersion = osVersion

	// Retrieve the hardware model and manufacturer.
	model, manufacturer, err := device.GetHWProductInfo()
	if err != nil {
		slog.Warn("Unable to determine device hardware details.", slog.Any("error", err))
	}

	dev.Model = model
	dev.Manufacturer = manufacturer

	return dev, nil
}

func (p *Preferences) GetDeviceInfo() *Device {
	return p.Device
}

func (p *Preferences) DeviceName() string {
	if p.Device != nil {
		return p.Device.Name
	}

	return unknownValue
}

func (p *Preferences) DeviceID() string {
	if p.Device != nil {
		return p.Device.ID
	}

	return unknownValue
}
