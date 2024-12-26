// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package device

import (
	"context"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func GenerateMQTTDevice(ctx context.Context) *mqtthass.Device {
	// Retrieve the hardware model and manufacturer.
	model, manufacturer, _ := getHWProductInfo() //nolint:errcheck // error doesn't matter

	return &mqtthass.Device{
		Name:         preferences.DeviceName(),
		URL:          preferences.AppURL,
		SWVersion:    preferences.AppVersion(),
		Manufacturer: manufacturer,
		Model:        model,
		Identifiers:  []string{preferences.AppID(), preferences.DeviceName(), preferences.DeviceID()},
	}
}
