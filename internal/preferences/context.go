// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	"context"
)

type contextKey string

const prefsContextKey contextKey = "preferences"

// ContextSetPrefs will store the preferences in the given context.
func ContextSetPrefs(ctx context.Context, p *Preferences) context.Context {
	return context.WithValue(ctx, prefsContextKey, p)
}

// ContextGetPrefs will attempt to fetch the preferences from the given context.
func ContextGetPrefs(ctx context.Context) (*Preferences, error) {
	prefs, ok := ctx.Value(prefsContextKey).(*Preferences)
	if !ok {
		return nil, ErrNoPreferences
	}

	return prefs, nil
}

// ContextGetMQTTPrefs will attempt to fetch the MQTT preferences from the given
// context.
func ContextGetMQTTPrefs(ctx context.Context) (*MQTTPreferences, error) {
	prefs, ok := ctx.Value(prefsContextKey).(*Preferences)
	if !ok {
		return nil, ErrNoPreferences
	}

	mqttPrefs := prefs.GetMQTTPreferences()
	if mqttPrefs == nil {
		return nil, ErrNoPreferences
	}

	return mqttPrefs, nil
}
