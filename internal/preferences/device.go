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
	ID   string `toml:"id" validate:"required,ascii"`
	Name string `toml:"name" validate:"required,ascii"`
}

var ErrSetDevicePreference = errors.New("could not set device preference")

// SetDevicePreferences sets the device preferences.
func SetDevicePreferences(device *Device) error {
	if err := prefsSrc.Set(prefDeviceID, device.ID); err != nil {
		return fmt.Errorf("%w: %w", ErrSetDevicePreference, err)
	}

	if err := prefsSrc.Set(prefDeviceName, device.Name); err != nil {
		return fmt.Errorf("%w: %w", ErrSetDevicePreference, err)
	}

	return nil
}

// DeviceID retrieves the device ID from the preferences.
func DeviceID() string {
	return prefsSrc.String(prefDeviceID)
}

// DeviceName retrieves the device name from the preferences.
func DeviceName() string {
	return prefsSrc.String(prefDeviceName)
}
