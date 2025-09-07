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

	"github.com/goforj/godump"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var ErrLoadConfig = errors.New("could not load config")

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
	ConfigFile = "config.toml"

	// DefaultServer is the default Home Assistant server address.
	DefaultServer = "http://localhost:8123"
)

type configData struct {
	mu   sync.Mutex
	src  *koanf.Koanf
	path string
}

var globalConfig = configData{
	mu:   sync.Mutex{},
	src:  koanf.New("."),
	path: GetPath(),
}

// Init initializes the config store. This will load the global (app) config
// values and set up a config backend that other components can use via the Load
// method. This only happens once.
var Init = sync.OnceValue(func() error {
	// Create the config directory if it does not exist.
	_, err := os.Stat(globalConfig.path)
	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(globalConfig.path, os.ModeDir)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrLoadConfig, err)
		}
	}
	// Load config file
	provider := file.Provider(filepath.Join(globalConfig.path, ConfigFile))
	err = globalConfig.src.Load(provider, toml.Parser())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %w", ErrLoadConfig, err)
	}
	// Watch for changes.
	provider.Watch(func(event any, err error) {
		if err != nil {
			slog.Error("Error occurred while watching config for changes.",
				slog.Any("error", err),
			)
			return
		}
		// Reload config on changes.
		slog.Debug("Config file changed, reloading config.")
		globalConfig.src = koanf.New(".")
		globalConfig.src.Load(provider, toml.Parser())
	})

	slog.Debug("Config backend initialized.")

	return nil
})

func GetPath() string {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		panic("could not determine config directory.")
	}
	return filepath.Join(userConfigDir, AppID)
}

func SetPath(path string) {
	globalConfig.path = path
}

// Load will load the config for a component, using the given file and
// environment prefixes, and marshaling the config into the given config object.
// Components should take care to ensure this is called only once, where
// required.
func Load(configPrefix string, cfg any) error {
	// Unmarshal config, overwriting defaults.
	if err := globalConfig.src.UnmarshalWithConf(configPrefix, cfg, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return fmt.Errorf("%w: %w", ErrLoadConfig, err)
	}

	slog.Debug("Loading config for component.",
		slog.String("component", configPrefix))

	return nil
}

// Set will set the given options in the config. After all options are set, the config file is written.
func Set(options map[string]any) error {
	// globalConfig.mu.Lock()
	// defer globalConfig.mu.Unlock()
	for key, value := range options {
		err := globalConfig.src.Set(key, value)
		if err != nil {
			slog.Error("Unable to set config option.",
				slog.String("key", key),
				slog.Any("value", value),
				slog.Any("error", err),
			)
		}
	}
	err := save()
	if err != nil {
		return fmt.Errorf("unable to save config: %w", err)
	}
	return nil
}

// save will save the new values of the specified preferences to the existing
// preferences file.
func save() error {
	// globalConfig.mu.Lock()
	// defer globalConfig.mu.Unlock()

	configFile := filepath.Join(globalConfig.path, ConfigFile)
	godump.Dump(configFile)

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
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0o750)
		if err != nil {
			return fmt.Errorf("unable to create new directory: %w", err)
		}
	}

	return nil
}
