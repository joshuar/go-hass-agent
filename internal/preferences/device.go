// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"errors"
	"fmt"
)

type Device struct {
	ID   string `toml:"id" validate:"required,ascii"`
	Name string `toml:"name" validate:"required,ascii"`
}

var ErrSetDevicePreference = errors.New("could not set device preference")

// SetDevicePreferences sets the device preferences.
func SetDevicePreferences(device *Device) error {
	if err := prefsSrc.Set("device.id", device.ID); err != nil {
		return fmt.Errorf("%w: %w", ErrSetDevicePreference, err)
	}

	if err := prefsSrc.Set("device.name", device.Name); err != nil {
		return fmt.Errorf("%w: %w", ErrSetDevicePreference, err)
	}

	return nil
}

// DeviceName retrieves the device name from the preferences.
func DeviceName() string {
	return prefsSrc.String("device.name")
}

// DeviceID retrieves the device ID from the preferences.
func DeviceID() string {
	return prefsSrc.String("device.id")
}
