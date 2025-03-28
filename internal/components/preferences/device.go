// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"errors"
	"fmt"
)

const (
	devicePrefPrefix = "device"
	prefDeviceID     = devicePrefPrefix + ".id"
	prefDeviceName   = devicePrefPrefix + ".name"
)

// Device contains the device-specific preferences.
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

var ErrSetDevicePreference = errors.New("could not set device preference")

// SetDeviceID will set the device ID.
func SetDeviceID(id string) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefDeviceID, id); err != nil {
			return errors.Join(ErrSetDevicePreference, err)
		}

		return nil
	}
}

// SetDeviceName will set the device name.
func SetDeviceName(name string) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefDeviceName, name); err != nil {
			return fmt.Errorf("%w: %w", ErrSetDevicePreference, err)
		}

		return nil
	}
}
