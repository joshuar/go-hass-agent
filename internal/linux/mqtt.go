// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	mqtthass "github.com/joshuar/go-hass-anything/v7/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func NewButton(entityID string) *mqtthass.EntityConfig {
	return mqtthass.NewEntityByID(entityID, preferences.AppName, preferences.MQTTTopicPrefix).
		AsButton().
		WithOriginInfo(preferences.MQTTOrigin()).
		WithDeviceInfo(mqttDevice())
}

func NewSlider(entityID string, step, min, max float64) *mqtthass.EntityConfig {
	return mqtthass.NewEntityByID(entityID, preferences.AppName, preferences.MQTTTopicPrefix).
		AsNumber(step, min, max, mqtthass.NumberSlider).
		WithOriginInfo(preferences.MQTTOrigin()).
		WithDeviceInfo(mqttDevice())
}

func mqttDevice() *mqtthass.Device {
	dev := NewDevice(preferences.AppName, preferences.AppVersion)
	return &mqtthass.Device{
		Name:         dev.DeviceName(),
		URL:          preferences.AppURL,
		SWVersion:    dev.OsVersion(),
		Manufacturer: dev.Manufacturer(),
		Model:        dev.Model(),
		Identifiers:  []string{dev.DeviceID()},
	}
}
