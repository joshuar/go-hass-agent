// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package preferences

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/validation"
)

const (
	// Global informational strings.
	AppName           = "Go Hass Agent"
	AppURL            = "https://github.com/joshuar/go-hass-agent"
	AppDescription    = "A Home Assistant, native app for desktop/laptop devices."
	FeatureRequestURL = AppURL + "/issues/new?assignees=joshuar&labels=&template=feature_request.md&title="
	IssueURL          = AppURL + "/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D"
	// Internal file defaults.
	preferencesFile  = "preferences.toml"
	defaultFilePerms = 0o600
	LogFile          = "go-hass-agent.log"
	// Preference names.
	prefAppID      = "app_id"
	prefRegistered = "registered"
	prefHeadless   = "headless"
	prefVersion    = "version"
	// Default values.
	DefaultServer          = "http://localhost:8123"
	defaultWebsocketServer = "ws://localhost:8123"
	defaultSecret          = "ALongSecretString"
	defaultMQTTTopicPrefix = "homeassistant"
	defaultMQTTServer      = "tcp://localhost:1883"
	DefaultAppID           = "go-hass-agent"
	PathDelim              = "."
)

// preferences defines all preferences for Go Hass Agent.
type preferences struct {
	MQTT         *MQTTPreferences `toml:"mqtt,omitempty"`
	Registration *Registration    `toml:"registration"`
	Hass         *Hass            `toml:"hass"`
	Device       *Device          `toml:"device"`
	Version      string           `toml:"version"`
	Registered   bool             `toml:"registered" validate:"boolean"`
}

var (
	// Package level internal variables.
	prefsSrc  = koanf.New(PathDelim)
	mu        = sync.Mutex{}
	prefsFile string
	//lint:ignore U1000 some of these will be used in the future
	gitVersion, gitCommit, gitTreeState, buildDate string
	appID                                          = DefaultAppID
	appVersion                                     = gitVersion
)

// Consistent error messages.
var (
	ErrLoadPreferences      = errors.New("error loading preferences")
	ErrSavePreferences      = errors.New("error saving preferences")
	ErrValidatePreferences  = errors.New("error validating preferences")
	ErrSetPreference        = errors.New("error setting preference")
	ErrPreferencesNotLoaded = errors.New("preferences not loaded")
)

// Default agent preferences.
var defaultPreferences = &preferences{
	Version:    AppVersion(),
	Registered: false,
	MQTT:       defaultMQTTPreferences,
	Registration: &Registration{
		Server: DefaultServer,
		Token:  defaultSecret,
	},
	Hass: &Hass{
		RestAPIURL:   DefaultServer,
		WebsocketURL: defaultWebsocketServer,
		WebhookID:    defaultSecret,
	},
}

// Init will retrieve the current preferences from the preference file on disk,
// overwriting values with any passed-in preferences. If there is a problem
// during retrieval, an error will be returned.
var Init = func(ctx context.Context, preferences ...SetPreference) error {
	return sync.OnceValue(func() error {
		prefsFile = filepath.Join(PathFromCtx(ctx), preferencesFile)
		ctx = slogctx.WithGroup(ctx, "preferences")

		slogctx.FromCtx(ctx).Debug("Loading preferences.",
			slog.String("file", prefsFile))

		// Load config file
		if err := prefsSrc.Load(file.Provider(prefsFile), toml.Parser()); err != nil {
			slogctx.FromCtx(ctx).Debug("No preferences found, using defaults.", slog.Any("error", err))
			if err := prefsSrc.Load(structs.Provider(defaultPreferences, "toml"), nil); err != nil {
				return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
			}
		}
		// Unmarshal config, overwriting defaults.
		if err := prefsSrc.UnmarshalWithConf("", defaultPreferences, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
			return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
		}

		// Set any preferences passed in.
		if len(preferences) > 0 {
			if err := Set(preferences...); err != nil {
				slogctx.FromCtx(ctx).Debug("Could not set initial custom preferences.",
					slog.Any("error", err))
			}
		}

		// Validate preferences.
		if err := validate(); err != nil {
			return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
		}

		return nil
	})()
}

