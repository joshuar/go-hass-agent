// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	ErrGetConfig  = errors.New("could not get from config")
	ErrLoadConfig = errors.New("could not load config")
)

var AppVersion = "_UNKNOWN_"

const (
	// AppName is the formatted application name.
	AppName = "Go Hass Agent"
	// AppID is the ID of the application.
	AppID = "go-hass-agent"
	// AppURL is the canonical URL for the application.
	AppURL = "https://github.com/joshuar/go-hass-agent"
	// AppDescription is the formatted summary of the application.
	AppDescription = "A Home Assistant, native app for desktop/laptop devices."
	// ConfigFile is the location of the server configuration file.
	ConfigFile = "preferences.toml"

	// DefaultServer is the default Home Assistant server address.
	DefaultServer = "http://localhost:8123"
)

type configData struct {
	mu sync.Mutex

	src  *koanf.Koanf
	path string
}

var globalConfig = configData{
	src:  koanf.New("."),
	path: GetPath(),
}

var customPath string

// Init initializes the config store. This will load the global (app) config
// values and set up a config backend that other components can use via the Load
// method. This only happens once.
var Init = sync.OnceValue(func() error {
	// Create the config directory if it does not exist.
	if err := checkPath(globalConfig.path); err != nil {
		return fmt.Errorf("%w: %w", ErrLoadConfig, err)
	}

	// Load config file
	provider := file.Provider(filepath.Join(globalConfig.path, ConfigFile))
	if err := globalConfig.src.Load(provider, toml.Parser()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %w", ErrLoadConfig, err)
	}
	// Watch for changes.
	// provider.Watch(func(event any, err error) {
	// 	if err != nil {
	// 		slog.Error("Error occurred while watching config for changes.",
	// 			slog.Any("error", err),
	// 		)
	// 		return
	// 	}
	// 	// Reload config on changes.
	// 	slog.Debug("Config file changed, reloading config.")
	// 	globalConfig.src = koanf.New(".")
	// 	globalConfig.src.Load(provider, toml.Parser())
	// })

	slog.Debug("Config backend initialized.")

	return nil
})

// GetPath returns the directory path under which the config file (and other files/data) is stored.
func GetPath() string {
	if customPath != "" {
		return customPath
	}
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		panic("could not determine config directory.")
	}
	return filepath.Join(userConfigDir, AppID)
}

// SetPath sets the directory path under which the config file (and other files/data) will be stored.
func SetPath(path string) {
	customPath = path
}

// Load will load the config for a component, using the given file and
// environment prefixes, and marshaling the config into the given config object.
// Components should take care to ensure this is called only once, where
// required.
func Load(path string, cfg any) error {
	// Unmarshal config, overwriting defaults.
	if err := globalConfig.src.UnmarshalWithConf(path, cfg, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return fmt.Errorf("could not load config %s: %w", path, err)
	}
	return nil
}

// Save will save the given config at the given path.
func Save(path string, config any) error {
	globalConfig.mu.Lock()
	err := globalConfig.src.Set(path, config)
	if err != nil {
		return fmt.Errorf("unable to save config: %w", err)
	}
	globalConfig.mu.Unlock()
	err = save()
	if err != nil {
		return fmt.Errorf("unable to save config: %w", err)
	}
	return nil
}

// Set will set the given options in the config. After all options are set, the config file is written.
func Set(options map[string]any) error {
	globalConfig.mu.Lock()
	for key, value := range options {
		if err := globalConfig.src.Set(key, value); err != nil {
			slog.Error("Unable to set config option.",
				slog.String("key", key),
				slog.Any("value", value),
				slog.Any("error", err),
			)
		}
	}
	globalConfig.mu.Unlock()
	if err := save(); err != nil {
		return fmt.Errorf("unable to save config: %w", err)
	}
	return nil
}

// Get will return the value located at the given path in the config.
func Get[T any](path string) (T, error) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	value, ok := globalConfig.src.Get(path).(T)
	if !ok {
		return value, fmt.Errorf("%w: %s: not %T", ErrGetConfig, path, value)
	}
	return value, nil
}

// Exists reports whether the given path exists in the config.
func Exists(path string) bool {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	return globalConfig.src.Exists(path)
}

// save will save the new values of the specified preferences to the existing
// preferences file.
func save() error {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()

	configFile := filepath.Join(globalConfig.path, ConfigFile)

	if err := checkPath(globalConfig.path); err != nil {
		return err
	}

	b, err := globalConfig.src.Marshal(toml.Parser())
	if err != nil {
		return fmt.Errorf("unable to marshal config: %w", err)
	}

	err = os.WriteFile(configFile, b, 0o600)
	if err != nil {
		return fmt.Errorf("unable to write config file %s: %w", configFile, err)
	}

	slog.Debug("Saved config to disk.",
		slog.String("file", configFile),
	)

	return nil
}

// checkPath checks that the given directory exists. If it doesn't it will be
// created.
func checkPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0750); err != nil {
			return fmt.Errorf("unable to create new directory: %w", err)
		}
	}
	return nil
}
