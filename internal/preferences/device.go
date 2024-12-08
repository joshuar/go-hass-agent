// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

type Device struct {
	ID   string `toml:"id" validate:"required,ascii"`
	Name string `toml:"name" validate:"required,ascii"`
}

// SetDevicePreferences sets the device preferences.
func SetDevicePreferences(device *Device) error {
	prefsSrc.Set("device.id", device.ID)
	prefsSrc.Set("device.name", device.Name)

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

// func (p *Preferences) GetDeviceInfo() *Device {
// 	return p.Device
// }

// func (p *Preferences) DeviceName() string {
// 	if p.Device != nil {
// 		return p.Device.Name
// 	}

// 	return unknownValue
// }

// func (p *Preferences) DeviceID() string {
// 	if p.Device != nil {
// 		return p.Device.ID
// 	}

// 	return unknownValue
// }
