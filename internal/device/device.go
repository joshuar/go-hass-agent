// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"github.com/joshuar/go-hass-agent/internal/hass"
)

type DeviceInfo interface {
	DeviceID() string
	AppID() string
	AppName() string
	AppVersion() string
	DeviceName() string
	Manufacturer() string
	Model() string
	OsName() string
	OsVersion() string
	SupportsEncryption() bool
	AppData() interface{}
}

func GenerateRegistrationRequest(d DeviceInfo) *hass.RegistrationRequest {
	if d.AppData() != nil {
		return &hass.RegistrationRequest{
			DeviceID:           d.DeviceID(),
			AppID:              d.AppID(),
			AppName:            d.AppName(),
			AppVersion:         d.AppVersion(),
			DeviceName:         d.DeviceName(),
			Manufacturer:       d.Manufacturer(),
			Model:              d.Model(),
			OsName:             d.OsName(),
			OsVersion:          d.OsVersion(),
			SupportsEncryption: d.SupportsEncryption(),
			AppData:            d.AppData(),
		}
	} else {
		return &hass.RegistrationRequest{
			DeviceID:           d.DeviceID(),
			AppID:              d.AppID(),
			AppName:            d.AppName(),
			AppVersion:         d.AppVersion(),
			DeviceName:         d.DeviceName(),
			Manufacturer:       d.Manufacturer(),
			Model:              d.Model(),
			OsName:             d.OsName(),
			OsVersion:          d.OsVersion(),
			SupportsEncryption: d.SupportsEncryption(),
		}
	}
}
