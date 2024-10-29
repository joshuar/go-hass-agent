// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package preferences

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

type Worker interface {
	ID() string
	DefaultPreferences() WorkerPreferences
}

func workerPreferencesFile(ctx context.Context, worker string) string {
	appID := AppIDFromContext(ctx)
	return filepath.Join(xdg.ConfigHome, appID, worker+"_"+preferencesFile)
}

// Preference represents a single preference in a preferences file.
type Preference struct {
	// Value is the actual preference value.
	Value any `toml:"value"`
	// Description is a string that describes the preference, and may be used
	// for display purposes.
	Description string `toml:"description,omitempty"`
	// Secret is a flag that indicates whether this preference represents a
	// secret. The value has no effect on the preference encoding in the TOML,
	// only on how to display the preference to the user (masked or plaintext).
	Secret bool `toml:"-"`
}

// AppPreferences is a structure that can be used to represent app preferences.
// As app preferences vary between apps, a map of Preference values is used.
type WorkerPreferences map[string]*Preference

// GetValue returns the value of the preference with the given key name and a
// bool to indicate whether it was found or not. If the preference was not
// found, the value will be nil.
func (p WorkerPreferences) GetValue(key string) (value any, found bool) {
	value, found = p[key]
	if !found {
		return nil, false
	}

	return value, true
}

// SetValue sets the preference with the given key name to the given value. It
// currently returns a nil error but may in the future return a non-nil error if
// the preference could not be set.
func (p WorkerPreferences) SetValue(key string, value any) error {
	p[key].Value = value

	return nil
}

// GetDescription returns the description of the preference with the given key
// name.
func (p WorkerPreferences) GetDescription(key string) string {
	return p[key].Description
}

// IsSecret returns a boolean to indicate whether the preference with the given
// key name should be masked or obscured when displaying to the user.
func (p WorkerPreferences) IsSecret(key string) bool {
	return p[key].Secret
}

// Keys returns all key names for all known preferences.
func (p WorkerPreferences) Keys() []string {
	keys := make([]string, 0, len(p))
	for key := range p {
		keys = append(keys, key)
	}

	return keys
}

// SaveApp will save the given preferences for the given app to file. If the
// preferences cannot be saved, a non-nil error will be returned.
//
//nolint:mnd
func SaveWorkerPreferences(ctx context.Context, worker string, prefs WorkerPreferences) error {
	data, err := toml.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("could not marshal app preferences: %w", err)
	}

	if err := os.WriteFile(workerPreferencesFile(ctx, worker), data, 0o600); err != nil {
		return fmt.Errorf("could not write app preferences: %w", err)
	}

	return nil
}

// LoadApp will load the given app preferences from file. If the preferences
// cannot be loaded, a non-nil error will be returned. Apps should take special
// care to handle os.ErrNotExists verses other returned errors. In the former case, it
// would be wise to treat as not an error and revert to using default
// preferences.
func LoadWorkerPreferences(ctx context.Context, worker Worker) (WorkerPreferences, error) {
	// Load config from file. If the preferences cannot be loaded for any reason
	// other than the preferences file does not exist , return an error.
	data, err := os.ReadFile(workerPreferencesFile(ctx, worker.ID()))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("could not read app preferences file: %w", err)
	}

	// If the preferences file does not exists, return the default preferences
	// for the app.
	if errors.Is(err, os.ErrNotExist) {
		// Save the newly created preferences to disk.
		if err := SaveWorkerPreferences(ctx, worker.ID(), worker.DefaultPreferences()); err != nil {
			return nil, fmt.Errorf("could not save default preferences: %w", err)
		}

		return worker.DefaultPreferences(), nil
	}

	// Otherwise, we have existing preferences. Unmarshal and return the
	// preferences if possible.
	prefs := make(WorkerPreferences)

	if err := toml.Unmarshal(data, &prefs); err != nil {
		return nil, fmt.Errorf("could not unmarshal app preferences: %w", err)
	}

	return prefs, nil
}
