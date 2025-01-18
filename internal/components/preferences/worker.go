// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go run github.com/matryer/moq -out worker_mocks_test.go . Worker
package preferences

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/pelletier/go-toml/v2"

	"github.com/joshuar/go-hass-agent/internal/components/validation"
)

const (
	workerPrefsPrefix = "worker"
)

// CommonWorkerPrefs contains worker preferences that all workers can/should
// implement. For e.g., a toggle to completely disable the worker.
type CommonWorkerPrefs struct {
	Disabled bool `toml:"disabled" comment:"Set this to true to disable all these sensors."`
}

// Worker represents a Worker from the point of the preferences package. A
// worker has a set of default preferences returned by the DefaultPreferences
// method and an ID that uniquely identifies the worker (and its preferences
// on disk).
type Worker[T any] interface {
	PreferencesID() string
	DefaultPreferences() T
}

var (
	ErrSaveWorkerPrefs = errors.New("error saving worker preferences")
	ErrLoadWorkerPrefs = errors.New("error loading worker preferences")
)

// LoadWorker reads the given worker's preferences from file.
func LoadWorker[T any](ctx context.Context, worker Worker[T]) (*T, error) {
	var (
		prefs        T
		defaultPrefs T
	)
	// Get the default preferences.
	defaultPrefs = worker.DefaultPreferences()
	// Set the key to the worker preferences in the preferences store.
	prefsKey := workerPrefsPrefix + "." + worker.PreferencesID()
	// Try to retrieve any existing preferences. Use those if possible.
	foundPrefs := prefsSrc.Get(prefsKey)
	if foundPrefs != nil {
		// Marshall the map[string]interface returned to []byte.
		data, err := toml.Marshal(foundPrefs)
		if err != nil {
			prefs = defaultPrefs
		} else {
			// Unmarshal the []byte back to the proper preferences type T.
			if err := toml.Unmarshal(data, &prefs); err != nil {
				prefs = defaultPrefs
			}
		}
	} else {
		prefs = defaultPrefs
		// Save the default preferences to the preferences source.
		if err := SaveWorker(ctx, worker, prefs); err != nil {
			return &prefs, fmt.Errorf("%w: %w", ErrLoadWorkerPrefs, err)
		}
	}

	// Validate the preferences, warn and use defaults if invalid.
	if err := validation.Validate.Struct(prefs); err != nil {
		slog.Warn("Worker preferences are invalid, reverting to defaults.",
			slog.String("worker", worker.PreferencesID()),
			slog.String("problems", validation.ParseValidationErrors(err)))
		// Save the default preferences to the preferences source.
		if err := SaveWorker(ctx, worker, defaultPrefs); err != nil {
			return &prefs, fmt.Errorf("%w: %w", ErrLoadWorkerPrefs, err)
		}

		return &defaultPrefs, nil
	}

	// Return preferences.
	return &prefs, nil
}

// SaveWorker saves the given worker's preferences to file.
func SaveWorker[T any](ctx context.Context, worker Worker[T], prefs T) error {
	// We can't define the structure for every possible worker beforehand, so
	// use map[string]any as the structure for saving.
	prefsMaps := make(map[string]any)
	// Marshal the worker's prefs object into bytes, using the toml tag
	// structure.
	data, err := toml.Marshal(&prefs)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSaveWorkerPrefs, err)
	}
	// Unmarshal back into a map[string]any that we can save into the preferences
	// file.
	if err := toml.Unmarshal(data, &prefsMaps); err != nil {
		return fmt.Errorf("%w: %w", ErrSaveWorkerPrefs, err)
	}
	// Merge the worker preferences into the preferences file.
	if err := prefsSrc.Set(workerPrefsPrefix+"."+worker.PreferencesID(), prefsMaps); err != nil {
		return fmt.Errorf("%w: %w", ErrSaveWorkerPrefs, err)
	}
	// Save the preferences.
	return Save(ctx)
}
