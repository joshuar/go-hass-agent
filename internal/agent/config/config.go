// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	_ "embed"

	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/go-playground/validator/v10"
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

// UpgradeConfig takes an AgentConfig checking and performing various fixes and
// changes to the agent config as it has evolved in difference versions
func UpgradeConfig(c AgentConfig) error {
	var configVersion string
	if err := c.Get(PrefVersion, &configVersion); err != nil {
		return fmt.Errorf("config version is not a valid value (%v)", err)
	}

	switch {
	// * Upgrade host to include scheme for versions < v.1.4.0
	case semver.Compare(configVersion, "v1.4.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v1.4.0")
		var hostString string
		if err := c.Get(PrefHost, &hostString); err != nil {
			return fmt.Errorf("upgrade < v.1.4.0: invalid host value (%v)", err)
		}
		var tlsBool bool
		if err := c.Get("UseTLS", &tlsBool); err != nil {
			return fmt.Errorf("upgrade < v.1.4.0: invalid TLS value (%v)", err)
		}
		switch tlsBool {
		case true:
			hostString = "https://" + hostString
		case false:
			hostString = "http://" + hostString
		}
		if err := c.Set(PrefHost, hostString); err != nil {
			return err
		}
		fallthrough
	// * Add ApiURL and WebSocketURL config options for versions < v1.4.3
	case semver.Compare(configVersion, "v1.4.3") < 0:
		log.Debug().Msg("Performing config upgrades for < v1.4.3")
		if err := generateAPIURL(c); err != nil {
			return err
		}
		if err := generateWebsocketURL(c); err != nil {
			return err
		}
		fallthrough
	case semver.Compare(AppVersion, "v3.0.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v3.0.0.")
		var err error
		path, err := c.StoragePath("sensorRegistry")
		if err != nil {
			return errors.New("could not get sensor registry path from config")
		}
		if _, err = os.Stat(path + "/0.dat"); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		err = registry.MigrateNuts2Json(path)
		if err != nil {
			return errors.New("failed to migrate sensor registry")
		}
		if err = os.Remove(path + "/0.dat"); err != nil {
			return errors.New("could not remove old sensor registry")
		}
		fallthrough
	case semver.Compare(AppVersion, "v5.0.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v5.0.0.")
		if err := viperconfig.MigrateFromFyne(); err != nil {
			return errors.Join(errors.New("failed to migrate Fyne config to Viper"), err)
		}
	}

	if err := c.Set(PrefVersion, AppVersion); err != nil {
		return err
	}
	return nil
}

func generateWebsocketURL(c AgentConfig) error {
	// TODO: look into websocket http upgrade method
	var host string
	if err := c.Get(PrefHost, &host); err != nil {
		return err
	}
	baseURL, _ := url.Parse(host)
	switch baseURL.Scheme {
	case "https":
		baseURL.Scheme = "wss"
	default:
		baseURL.Scheme = "ws"
	}
	baseURL = baseURL.JoinPath(websocketPath)
	return c.Set(PrefWebsocketURL, baseURL.String())
}

func generateAPIURL(c AgentConfig) error {
	var cloudhookURL, remoteUIURL, webhookID, host string
	if err := c.Get(PrefCloudhookURL, &cloudhookURL); err != nil {
		return err
	}
	if err := c.Get(PrefRemoteUIURL, &remoteUIURL); err != nil {
		return err
	}
	if err := c.Get(PrefWebhookID, &webhookID); err != nil {
		return err
	}
	if err := c.Get(PrefHost, &host); err != nil {
		return err
	}
	var apiURL string
	switch {
	case cloudhookURL != "":
		apiURL = cloudhookURL
	case remoteUIURL != "" && webhookID != "":
		apiURL = remoteUIURL + webHookPath + webhookID
	case webhookID != "" && host != "":
		baseURL, _ := url.Parse(host)
		baseURL = baseURL.JoinPath(webHookPath, webhookID)
		apiURL = baseURL.String()
	default:
		apiURL = ""
	}
	return c.Set(PrefAPIURL, apiURL)
}
