// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"

	"github.com/joshuar/go-hass-agent/internal/device"
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

// NewDevice creates a new device. This is used during registration with Home
// Assistant to identify the host running Go Hass Agent. While most of the
// information generated is only needed during registration, the device ID and
// Name will be stored in the preferences for later reference.
func NewDevice() *Device {
	dev := &Device{
		AppName:    AppName,
		AppVersion: AppVersion(),
		AppID:      appID,
	}

	// Retrieve the name as the device name.
	name, err := device.GetHostname()
	if err != nil {
		slog.Warn("Unable to determine device hostname.",
			slog.Any("error", err))
	}

	dev.Name = name

	// Generate a new unique Device ID
	id, err := newDeviceID()
	if err != nil {
		slog.Warn("Unable to generate a device ID.",
			slog.Any("error", err))
	}

	dev.ID = id

	// Retrieve the OS name and version.
	osName, osVersion, err := device.GetOSID()
	if err != nil {
		slog.Warn("Unable to determine OS details.",
			slog.Any("error", err))
	}

	dev.OsName = osName
	dev.OsVersion = osVersion

	// Retrieve the hardware model and manufacturer.
	model, manufacturer, err := device.GetHWProductInfo()
	if err != nil {
		slog.Warn("Unable to determine device hardware details.",
			slog.Any("error", err))
	}

	dev.Model = model
	dev.Manufacturer = manufacturer

	// Set the device id and name in the preferences store.
	SetPreferences(
		SetDeviceID(dev.ID),
		SetDeviceName(dev.Name),
	)

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
