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

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"

	"github.com/joshuar/go-hass-agent/internal/validation"
)

const (
	AppName           = "Go Hass Agent"
	AppURL            = "https://github.com/joshuar/go-hass-agent"
	FeatureRequestURL = AppURL + "/issues/new?assignees=joshuar&labels=&template=feature_request.md&title="
	IssueURL          = AppURL + "/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D"
	AppDescription    = "A Home Assistant, native app for desktop/laptop devices."
	AppID             = "go-hass-agent"
	MQTTTopicPrefix   = "homeassistant"
	LogFile           = "go-hass-agent.log"
	preferencesFile   = "preferences.toml"
	defaultFilePerms  = 0o600

	DefaultServer = "http://localhost:8123"
	DefaultSecret = "ALongSecretString"

	unknownValue = "Unknown"
)

var (
	//lint:ignore U1000 some of these will be used in the future
	gitVersion, gitCommit, gitTreeState, buildDate string
	AppVersion                                     = gitVersion
)

var (
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
	file         string
	Registered   bool `toml:"registered" validate:"boolean"`
}

// Load will retrieve the current preferences from the preference file on disk.
// If there is a problem during retrieval, an error will be returned.
func Load(ctx context.Context) (*Preferences, error) {
	appID := AppIDFromContext(ctx)

	file := filepath.Join(xdg.ConfigHome, appID, preferencesFile)
	prefs := DefaultPreferences(file)

	b, err := os.ReadFile(prefs.file)
	if err != nil {
		return prefs, errors.Join(ErrNoPreferences, err)
	}

	err = toml.Unmarshal(b, &prefs)
	if err != nil {
		return prefs, errors.Join(ErrFileContents, err)
	}

	return prefs, nil
}

// Reset will remove the preferences file.
func Reset(ctx context.Context) error {
	appID := AppIDFromContext(ctx)
	file := filepath.Join(xdg.ConfigHome, appID, preferencesFile)

	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return fmt.Errorf("preferences not found: %w", err)
	}

	err = os.Remove(file)
	if err != nil {
		return fmt.Errorf("unable to reset preferences: %w", err)
	}

	return nil
}

// DefaultPreferences returns a Preferences object which contains default
// values. While the default values will be valid values, they might not be
// usable/relevant values.
func DefaultPreferences(file string) *Preferences {
	if AppVersion == "" {
		AppVersion = "Unknown"
	}

	device, err := newDevice()
	if err != nil {
		slog.Warn("Problem generating new device info.", slog.Any("error", err))
	}

	return &Preferences{
		Version:      AppVersion,
		Registered:   false,
		Registration: DefaultRegistrationPreferences(),
		Hass:         DefaultHassPreferences(),
		MQTT:         DefaultMQTTPreferences(),
		Device:       device,
		mu:           &sync.Mutex{},
		file:         file,
	}
}

func (p *Preferences) Validate() error {
	err := validation.Validate.Struct(p)
	if err != nil {
		return fmt.Errorf("validation failed: %s", validation.ParseValidationErrors(err))
	}

	return nil
}

// Save will save the new values of the specified preferences to the existing
// preferences file. NOTE: if the preferences file does not exist, Save will
// return an error. Use New if saving preferences for the first time.
func (p *Preferences) Save() error {
	if err := p.Validate(); err != nil {
		return err
	}

	if err := checkPath(filepath.Dir(p.file)); err != nil {
		return err
	}

	b, err := toml.Marshal(p)
	if err != nil {
		return fmt.Errorf("unable to format preferences: %w", err)
	}

	err = os.WriteFile(p.file, b, defaultFilePerms)
	if err != nil {
		return fmt.Errorf("unable to write preferences file: %w", err)
	}

	return nil
}

func (p *Preferences) AgentVersion() string {
	return p.Version
}

func (p *Preferences) AgentRegistered() bool {
	return p.Registered
}

func (p *Preferences) RestAPIURL() string {
	if p.Hass != nil {
		return p.Hass.RestAPIURL
	}

	return ""
}

func (p *Preferences) WebsocketURL() string {
	if p.Hass != nil {
		return p.Hass.WebsocketURL
	}

	return ""
}

func (p *Preferences) WebhookID() string {
	if p.Hass != nil {
		return p.Hass.WebhookID
	}

	return ""
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
