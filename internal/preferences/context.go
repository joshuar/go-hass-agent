// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package preferences

import (
	"context"
	"path/filepath"

	"github.com/adrg/xdg"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
)

type contextKey string

const (
	registrationCtxKey contextKey = "registration"
	headlessCtxKey     contextKey = "headless"
	mqttDeviceCtxKey   contextKey = "mqttDevice"
	mqttPrefsCtxKey    contextKey = "mqttPrefs"
	pathCtxKey         contextKey = "path"
)

// RegistrationToCtx stores the registration details passed on the
// command-line to the context.
func RegistrationToCtx(ctx context.Context, registration Registration) context.Context {
	newCtx := context.WithValue(ctx, registrationCtxKey, registration)
	return newCtx
}

// RegistrationFromCtx retrieves the registration details passed on the
// command-line from the context.
func RegistrationFromCtx(ctx context.Context) *Registration {
	registration, ok := ctx.Value(registrationCtxKey).(Registration)
	if !ok {
		return nil
	}

	return &registration
}

// MQTTDeviceToCtx stores the MQTT device in the context.
func MQTTDeviceToCtx(ctx context.Context) context.Context {
	newCtx := context.WithValue(ctx, mqttDeviceCtxKey, getMQTTDevice())
	return newCtx
}

// MQTTDeviceFromCtx retrieves the MQTT device from the context.
func MQTTDeviceFromFromCtx(ctx context.Context) *mqtthass.Device {
	device, ok := ctx.Value(mqttDeviceCtxKey).(*mqtthass.Device)
	if !ok {
		return nil
	}

	return device
}

// MQTTPrefsToCtx stores the MQTT preferences in the context.
func MQTTPrefsToCtx(ctx context.Context) context.Context {
	newCtx := context.WithValue(ctx, mqttPrefsCtxKey, getMQTTPreferences())
	return newCtx
}

// MQTTPrefsFromCtx retrieves the MQTT preferences from the context.
func MQTTPrefsFromFromCtx(ctx context.Context) *MQTT {
	prefs, ok := ctx.Value(mqttPrefsCtxKey).(*MQTT)
	if !ok {
		return nil
	}

	return prefs
}

func PathToCtx(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, pathCtxKey, path)
}

func PathFromCtx(ctx context.Context) string {
	path, ok := ctx.Value(pathCtxKey).(string)
	if !ok {
		return filepath.Join(xdg.ConfigHome, DefaultAppID)
	}

	return path
}
