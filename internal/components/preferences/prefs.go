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
	"sync"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	"github.com/joshuar/go-hass-agent/internal/components/validation"
)

const (
	AppName        = "Go Hass Agent"
	AppURL         = "https://github.com/joshuar/go-hass-agent"
	AppDescription = "A Home Assistant, native app for desktop/laptop devices."

	FeatureRequestURL = AppURL + "/issues/new?assignees=joshuar&labels=&template=feature_request.md&title="
	IssueURL          = AppURL + "/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D"

	preferencesFile  = "preferences.toml"
	defaultFilePerms = 0o600
	LogFile          = "go-hass-agent.log"
)

// Preference names.
const (
	prefsEnvPrefix = "GOHASSAGENT_"
	PrefsEnvAppID  = prefsEnvPrefix + "APPID"

	prefAppID      = "app_id"
	prefRegistered = "registered"
	prefHeadless   = "headless"
)

// Default values.
const (
	defaultServer          = "http://localhost:8123"
	defaultWebsocketServer = "ws://localhost:8123"
	defaultSecret          = "ALongSecretString"
	defaultMQTTTopicPrefix = "homeassistant"
	defaultMQTTServer      = "tcp://localhost:1883"
	DefaultAppID           = "go-hass-agent"
)

var (
	prefsSrc = koanf.New(".")
	mu       = sync.Mutex{}
	//lint:ignore U1000 some of these will be used in the future
	gitVersion, gitCommit, gitTreeState, buildDate string
	appVersion                                     = gitVersion
	appID                                          = DefaultAppID
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
var defaultAgentPreferences = &preferences{
	Version:    AppVersion(),
	Registered: false,
	MQTT: &MQTT{
		MQTTEnabled:     false,
		MQTTTopicPrefix: defaultMQTTTopicPrefix,
		MQTTServer:      defaultMQTTServer,
	},
	Registration: &Registration{
		Server: defaultServer,
		Token:  defaultSecret,
	},
	Hass: &Hass{
		RestAPIURL:   defaultServer,
		WebsocketURL: defaultWebsocketServer,
		WebhookID:    defaultSecret,
	},
}

type API struct{}

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

// Load will retrieve the current preferences from the preference file on disk,
// overwriting values with any passed-in preferences. If there is a problem
// during retrieval, an error will be returned.
var Load = func(ctx context.Context, preferences ...SetPreference) error {
	return sync.OnceValue(func() error {
		prefsFile := filepath.Join(PathFromCtx(ctx), preferencesFile)

		// api := &API{}
		if runtimeAppID, found := os.LookupEnv(PrefsEnvAppID); found {
			appID = runtimeAppID
		}

		slog.Debug("Loading preferences.",
			slog.String("file", prefsFile))

		// Load config file
		if err := prefsSrc.Load(file.Provider(prefsFile), toml.Parser()); err != nil {
			slog.Warn("No preferences found, using defaults.", slog.Any("error", err))
			if err := prefsSrc.Load(structs.Provider(defaultAgentPreferences, "toml"), nil); err != nil {
				return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
			}
		}
		// Unmarshal config, overwriting defaults.
		if err := prefsSrc.UnmarshalWithConf("", defaultAgentPreferences, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
			return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
		}

		// Set any preferences passed in.
		SetPreferences(preferences...)

		// Validate preferences.
		if err := validate(); err != nil {
			return fmt.Errorf("%w: %w", ErrLoadPreferences, err)
		}

		return nil
	})()
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
func Save(ctx context.Context) error {
	mu.Lock()
	defer mu.Unlock()

	if prefsSrc == nil {
		return ErrLoadPreferences
	}

	prefsFile := filepath.Join(PathFromCtx(ctx), preferencesFile)

	slog.Debug("Saving preferences.", slog.String("file", prefsFile))

	if err := validate(); err != nil {
		return err
	}

	if err := checkPath(PathFromCtx(ctx)); err != nil {
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

// SetHeadless sets an whether the agent should run headless (i.e., without a
// GUI).
func SetHeadless(value bool) SetPreference {
	return func() error {
		if err := prefsSrc.Set(prefHeadless, value); err != nil {
			return fmt.Errorf("%w: %w", ErrSetPreference, err)
		}

		return nil
	}
}

// Headless retrieves whether the agent is running headless (i.e., without a
// GUI).
func Headless() bool {
	return prefsSrc.Bool(prefHeadless)
}

// AppVersion returns the version of Go Hass Agent.
func AppVersion() string {
	if appVersion != "" {
		return appVersion
	}

	return "Unknown"
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
