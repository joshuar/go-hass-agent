// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver,comment-spacings
package preferences

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	mqtthass "github.com/joshuar/go-hass-anything/v9/pkg/hass"
	"github.com/pelletier/go-toml/v2"
	"github.com/sourcegraph/conc/pool"
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
	preferencesPath = filepath.Join(xdg.ConfigHome, "go-hass-agent")
	preferencesFile = "preferences.toml"

	ErrNoPreferences = errors.New("no preferences file found, using defaults")
	ErrFileContents  = errors.New("could not read file contents")
)

//nolint:tagalign
type Preferences struct {
	mu             *sync.Mutex
	Version        string `toml:"agent.version" validate:"required"`
	Host           string `toml:"registration.host" validate:"required,http_url"`
	Token          string `toml:"registration.token" validate:"required,ascii"`
	DeviceID       string `toml:"device.id" validate:"required,ascii"`
	DeviceName     string `toml:"device.name" validate:"required,ascii"`
	RestAPIURL     string `toml:"hass.apiurl,omitempty" validate:"http_url,required_without=CloudhookURL RemoteUIURL"`
	CloudhookURL   string `toml:"hass.cloudhookurl,omitempty" validate:"omitempty,http_url"`
	WebsocketURL   string `toml:"hass.websocketurl" validate:"required,url"`
	WebhookID      string `toml:"hass.webhookid" validate:"required,ascii"`
	RemoteUIURL    string `toml:"hass.remoteuiurl,omitempty" validate:"omitempty,http_url"`
	Secret         string `toml:"hass.secret,omitempty" validate:"omitempty"`
	MQTTPassword   string `toml:"mqtt.password,omitempty" validate:"omitempty"`
	MQTTUser       string `toml:"mqtt.user,omitempty" validate:"omitempty"`
	MQTTServer     string `toml:"mqtt.server,omitempty" validate:"omitempty,uri"`
	Registered     bool   `toml:"hass.registered" validate:"boolean"`
	MQTTEnabled    bool   `toml:"mqtt.enabled" validate:"boolean"`
	MQTTRegistered bool   `toml:"mqtt.registered" validate:"boolean"`
}

type Preference func(*Preferences) error

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

func SetVersion(version string) Preference {
	return func(p *Preferences) error {
		p.Version = version

		return nil
	}
}

func SetDeviceID(id string) Preference {
	return func(p *Preferences) error {
		p.DeviceID = id

		return nil
	}
}

func SetDeviceName(name string) Preference {
	return func(p *Preferences) error {
		p.DeviceName = name

		return nil
	}
}

func SetRestAPIURL(url string) Preference {
	return func(p *Preferences) error {
		p.RestAPIURL = url

		return nil
	}
}

func SetCloudhookURL(url string) Preference {
	return func(p *Preferences) error {
		p.CloudhookURL = url

		return nil
	}
}

func SetRemoteUIURL(url string) Preference {
	return func(p *Preferences) error {
		p.RemoteUIURL = url

		return nil
	}
}

func SetSecret(secret string) Preference {
	return func(p *Preferences) error {
		p.Secret = secret

		return nil
	}
}

func SetHost(host string) Preference {
	return func(p *Preferences) error {
		p.Host = host

		return nil
	}
}

func SetToken(token string) Preference {
	return func(p *Preferences) error {
		p.Token = token

		return nil
	}
}

func SetWebhookID(id string) Preference {
	return func(p *Preferences) error {
		p.WebhookID = id

		return nil
	}
}

func SetWebsocketURL(url string) Preference {
	return func(p *Preferences) error {
		p.WebsocketURL = url

		return nil
	}
}

func SetRegistered(status bool) Preference {
	return func(p *Preferences) error {
		p.Registered = status

		return nil
	}
}

func SetMQTTEnabled(status bool) Preference {
	return func(p *Preferences) error {
		p.MQTTEnabled = status

		return nil
	}
}

func SetMQTTServer(server string) Preference {
	return func(p *Preferences) error {
		p.MQTTServer = server

		return nil
	}
}

func SetMQTTUser(user string) Preference {
	return func(p *Preferences) error {
		p.MQTTUser = user

		return nil
	}
}

func SetMQTTPassword(password string) Preference {
	return func(p *Preferences) error {
		p.MQTTPassword = password

		return nil
	}
}

func SetMQTTRegistered(status bool) Preference {
	return func(p *Preferences) error {
		p.MQTTRegistered = status

		return nil
	}
}

func (p *Preferences) GetMQTTEnabled() bool {
	return p.MQTTEnabled
}

// MQTTServer returns the broker URI from the preferences.
func (p *Preferences) GetMQTTServer() string {
	return p.MQTTServer
}

// MQTTUser returns any username required for connecting to the broker from the
// preferences.
func (p *Preferences) GetMQTTUser() string {
	return p.MQTTUser
}

// MQTTPassword returns any password required for connecting to the broker from the
// preferences.
func (p *Preferences) GetMQTTPassword() string {
	return p.MQTTPassword
}

// GetTopicPrefix returns the prefix for topics on MQTT.
func (p *Preferences) GetTopicPrefix() string {
	return "homeassistant"
}

//nolint:exhaustruct
func defaultPreferences() *Preferences {
	return &Preferences{
		Version:      AppVersion,
		Host:         "http://localhost:8123",
		WebsocketURL: "http://localhost:8123",
		Token:        "replaceMe",
		WebhookID:    "replaceMe",
		Registered:   false,
		MQTTEnabled:  false,
		mu:           &sync.Mutex{},
	}
}

// Load will retrieve the current preferences from the preference file on disk.
// If there is a problem during retrieval, an error will be returned.
func Load() (*Preferences, error) {
	file := filepath.Join(preferencesPath, preferencesFile)
	prefs := defaultPreferences()

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

// Save will save the new values of the specified preferences to the existing
// preferences file. NOTE: if the preferences file does not exist, Save will
// return an error. Use New if saving preferences for the first time.
func Save(setters ...Preference) error {
	if err := checkPath(preferencesPath); err != nil {
		return err
	}

	prefs, err := Load()
	if err != nil && !errors.Is(err, ErrNoPreferences) {
		return err
	}

	if err := set(prefs, setters...); err != nil {
		return err
	}

	if err := validatePreferences(prefs); err != nil {
		return showValidationErrors(err)
	}

	file := filepath.Join(preferencesPath, preferencesFile)

	return write(prefs, file)
}

// Reset will remove the preferences directory.
func Reset() error {
	err := os.RemoveAll(preferencesPath)
	if err != nil {
		return fmt.Errorf("unable to reset preferences: %w", err)
	}

	return nil
}

func set(prefs *Preferences, setters ...Preference) error {
	taskPool := pool.New().WithErrors()

	for _, setter := range setters {
		setPref := setter

		taskPool.Go(func() error {
			prefs.mu.Lock()
			defer prefs.mu.Unlock()

			return setPref(prefs)
		})
	}

	err := taskPool.Wait()
	if err != nil {
		return fmt.Errorf("problem detected setting preferences: %w", err)
	}

	return nil
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

// MQTTOrigin defines Go Hass Agent as the origin for MQTT functionality.
func MQTTOrigin() *mqtthass.Origin {
	return &mqtthass.Origin{
		Name:    AppName,
		Version: AppVersion,
		URL:     AppURL,
	}
}
