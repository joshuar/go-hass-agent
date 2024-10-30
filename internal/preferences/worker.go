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

// Worker represents a Worker from the point of the preferences package. A
// worker has a set of default preferences returned by the DefaultPreferences
// method and an ID that uniquely identifies the worker (and its preferences
// file on disk).
type Worker[T any] interface {
	PreferencesID() string
	DefaultPreferences() T
}

// SaveWorkerPreferences will save the given preferences for the given app to file. If the
// preferences cannot be saved, a non-nil error will be returned.
func SaveWorkerPreferences[T any](ctx context.Context, worker string, prefs T) error {
	data, err := toml.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("could not marshal app preferences: %w", err)
	}

	if err := os.WriteFile(workerPreferencesFile(ctx, worker), data, defaultFilePerms); err != nil {
		return fmt.Errorf("could not write app preferences: %w", err)
	}

	return nil
}

// LoadWorkerPreferences will load the given worker preferences from file. If
// the preferences cannot be loaded, a non-nil error will be returned. If the
// preferences file doesn't exist, the default worker preferences will be
// returned.
func LoadWorkerPreferences[T any](ctx context.Context, worker Worker[T]) (T, error) {
	// Load config from file. If the preferences cannot be loaded for any reason
	// other than the preferences file does not exist , return an error.
	data, err := os.ReadFile(workerPreferencesFile(ctx, worker.PreferencesID()))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return *new(T), fmt.Errorf("could not read app preferences file: %w", err)
	}

	// If the preferences file does not exists, return the default preferences
	// for the worker.
	if errors.Is(err, os.ErrNotExist) {
		// Save the newly created preferences to disk.
		if err := SaveWorkerPreferences(ctx, worker.PreferencesID(), worker.DefaultPreferences()); err != nil {
			return *new(T), fmt.Errorf("could not save default preferences: %w", err)
		}

		return worker.DefaultPreferences(), nil
	}

	// Otherwise, we have existing preferences. Unmarshal and return the
	// preferences if possible.
	var prefs T

	if err := toml.Unmarshal(data, &prefs); err != nil {
		return *new(T), fmt.Errorf("could not unmarshal app preferences: %w", err)
	}

	return prefs, nil
}

func workerPreferencesFile(ctx context.Context, worker string) string {
	appID := AppIDFromContext(ctx)
	return filepath.Join(xdg.ConfigHome, appID, worker+"_"+preferencesFile)
}
