// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
//revive:disable:unused-receiver,comment-spacings
package preferences

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	"github.com/pelletier/go-toml/v2"
)

const (
	AppName           = "Go Hass Agent"
	AppURL            = "https://github.com/joshuar/go-hass-agent"
	FeatureRequestURL = AppURL + "/issues/new?assignees=joshuar&labels=&template=feature_request.md&title="
	IssueURL          = AppURL + "/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D"
	AppDescription    = "A Home Assistant, native app for desktop/laptop devices."
	AppID             = "com.github.joshuar.go-hass-agent"
	MQTTTopicPrefix   = "homeassistant"
	LogFile           = "go-hass-agent.log"
)

var (
	//lint:ignore U1000 some of these will be used in the future
	gitVersion, gitCommit, gitTreeState, buildDate string
	AppVersion                                     = gitVersion
)

var (
	preferencesPath = filepath.Join(xdg.ConfigHome, AppID)
	preferencesFile = "preferences.toml"

	ErrNoPreferences = errors.New("no preferences file found, using defaults")
	ErrFileContents  = errors.New("could not read file contents")
)

//nolint:tagalign
type Preferences struct {
	mu           *sync.Mutex
	MQTT         *MQTT         `toml:"mqtt,omitempty"`
	Registration *Registration `toml:"registration"`
	Hass         *Hass         `toml:"hass"`
	Device       *Device       `toml:"device"`
	Version      string        `toml:"version" validate:"required"`
	Registered   bool          `toml:"registered" validate:"boolean"`
}

type MQTT struct {
	MQTTServer      string `toml:"server,omitempty" validate:"omitempty,uri"`
	MQTTUser        string `toml:"user,omitempty" validate:"omitempty"`
	MQTTPassword    string `toml:"password,omitempty" validate:"omitempty"`
	MQTTTopicPrefix string `toml:"topic_prefix,omitempty" validate:"omitempty,ascii"`
	MQTTEnabled     bool   `toml:"enabled" validate:"boolean"`
}

// SetPath sets the path to the preferences file to the given path. If this
// function is not called, a default path is used.
func SetPath(path string) {
	preferencesPath = path
}

// SetFile sets the filename of the preferences file to the given name. If this
// function is not called, a default filename is used.
func SetFile(name string) {
	preferencesFile = name
}

// Path returns the current path to the preferences file. Use GetFile to
// retrieve the filename.
func Path() string {
	return preferencesPath
}

// File returns the filename of the preferences file. Use GetPath to retrieve
// its path.
func File() string {
	return preferencesFile
}

func (p *Preferences) Validate() error {
	err := validate.Struct(p)
	if err != nil {
		showValidationErrors(err)

		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// Save will save the new values of the specified preferences to the existing
// preferences file. NOTE: if the preferences file does not exist, Save will
// return an error. Use New if saving preferences for the first time.
func (p *Preferences) Save() error {
	if err := checkPath(preferencesPath); err != nil {
		return err
	}

	if err := p.Validate(); err != nil {
		return err
	}

	file := filepath.Join(preferencesPath, preferencesFile)

	return write(p, file)
}

// GetMQTTPreferences returns the subset of MQTT preferences.
func (p *Preferences) GetMQTTPreferences() *MQTT {
	return p.MQTT
}

// Load will retrieve the current preferences from the preference file on disk.
// If there is a problem during retrieval, an error will be returned.
func Load() (*Preferences, error) {
	file := filepath.Join(preferencesPath, preferencesFile)
	prefs := DefaultPreferences()

	b, err := os.ReadFile(file)
	if err != nil {
		return prefs, errors.Join(ErrNoPreferences, err)
	}

	err = toml.Unmarshal(b, &prefs)
	if err != nil {
		return prefs, errors.Join(ErrFileContents, err)
	}

	return prefs, nil
}

// Reset will remove the preferences directory.
func Reset() error {
	err := os.RemoveAll(preferencesPath)
	if err != nil {
		return fmt.Errorf("unable to reset preferences: %w", err)
	}

	return nil
}

// DefaultPreferences returns a Preferences object which contains default
// values. While the default values will be valid values, they might not be
// usable/relevant values.
func DefaultPreferences() *Preferences {
	if AppVersion == "" {
		AppVersion = "Unknown"
	}

	device, err := newDevice()
	if err != nil {
		slog.Warn("Problem generating new device info.", "error", err.Error())
	}

	return &Preferences{
		Version:    AppVersion,
		Registered: false,
		Registration: &Registration{
			Server: "http://localhost:8123",
			Token:  "ASecretLongLivedToken",
		},
		Hass: &Hass{
			IgnoreHassURLs: false,
			WebhookID:      "ALongString",
			RestAPIURL:     "https://localhost:8123",
			WebsocketURL:   "https://localhost:8123",
		},
		MQTT:   &MQTT{MQTTEnabled: false},
		Device: device,
		mu:     &sync.Mutex{},
	}
}

// IsMQTTEnabled is a conveinience function to determine whether MQTT
// functionality has been enabled in the agent.
func (p *MQTT) IsMQTTEnabled() bool {
	return p.MQTTEnabled
}

// Server returns the broker URI from the preferences.
func (p *MQTT) Server() string {
	return p.MQTTServer
}

// User returns any username required for connecting to the broker from the
// preferences.
func (p *MQTT) User() string {
	return p.MQTTUser
}

// Password returns any password required for connecting to the broker from the
// preferences.
func (p *MQTT) Password() string {
	return p.MQTTPassword
}

// TopicPrefix returns the prefix for topics on MQTT.
func (p *MQTT) TopicPrefix() string {
	if p.MQTTTopicPrefix == "" {
		return MQTTTopicPrefix
	}

	return p.MQTTTopicPrefix
}

// MQTTOrigin defines Go Hass Agent as the origin for MQTT functionality.
func MQTTOrigin() *mqtthass.Origin {
	return &mqtthass.Origin{
		Name:    AppName,
		Version: AppVersion,
		URL:     AppURL,
	}
}

//nolint:mnd
func write(prefs *Preferences, file string) error {
	b, err := toml.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("unable to format preferences: %w", err)
	}

	err = os.WriteFile(file, b, 0o600)
	if err != nil {
		return fmt.Errorf("unable to write preferences file: %w", err)
	}

	return nil
}

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