// load will load the preferences for a component, using the given file and
// environment prefixes, and marshaling the config into the given config object.
func load(configPrefix string, cfg any) error {
	// Load config file
	if err := prefsSrc.Load(file.Provider(prefsFile), toml.Parser()); err != nil {
		return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
	}
	// Unmarshal config, overwriting defaults.
	if err := prefsSrc.UnmarshalWithConf(configPrefix, cfg, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
	}

	slog.Debug("Loading config for component.",
		slog.String("component", configPrefix))

	return nil
}

// validate ensures the configuration is valid.
func validate() error {
	currentPreferences := &preferences{}

	// Unmarshal current preferences.
	if err := prefsSrc.UnmarshalWithConf("", currentPreferences, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
	}

	// Ensure hass preferences are valid.
	if err := validation.Validate.Struct(currentPreferences.Hass); err != nil {
		return fmt.Errorf("%w: %s", ErrValidatePreferences, validation.ParseValidationErrors(err))
	}

	if currentPreferences.MQTT != nil {
		if currentPreferences.MQTT.MQTTEnabled {
			// Validate MQTT preferences are valid.
			err := validation.Validate.Struct(currentPreferences.MQTT)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrValidatePreferences, validation.ParseValidationErrors(err))
			}
		}
	}

	slog.Debug("Preferences are valid.")

	return nil
}

// save will save the new values of the specified preferences to the existing
// preferences file.
func save() error {
	mu.Lock()
	defer mu.Unlock()

	if prefsSrc == nil {
		return ErrLoadPreferences
	}

	slog.Debug("Saving preferences.", slog.String("file", prefsFile))

	if err := prefsSrc.Set(prefVersion, appVersion); err != nil {
		slog.Warn("Cannot update version in preferences file.",
			slog.Any("error", err))
	}

	if err := validate(); err != nil {
		return err
	}

	if err := checkPath(filepath.Dir(prefsFile)); err != nil {
		return err
	}

	b, err := prefsSrc.Marshal(toml.Parser())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSavePreferences, err)
	}

	err = os.WriteFile(prefsFile, b, defaultFilePerms)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSavePreferences, err)
	}

	return nil
}

// SetPreference will set a preference in the preferences
// store. If it fails, it will return a non-nil error.
type SetPreference func() error

// Set will set the given preferences. It will emit WARN level log
// messages for each preference that failed to get set.
func Set(preferences ...SetPreference) error {
	for _, preference := range preferences {
		if err := preference(); err != nil {
			slog.Warn("Error setting preference.",
				slog.Any("error", err))
		}
	}

	if err := save(); err != nil {
		return errors.Join(ErrSavePreferences, err)
	}

	return nil
}

// SetRegistered sets whether Go Hass Agent has been registered with Home
// Assistant.
func SetRegistered(value bool) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefRegistered, value); err != nil {
			return fmt.Errorf("%w: %w", ErrSetPreference, err)
		}

		return nil
	}
}

// Registered returns the registration status of Go Hass Agent.
func Registered() bool {
	return prefsSrc.Bool(prefRegistered)
}

// AppVersion returns the version of Go Hass Agent.
func AppVersion() string {
	if appVersion != "" {
		return appVersion
	}

	return "Unknown"
}

// Version returns the version of the preferences file (i.e., the
// last version of Go Hass Agent to write the preferences.toml file).
func Version() string {
	return prefsSrc.String(prefVersion)
}

// Reset will remove the preferences file.
func Reset(ctx context.Context) error {
	prefsFile := filepath.Join(PathFromCtx(ctx), preferencesFile)

	slog.Debug("Removing preferences.", slog.String("file", prefsFile))

	_, err := os.Stat(prefsFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
	}

	err = os.Remove(prefsFile)
	if err != nil {
		return fmt.Errorf("unable to reset preferences: %w", err)
	}

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
