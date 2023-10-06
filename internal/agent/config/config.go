// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	_ "embed"
	"path/filepath"

	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/go-playground/validator/v10"
	fyneconfig "github.com/joshuar/go-hass-agent/internal/agent/config/fyneConfig"
	viperconfig "github.com/joshuar/go-hass-agent/internal/agent/config/viperConfig"
	"github.com/joshuar/go-hass-agent/internal/tracker/registry"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
)

// AgentConfig represents the methods that the agent uses to interact with
// its config. It is effectively a CRUD interface to wherever the configuration
// is stored.
//
//go:generate moq -out mockAgentConfig_test.go . AgentConfig
type AgentConfig interface {
	Get(string, interface{}) error
	Set(string, interface{}) error
	Delete(string) error
	StoragePath(string) (string, error)
}

type ConfigFileNotFoundError struct {
	Err error
}

func (e *ConfigFileNotFoundError) Error() string {
	return e.Err.Error()
}

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
	AppName       = "go-hass-agent"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var AppVersion string

// ValidateConfig takes an AgentConfig and ensures that it meets the minimum
// requirements for the agent to function correctly
func ValidateConfig(c AgentConfig) error {
	log.Debug().Msg("Running ValidateConfig.")
	cfgValidator := validator.New()

	validate := func(key, rules, errMsg string) error {
		var value string
		err := c.Get(key, &value)
		if err != nil {
			return fmt.Errorf("unable to retrieve %s from config: %v", key, err)
		}
		err = cfgValidator.Var(value, rules)
		if err != nil {
			return errors.New(errMsg)
		}
		return nil
	}

	if err := validate(PrefAPIURL,
		"required,url",
		"apiURL does not match either a URL, hostname or hostname:port",
	); err != nil {
		return err
	}
	if err := validate(PrefWebsocketURL,
		"required,url",
		"websocketURL does not match either a URL, hostname or hostname:port",
	); err != nil {
		return err
	}
	if err := validate(PrefToken,
		"required,ascii",
		"invalid long-lived token format",
	); err != nil {
		return err
	}
	if err := validate(PrefWebhookID,
		"required,ascii",
		"invalid webhookID format",
	); err != nil {
		return err
	}

	return nil
}

// UpgradeConfig checks for and performs various fixes and
// changes to the agent config as it has evolved in different versions.
func UpgradeConfig(path string) error {
	log.Debug().Msg("Running UpgradeConfig.")
	var configVersion string
	// retrieve the configVersion, or the version of the app that last read/validated the config.
	if semver.Compare(AppVersion, "v5.0.0") < 0 {
		fc := fyneconfig.NewFyneConfig()
		if err := fc.Get("Version", &configVersion); err != nil {
			return &ConfigFileNotFoundError{
				Err: errors.New("could not retrieve config version"),
			}
		}
	} else {
		vc, err := viperconfig.New(path)
		if err != nil {
			return &ConfigFileNotFoundError{
				Err: errors.New("could not open viper config"),
			}
		}
		if err := vc.Get("Version", &configVersion); err != nil {
			return &ConfigFileNotFoundError{
				Err: errors.New("could not retrieve config version"),
			}
		}
	}

	// depending on the configVersion, do the appropriate upgrades. Note that
	// some switch statements will need to fallthrough as some require previous
	// upgrades to have happened. No doubt at some point, this becomes
	// intractable and the upgrade path will need to be truncated at some
	// previous version.
	log.Debug().Msgf("Checking for upgrades needed for config version %s.", configVersion)
	switch {
	// * Upgrade host to include scheme for versions < v.1.4.0
	case semver.Compare(configVersion, "v1.4.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v1.4.0")
		fc := fyneconfig.NewFyneConfig()
		var host string

		if err := fc.Get("Host", &host); err != nil {
			return fmt.Errorf("upgrade < v.1.4.0: invalid host value (%v)", err)
		}
		var tlsBool bool
		if err := fc.Get("UseTLS", &tlsBool); err != nil {
			return fmt.Errorf("upgrade < v.1.4.0: invalid TLS value (%v)", err)
		}

		switch tlsBool {
		case true:
			host = "https://" + host
		case false:
			host = "http://" + host
		}
		if err := fc.Set("Host", host); err != nil {
			return err
		}
		fallthrough
	// * Add ApiURL and WebSocketURL config options for versions < v1.4.3
	case semver.Compare(configVersion, "v1.4.3") < 0:
		log.Debug().Msg("Performing config upgrades for < v1.4.3")
		fc := fyneconfig.NewFyneConfig()
		var host, cloudhookURL, remoteUIURL, webhookID string

		if err := fc.Get("Host", &host); err != nil {
			return fmt.Errorf("upgrade < v.1.4.3: invalid host value (%v)", err)
		}
		if err := fc.Get("CloudHookURL", &cloudhookURL); err != nil {
			return fmt.Errorf("upgrade < v.1.4.3: invalid cloudhookurl value (%v)", err)
		}
		if err := fc.Get("RemoteUIURL", &remoteUIURL); err != nil {
			return fmt.Errorf("upgrade < v.1.4.3: invalid remoteUIURL value (%v)", err)
		}
		if err := fc.Get("WebhookID", &webhookID); err != nil {
			return fmt.Errorf("upgrade < v.1.4.3: invalid webhookID value (%v)", err)
		}

		apiURL := generateAPIURL(host, cloudhookURL, remoteUIURL, webhookID)
		if apiURL == "" {
			return errors.New("could not generate apiURL")
		}
		if err := fc.Set("ApiURL", apiURL); err != nil {
			return err
		}

		websocketURL := generateWebsocketURL(host)
		if websocketURL == "" {
			return errors.New("could not generate websocketURL")
		}
		if err := fc.Set("WebSocketURL", apiURL); err != nil {
			return err
		}

		fallthrough
	// * Switch to jsonFiles registry
	case semver.Compare(configVersion, "v3.0.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v3.0.0.")
		fc := fyneconfig.NewFyneConfig()
		var err error

		p, err := fc.StoragePath("sensorRegistry")
		if err != nil {
			return errors.New("could not get sensor registry path from config")
		}
		if _, err = os.Stat(p + "/0.dat"); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		err = registry.MigrateNuts2Json(p)
		if err != nil {
			return errors.Join(errors.New("failed to migrate sensor registry"), err)
		}
		if err = os.Remove(p + "/0.dat"); err != nil {
			return errors.New("could not remove old sensor registry")
		}
		fallthrough
	// * Switch to Viper config
	case semver.Compare(configVersion, "v5.0.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v5.0.0.")
		// migrate config values
		if err := viperToFyne(path); err != nil {
			return errors.Join(errors.New("failed to migrate Fyne config to Viper"), err)
		}
		// migrate registry directory. This is non-critical, entities will be
		// re-registered if this fails.
		fc := fyneconfig.NewFyneConfig()
		oldReg, err := fc.StoragePath("sensorRegistry")
		newReg := filepath.Join(path, "sensorRegistry")
		if err != nil {
			log.Warn().Err(err).Msg("Unable to retrieve old storage path. Registry will not be migrated.")
			return nil
		}
		if err := os.Rename(oldReg, newReg); err != nil {
			log.Warn().Err(err).Msg("failed to migrate registry")
			return nil
		}
	}
	return nil
}

