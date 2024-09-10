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

	"github.com/pelletier/go-toml/v2"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
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

	defaultServer = "http://localhost:8123"
	defaultSecret = "ALongSecretString"
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

type MQTT struct {
	MQTTServer      string `toml:"server,omitempty" validate:"omitempty,uri" kong:"required,help='MQTT server URI. Required.',placeholder='scheme://some.host:port'"` //nolint:lll
	MQTTUser        string `toml:"user,omitempty" validate:"omitempty" kong:"optional,help='MQTT username.'"`
	MQTTPassword    string `toml:"password,omitempty" validate:"omitempty" kong:"optional,help='MQTT password.'"`
	MQTTTopicPrefix string `toml:"topic_prefix,omitempty" validate:"omitempty,ascii" kong:"optional,help='MQTT topic prefix.'"`
	MQTTEnabled     bool   `toml:"enabled" validate:"boolean" kong:"-"`
}

func (p *Preferences) Validate() error {
	err := validate.Struct(p)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrValidationFailed, parseValidationErrors(err))
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

// GetMQTTPreferences returns the subset of MQTT preferences.
func (p *Preferences) GetMQTTPreferences() *MQTT {
	return p.MQTT
}

// Load will retrieve the current preferences from the preference file on disk.
// If there is a problem during retrieval, an error will be returned.
func Load(path string) (*Preferences, error) {
	file := filepath.Join(path, preferencesFile)
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

// Reset will remove the preferences directory.
func Reset(path string) error {
	file := filepath.Join(path, preferencesFile)

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
		Version:    AppVersion,
		Registered: false,
		Registration: &Registration{
			Server: defaultServer,
			Token:  defaultSecret,
		},
		Hass: &Hass{
			IgnoreHassURLs: false,
			WebhookID:      defaultSecret,
			RestAPIURL:     defaultServer,
			WebsocketURL:   defaultServer,
		},
		MQTT:   &MQTT{MQTTEnabled: false},
		Device: device,
		mu:     &sync.Mutex{},
		file:   file,
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

func (p *Preferences) DeviceName() string {
	if p.Device != nil {
		return p.Device.Name
	}

	return "Unknown Device"
}

func (p *Preferences) DeviceID() string {
	if p.Device != nil {
		return p.Device.ID
	}

	return "Unknown Device ID"
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

func (p *Preferences) Token() string {
	if p.Registration != nil {
		return p.Registration.Token
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
