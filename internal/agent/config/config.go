// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"

	fyneconfig "github.com/joshuar/go-hass-agent/internal/agent/config/fyneConfig"
	viperconfig "github.com/joshuar/go-hass-agent/internal/agent/config/viperConfig"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
	AppName       = "go-hass-agent"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var AppVersion string

// Config represents the methods that the agent uses to interact with
// its config. It is effectively a CRUD interface to wherever the configuration
// is stored.
//
//go:generate moq -out mockConfig_test.go . Config
type Config interface {
	Get(key string, value any) error
	Set(key string, value any) error
	Delete(key string) error
	Path() string
	StoragePath(path string) (string, error)
}

func Load(configPath string) (Config, error) {
	var cfg Config
	var err error
	if cfg, err = viperconfig.New(configPath); err != nil {
		return nil, err
	}
	var registered bool
	err = cfg.Get(PrefRegistered, &registered)
	if err != nil {
		log.Debug().Err(err).Msg("Registration status not found.")
	}
	if registered {
		if err = UpgradeConfig(cfg); err != nil {
			if _, ok := err.(*FileNotFoundError); !ok {
				return nil, errors.New("could not upgrade config")
			}
		}
		if err = ValidateConfig(cfg); err != nil {
			return nil, errors.New("could not validate config")
		}
	}
	return cfg, nil
}

type FileNotFoundError struct {
	error
}

func (e FileNotFoundError) Unwrap() error {
	return e.error
}

type InvalidFormatError struct {
	error
}

func (e InvalidFormatError) Unwrap() error {
	return e.error
}

type UpgradeError struct {
	error
}

func (e UpgradeError) Unwrap() error {
	return e.error
}

// ValidateConfig takes an AgentConfig and ensures that it meets the minimum
// requirements for the agent to function correctly.
func ValidateConfig(c Config) error {
	log.Debug().Msg("Validating config.")
	cfgValidator := validator.New()

	validate := func(key, rules, errMsg string) error {
		var value string
		err := c.Get(key, &value)
		if err != nil {
			return &InvalidFormatError{error: fmt.Errorf("unable to retrieve %s from config: %v", key, err)}
		}
		err = cfgValidator.Var(value, rules)
		if err != nil {
			return &InvalidFormatError{error: errors.New(errMsg)}
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
func UpgradeConfig(vc Config) error {
	log.Debug().Msg("Checking for config upgrades.")
	var configVersion string
	// retrieve the configVersion, or the version of the app that last read/validated the config.
	if semver.Compare(AppVersion, "v5.0.0") < 0 {
		fc := fyneconfig.NewFyneConfig()
		if err := fc.Get("Version", &configVersion); err != nil {
			return &FileNotFoundError{error: err}
		}
	} else {
		if err := vc.Get("Version", &configVersion); err != nil {
			return &FileNotFoundError{error: err}
		}
	}

	// depending on the configVersion, do the appropriate upgrades. Note that
	// some switch statements will need to fallthrough as some require previous
	// upgrades to have happened. No doubt at some point, this becomes
	// intractable and the upgrade path will need to be truncated at some
	// previous version.
	log.Debug().Msgf("Checking for upgrades needed for config version %s.", configVersion)
	switch {
	// * Minimum upgradeable version.
	case semver.Compare(configVersion, "v3.0.0") < 0:
		return &UpgradeError{
			error: errors.New("cannot upgrade versions < v3.0.0. Please remove the config directory and start fresh to continue"),
		}
	// * Switch to Viper config.
	case semver.Compare(configVersion, "v5.0.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v5.0.0.")
		// migrate config values
		if err := viperToFyne(vc.Path()); err != nil {
			return &UpgradeError{
				error: errors.Join(errors.New("failed to migrate Fyne config to Viper"), err),
			}
		}
		// migrate registry directory. This is non-critical, entities will be
		// re-registered if this fails.
		fc := fyneconfig.NewFyneConfig()
		oldReg, err := fc.StoragePath("sensorRegistry")
		newReg := filepath.Join(vc.Path(), "sensorRegistry")
		if err != nil {
			log.Warn().Err(err).Msg("Unable to retrieve old storage path. Registry will not be migrated.")
			return nil
		}
		_, err = os.Stat(oldReg)
		if !os.IsNotExist(err) {
			if err := os.Rename(oldReg, newReg); err != nil {
				log.Warn().Err(err).Msg("failed to migrate registry")
				return nil
			}
		}
		fallthrough
	default:
		if err := vc.Set(PrefVersion, AppVersion); err != nil {
			log.Warn().Err(err).Msg("Unable to set config version to app version.")
		}
	}
	return nil
}

type pref struct {
	fyne  string
	viper string
}

func viperToFyne(configPath string) error {
	prefs := map[string]pref{
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
