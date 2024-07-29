// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
package preferences

import (
	"fmt"

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
		return nil, fmt.Errorf("unable to create device: %w", err)
	}

	dev.Name = name

	// Generate a new unique Device ID
	id, err := device.NewDeviceID()
	if err != nil {
		return nil, fmt.Errorf("unable to create device: %w", err)
	}

	dev.ID = id

	// Retrieve the OS name and version.
	osName, osVersion, err := device.GetOSID()
	if err != nil {
		return nil, fmt.Errorf("unable to create device: %w", err)
	}

	dev.OsName = osName
	dev.OsVersion = osVersion

	// Retrieve the hardware model and manufacturer.
	model, manufacturer, err := device.GetHWProductInfo()
	if err != nil {
		return nil, fmt.Errorf("unable to create device: %w", err)
	}

	dev.Model = model
	dev.Manufacturer = manufacturer

	return dev, nil
}