func generateWebsocketURL(host string) string {
	// TODO: look into websocket http upgrade method
	baseURL, err := url.Parse(host)
	if err != nil {
		log.Warn().Err(err).Msg("Host string not a URL. Cannot generate websocket URL.")
		return ""
	}
	switch baseURL.Scheme {
	case "https":
		baseURL.Scheme = "wss"
	case "http":
		baseURL.Scheme = "ws"
	default:
		log.Warn().Msg("Unknown URL scheme.")
		return ""
	}
	baseURL = baseURL.JoinPath(websocketPath)
	return baseURL.String()
}

func generateAPIURL(host, cloudhookURL, remoteUIURL, webhookID string) string {
	switch {
	case cloudhookURL != "":
		return cloudhookURL
	case remoteUIURL != "" && webhookID != "":
		baseURL, _ := url.Parse(remoteUIURL)
		baseURL = baseURL.JoinPath(webHookPath, webhookID)
		return baseURL.String()
	case webhookID != "" && host != "":
		baseURL, _ := url.Parse(host)
		baseURL = baseURL.JoinPath(webHookPath, webhookID)
		return baseURL.String()
	default:
		return ""
	}
}

type pref struct {
	fyne  string
	viper string
}

var (
	prefs = map[string]pref{
		"PrefAPIURL":       {fyne: "ApiURL", viper: "hass.apiurl"},
		"PrefWebsocketURL": {fyne: "WebSocketURL", viper: "hass.websocketurl"},
		"PrefCloudhookURL": {fyne: "CloudhookURL", viper: "hass.cloudhookurl"},
		"PrefRemoteUIURL":  {fyne: "RemoteUIURL", viper: "hass.remoteuiurl"},
		"PrefToken":        {fyne: "Token", viper: "hass.token"},
		"PrefWebhookID":    {fyne: "WebhookID", viper: "hass.webhookid"},
		"PrefSecret":       {fyne: "secret", viper: "hass.secret"},
		"PrefHost":         {fyne: "Host", viper: "hass.host"},
		"PrefVersion":      {fyne: "Version", viper: "agent.version"},
		"PrefDeviceName":   {fyne: "DeviceName", viper: "device.name"},
		"PrefDeviceID":     {fyne: "DeviceID", viper: "device.id"},
	}
)

func viperToFyne(configPath string) error {
	var err error
	fs, err := os.Stat(filepath.Join(configPath, "go-hass-agent.toml"))
	if fs != nil && err == nil {
		log.Debug().Msg("Config already migrated. Not doing anything.")
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return errors.Join(errors.New("filesystem error"), err)
	}

	vc, err := viperconfig.New(configPath)
	if err != nil {
		return errors.New("could not open viper config")
	}

	fc := fyneconfig.NewFyneConfig()

	for _, m := range prefs {
		var err error
		var value string
		log.Debug().
			Str("from", m.fyne).Str("to", m.viper).
			Msg("Migrating preference.")
		if err = fc.Get(m.fyne, &value); err != nil && value != "NOTSET" {
			return errors.Join(errors.New("fyne config error"), err)
		}
		if value != "NOTSET" {
			if err = vc.Set(m.viper, value); err != nil {
				return errors.Join(errors.New("viper config error"), err)
			}
		}
	}
	return vc.Set("hass.registered", true)
}
