// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package preferences

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/joshuar/go-hass-agent/internal/validation"
)

const (
	prefsEnvPrefix      = "GOHASSAGENT_"
	AppName             = "Go Hass Agent"
	AppURL              = "https://github.com/joshuar/go-hass-agent"
	FeatureRequestURL   = AppURL + "/issues/new?assignees=joshuar&labels=&template=feature_request.md&title="
	IssueURL            = AppURL + "/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D"
	AppDescription      = "A Home Assistant, native app for desktop/laptop devices."
	defaultAppID        = "go-hass-agent"
	MQTTTopicPrefix     = "homeassistant"
	LogFile             = "go-hass-agent.log"
	defaultFilePerms    = 0o600
	preferencesFilename = "preferences.toml"

	DefaultServer = "http://localhost:8123"
	DefaultSecret = "ALongSecretString"

	prefRegistered = "registered"
)

var (
	//lint:ignore U1000 some of these will be used in the future
	gitVersion, gitCommit, gitTreeState, buildDate string
	AppVersion                                     = gitVersion
)

// Consistent error messages.
var (
	ErrLoadPreferences     = errors.New("error loading preferences")
	ErrSavePreferences     = errors.New("error saving preferences")
	ErrValidatePreferences = errors.New("error validating preferences")
	ErrSetPreference       = errors.New("error setting preference")
)

// Default agent preferences.
var defaultAgentPreferences = &preferences{
	Version:    AppVersion,
	Registered: false,
	MQTT: &MQTT{
		MQTTEnabled: false,
	},
	Registration: &Registration{
		Server: DefaultServer,
		Token:  DefaultSecret,
	},
	Hass: &Hass{
		RestAPIURL:   DefaultServer,
		WebsocketURL: DefaultServer,
		WebhookID:    DefaultSecret,
	},
	// WorkerPrefs: make(map[string]map[string]any),
}

var (
	prefsSrc        = koanf.New(".")
	preferencesFile = filepath.Join(xdg.ConfigHome, defaultAppID, preferencesFilename)
	mu              = sync.Mutex{}
)

// preferences defines all preferences for Go Hass Agent.
type preferences struct {
	MQTT         *MQTT         `toml:"mqtt,omitempty"`
	Registration *Registration `toml:"registration"`
	Hass         *Hass         `toml:"hass"`
	Device       *Device       `toml:"device"`
	Version      string        `toml:"version" validate:"required"`
	Registered   bool          `toml:"registered" validate:"boolean"`
}

// Load will retrieve the current preferences from the preference file on disk.
// If there is a problem during retrieval, an error will be returned.
var Load = func(ctx context.Context) error {
	return sync.OnceValue(func() error {
		appID := AppIDFromContext(ctx)
		preferencesFile = filepath.Join(xdg.ConfigHome, appID, preferencesFilename)

		slog.Debug("Loading preferences.", slog.String("file", preferencesFile))

		// Load config file
		if err := prefsSrc.Load(file.Provider(preferencesFile), toml.Parser()); err != nil {
			return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
		}
		// Merge config with any environment variables.
		if err := prefsSrc.Load(env.Provider(prefsEnvPrefix, ".", func(s string) string {
			return strings.Replace(strings.ToLower(
				strings.TrimPrefix(s, prefsEnvPrefix)), "_", ".", -1)
		}), nil); err != nil {
			return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
		}
		// Unmarshal config, overwriting defaults.
		if err := prefsSrc.UnmarshalWithConf("", defaultAgentPreferences, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
			return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
		}

		return nil
	})()
}

// Reset will remove the preferences file.
func Reset(ctx context.Context) error {
	appID := AppIDFromContext(ctx)
	preferencesFile = filepath.Join(xdg.ConfigHome, appID, preferencesFilename)

	slog.Debug("Removing preferences.", slog.String("file", preferencesFile))

	_, err := os.Stat(preferencesFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
	}

	err = os.Remove(preferencesFile)
	if err != nil {
		return fmt.Errorf("unable to reset preferences: %w", err)
	}

	return nil
}

// Validate ensures the configuration is valid.
func Validate() error {
	currentPreferences := &preferences{}

	// Unmarshal current preferences.
	if err := prefsSrc.UnmarshalWithConf("", currentPreferences, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
	}

	// Validate current preferences.
	err := validation.Validate.Struct(currentPreferences)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrValidatePreferences, validation.ParseValidationErrors(err))
	}

	return nil
}

// Save will save the new values of the specified preferences to the existing
// preferences file. NOTE: if the preferences file does not exist, Save will
// return an error. Use New if saving preferences for the first time.
func Save(ctx context.Context) error {
	mu.Lock()
	defer mu.Unlock()

	appID := AppIDFromContext(ctx)
	preferencesFile = filepath.Join(xdg.ConfigHome, appID, preferencesFilename)

	slog.Debug("Saving preferences.", slog.String("file", preferencesFile))

	if err := Validate(); err != nil {
		return err
	}

	if err := checkPath(filepath.Dir(preferencesFile)); err != nil {
		return err
	}

	b, err := prefsSrc.Marshal(toml.Parser())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSavePreferences, err)
	}

	err = os.WriteFile(preferencesFile, b, defaultFilePerms)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSavePreferences, err)
	}

	return nil
}

// SetRegistered sets whether Go Hass Agent has been registered with Home
// Assistant.
func SetRegistered(value bool) error {
	if err := prefsSrc.Set(prefRegistered, value); err != nil {
		return fmt.Errorf("%w: %w", ErrSetPreference, err)
	}

	return nil
}

// Registered returns the registration status of Go Hass Agent.
func Registered() bool {
	return prefsSrc.Bool(prefRegistered)
}

// checkPath checks that the given directory exists. If it doesn't it will be
// created.
func checkPath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to create new directory: %w", err)
		}
	}

	return nil
}
