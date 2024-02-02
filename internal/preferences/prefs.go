// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package preferences

import (
	_ "embed"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/sync/errgroup"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var AppVersion string

var (
	preferencesPath = filepath.Join(xdg.ConfigHome, "go-hass-agent")
	preferencesFile = "preferences.toml"
)

type Preferences struct {
	mu           *sync.Mutex
	Version      string `toml:"agent.version" validate:"required"`
	Host         string `toml:"registration.host" validate:"required,http_url"`
	Token        string `toml:"registration.token" validate:"required,ascii"`
	ID           string `toml:"device.id" validate:"required,ascii"`
	Name         string `toml:"device.name" validate:"required,hostname"`
	RestAPIURL   string `toml:"hass.apiurl,omitempty" validate:"http_url,required_without=CloudhookURL RemoteUIURL"`
	CloudhookURL string `toml:"hass.cloudhookurl,omitempty" validate:"omitempty,http_url"`
	WebsocketURL string `toml:"hass.websocketurl" validate:"required,url"`
	WebhookID    string `toml:"hass.webhookid" validate:"required,ascii"`
	RemoteUIURL  string `toml:"hass.remoteuiurl,omitempty" validate:"omitempty,http_url"`
	Secret       string `toml:"hass.secret,omitempty" validate:"omitempty"`
	MQTTPassword string `toml:"mqtt.password,omitempty" validate:"omitempty"`
	MQTTUser     string `toml:"mqtt.user,omitempty" validate:"omitempty"`
	MQTTServer   string `toml:"mqtt.server,omitempty" validate:"omitempty,uri"`
	Registered   bool   `toml:"hass.registered" validate:"boolean"`
	MQTTEnabled  bool   `toml:"mqtt.enabled" validate:"boolean"`
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

// GetPath returns the current path to the preferences file. Use GetFile to
// retrieve the filename.
func GetPath() string {
	return preferencesPath
}

// GetFile returns the filename of the preferences file. Use GetPath to retrieve
// its path.
func GetFile() string {
	return preferencesFile
}

func Version(version string) Preference {
	return func(p *Preferences) error {
		p.Version = version
		return nil
	}
}

func ID(id string) Preference {
	return func(p *Preferences) error {
		p.ID = id
		return nil
	}
}

func Name(name string) Preference {
	return func(p *Preferences) error {
		p.Name = name
		return nil
	}
}

func RestAPIURL(url string) Preference {
	return func(p *Preferences) error {
		p.RestAPIURL = url
		return nil
	}
}

func CloudhookURL(url string) Preference {
	return func(p *Preferences) error {
		p.CloudhookURL = url
		return nil
	}
}

func RemoteUIURL(url string) Preference {
	return func(p *Preferences) error {
		p.RemoteUIURL = url
		return nil
	}
}

func Secret(secret string) Preference {
	return func(p *Preferences) error {
		p.Secret = secret
		return nil
	}
}

func Host(host string) Preference {
	return func(p *Preferences) error {
		p.Host = host
		return nil
	}
}

func Token(token string) Preference {
	return func(p *Preferences) error {
		p.Token = token
		return nil
	}
}

func WebhookID(id string) Preference {
	return func(p *Preferences) error {
		p.WebhookID = id
		return nil
	}
}

func WebsocketURL(url string) Preference {
	return func(p *Preferences) error {
		p.WebsocketURL = url
		return nil
	}
}

func Registered(status bool) Preference {
	return func(p *Preferences) error {
		p.Registered = status
		return nil
	}
}

func MQTTEnabled(status bool) Preference {
	return func(p *Preferences) error {
		p.MQTTEnabled = status
		return nil
	}
}

func MQTTServer(server string) Preference {
	return func(p *Preferences) error {
		p.MQTTServer = server
		return nil
	}
}

func MQTTUser(user string) Preference {
	return func(p *Preferences) error {
		p.MQTTUser = user
		return nil
	}
}

func MQTTPassword(password string) Preference {
	return func(p *Preferences) error {
		p.MQTTPassword = password
		return nil
	}
}

func defaultPreferences() *Preferences {
	return &Preferences{
		Version: AppVersion,
		mu:      &sync.Mutex{},
	}
}

// Load will retrieve the current preferences from the preference file on disk.
// If there is a problem during retrieval, an error will be returned.
func Load() (*Preferences, error) {
	file := filepath.Join(preferencesPath, preferencesFile)
	prefs := defaultPreferences()

	b, err := os.ReadFile(file)
	if err != nil {
		return prefs, err
	}

	err = toml.Unmarshal(b, &prefs)
	if err != nil {
		return prefs, err
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
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := set(prefs, setters...); err != nil {
		return err
	}

	if err := validatePreferences(prefs); err != nil {
		err := showValidationErrors(err)
		return err
	}

	file := filepath.Join(preferencesPath, preferencesFile)
	return write(prefs, file)
}

func set(prefs *Preferences, setters ...Preference) error {
	g := new(errgroup.Group)
	for _, setter := range setters {
		setPref := setter
		g.Go(func() error {
			prefs.mu.Lock()
			defer prefs.mu.Unlock()
			return setPref(prefs)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func write(prefs *Preferences, file string) error {
	b, err := toml.Marshal(prefs)
	if err != nil {
		return err
	}
	err = os.WriteFile(file, b, 0o600)
	if err != nil {
		return err
	}
	return nil
}

func checkPath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}
