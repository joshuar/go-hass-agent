// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package agent

import (
	"context"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type contextKey string

const (
	registrationCtxKey contextKey = "registration"
	headlessCtxKey     contextKey = "headless"
	mqttDeviceCtxKey   contextKey = "mqttDevice"
	mqttPrefsCtxKey    contextKey = "mqttPrefs"
)

// RegistrationToCtx stores the registration details passed on the
// command-line to the context.
func RegistrationToCtx(ctx context.Context, registration preferences.Registration) context.Context {
	newCtx := context.WithValue(ctx, registrationCtxKey, registration)
	return newCtx
}

// RegistrationFromCtx retrieves the registration details passed on the
// command-line from the context.
func RegistrationFromCtx(ctx context.Context) *preferences.Registration {
	registration, ok := ctx.Value(registrationCtxKey).(preferences.Registration)
	if !ok {
		return nil
	}

	return &registration
}

// HeadlessToCtx stores the value of the headless command-line option in the context.
func HeadlessToCtx(ctx context.Context, headless bool) context.Context {
	newCtx := context.WithValue(ctx, headlessCtxKey, headless)
	return newCtx
}

// HeadlessFromCtx retrieves the value of the headless command-line option from
// the context.
func HeadlessFromCtx(ctx context.Context) bool {
	headless, ok := ctx.Value(headlessCtxKey).(bool)
	if !ok {
		return false
	}

	return headless
}

// MQTTDeviceToCtx stores the MQTT device in the context.
func MQTTDeviceToCtx(ctx context.Context, device *mqtthass.Device) context.Context {
	newCtx := context.WithValue(ctx, mqttDeviceCtxKey, device)
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
func MQTTPrefsToCtx(ctx context.Context, prefs mqttapi.Preferences) context.Context {
	newCtx := context.WithValue(ctx, mqttPrefsCtxKey, prefs)
	return newCtx
}

// MQTTPrefsFromCtx retrieves the MQTT preferences from the context.
func MQTTPrefsFromFromCtx(ctx context.Context) mqttapi.Preferences {
	prefs, ok := ctx.Value(mqttPrefsCtxKey).(mqttapi.Preferences)
	if !ok {
		return nil
	}

	return prefs
}
