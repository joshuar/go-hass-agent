// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package preferences

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
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
	LogFile             = "go-hass-agent.log"
	defaultFilePerms    = 0o600
	preferencesFilename = "preferences.toml"

	prefRegistered = "registered"
)

const (
	defaultServer          = "http://localhost:8123"
	defaultWebsocketServer = "ws://localhost:8123"
	defaultSecret          = "ALongSecretString"
	defaultMQTTTopicPrefix = "homeassistant"
	defaultMQTTServer      = "tcp://localhost:1883"
)

var (
	//lint:ignore U1000 some of these will be used in the future
	gitVersion, gitCommit, gitTreeState, buildDate string
	appVersion                                     = gitVersion
	appID                                          = "go-hass-agent"
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
	Version:    AppVersion(),
	Registered: false,
	MQTT: &MQTT{
		MQTTEnabled:     false,
		MQTTTopicPrefix: defaultMQTTTopicPrefix,
	},
	Registration: &Registration{
		Server: defaultServer,
		Token:  defaultSecret,
	},
	Hass: &Hass{
		RestAPIURL:   defaultServer,
		WebsocketURL: defaultServer,
		WebhookID:    defaultSecret,
	},
}

var (
	prefsSrc        = koanf.New(".")
	preferencesFile = filepath.Join(xdg.ConfigHome, appID, preferencesFilename)
	mu              = sync.Mutex{}
)

// SetPreference will set a preference in the preferences
// store. If it fails, it will return a non-nil error.
type SetPreference func() error

// SetPreferences will set the given preferences. It will emit WARN level log
// messages for each preference that failed to get set.
func SetPreferences(preferences ...SetPreference) {
	for _, preference := range preferences {
		if err := preference(); err != nil {
			slog.Warn("Error setting preference.",
				slog.Any("error", err))
		}
	}
}

// preferences defines all preferences for Go Hass Agent.
type preferences struct {
	MQTT         *MQTT         `toml:"mqtt,omitempty"`
	Registration *Registration `toml:"registration"`
	Hass         *Hass         `toml:"hass"`
	Device       *Device       `toml:"device"`
	Version      string        `toml:"version"`
	Registered   bool          `toml:"registered" validate:"boolean"`
}

// Load will retrieve the current preferences from the preference file on disk.
// If there is a problem during retrieval, an error will be returned.
var Load = func() error {
	return sync.OnceValue(func() error {
		preferencesFile = filepath.Join(Path(), preferencesFilename)

		slog.Debug("Loading preferences.", slog.String("file", preferencesFile))

		// Load config file
		if err := prefsSrc.Load(file.Provider(preferencesFile), toml.Parser()); err != nil {
			slog.Warn("No preferences found, using defaults.", slog.Any("error", err))
			if err := prefsSrc.Load(structs.Provider(defaultAgentPreferences, "toml"), nil); err != nil {
				return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
			}
		}
		// Unmarshal config, overwriting defaults.
		if err := prefsSrc.UnmarshalWithConf("", defaultAgentPreferences, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
			return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
		}

		// Validate preferences.
		return validate()
	})()
}

// Reset will remove the preferences file.
func Reset() error {
	preferencesFile = filepath.Join(Path(), preferencesFilename)

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

// Save will save the new values of the specified preferences to the existing
// preferences file. NOTE: if the preferences file does not exist, Save will
// return an error. Use New if saving preferences for the first time.
func Save() error {
	mu.Lock()
	defer mu.Unlock()

	preferencesFile = filepath.Join(Path(), preferencesFilename)

	slog.Debug("Saving preferences.", slog.String("file", preferencesFile))

	if err := validate(); err != nil {
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

// SetAppID sets an ID that is used as part of the path to the preferences file.
func SetAppID(id string) {
	appID = id
}

// AppID retrieves the ID.
func AppID() string {
	return appID
}

// AppVersion returns the version of Go Hass Agent.
func AppVersion() string {
	if appVersion != "" {
		return appVersion
	}

	return "Unknown"
}

// Path returns a path where preferences are stored.
func Path() string {
	return filepath.Join(xdg.ConfigHome, appID)
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
