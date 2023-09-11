// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/tracker/registry"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
)

//go:generate moq -out mock_agentConfig_test.go . AgentConfig
type AgentConfig interface {
	Get(string, interface{}) error
	Set(string, interface{}) error
	Delete(string) error
	StoragePath(string) (string, error)
}

// ValidateConfig takes an agentConfig and ensures that it meets the minimum
// requirements for the agent to function correctly
func ValidateConfig(c AgentConfig) error {
	validator := validator.New()

	validate := func(key, rules, errMsg string) error {
		var value string
		err := c.Get(key, &value)
		if err != nil {
			return fmt.Errorf("unable to retrieve %s from config: %v", key, err)
		}
		err = validator.Var(value, rules)
		if err != nil {
			return errors.New(errMsg)
		}
		return nil
	}

	if err := validate(config.PrefApiURL,
		"required,url",
		"apiURL does not match either a URL, hostname or hostname:port",
	); err != nil {
		return err
	}
	if err := validate(config.PrefWebsocketURL,
		"required,url",
		"websocketURL does not match either a URL, hostname or hostname:port",
	); err != nil {
		return err
	}
	if err := validate(config.PrefToken,
		"required,ascii",
		"invalid long-lived token format",
	); err != nil {
		return err
	}
	if err := validate(config.PrefWebhookID,
		"required,ascii",
		"invalid webhookID format",
	); err != nil {
		return err
	}

	return nil
}

// UpgradeConfig takes an agentConfig checking and performing various fixes and
// changes to the agent config as it has evolved in difference versions
func UpgradeConfig(c AgentConfig) error {
	var configVersion string
	if err := c.Get(config.PrefVersion, &configVersion); err != nil {
		return fmt.Errorf("config version is not a valid value (%v)", err)
	}

	switch {
	// * Upgrade host to include scheme for versions < v.1.4.0
	case semver.Compare(configVersion, "v1.4.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v1.4.0")
		var hostString string
		if err := c.Get(config.PrefHost, &hostString); err != nil {
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
		if err := c.Set(config.PrefHost, hostString); err != nil {
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
	case semver.Compare(Version, "v3.0.0") < 0:
		log.Debug().Msg("Performing config upgrades for < v3.0.0.")
		var err error
		path, err := c.StoragePath("sensorRegistry")
		if err != nil {
			return errors.New("could not get sensor registry path from config")
		}
		if _, err := os.Stat(path + "/0.dat"); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		err = registry.MigrateNuts2Json(path)
		if err != nil {
			return errors.New("failed to migrate sensor registry")
		}
		if err = os.Remove(path + "/0.dat"); err != nil {
			return errors.New("could not remove old sensor registry")
		}
	}

	if err := c.Set(config.PrefVersion, Version); err != nil {
		return err
	}

	// ! https://github.com/fyne-io/fyne/issues/3170
	time.Sleep(110 * time.Millisecond)

	return nil
}

func generateWebsocketURL(c AgentConfig) error {
	// TODO: look into websocket http upgrade method
	var host string
	if err := c.Get(config.PrefHost, &host); err != nil {
		return err
	}
	url, _ := url.Parse(host)
	switch url.Scheme {
	case "https":
		url.Scheme = "wss"
	case "http":
		fallthrough
	default:
		url.Scheme = "ws"
	}
	url = url.JoinPath(websocketPath)
	return c.Set(config.PrefWebsocketURL, url.String())
}

func generateAPIURL(c AgentConfig) error {
	var cloudhookURL, remoteUIURL, webhookID, host string
	if err := c.Get(config.PrefCloudhookURL, &cloudhookURL); err != nil {
		return err
	}
	if err := c.Get(config.PrefRemoteUIURL, &remoteUIURL); err != nil {
		return err
	}
	if err := c.Get(config.PrefWebhookID, &webhookID); err != nil {
		return err
	}
	if err := c.Get(config.PrefHost, &host); err != nil {
		return err
	}
	var apiURL string
	switch {
	case cloudhookURL != "":
		apiURL = cloudhookURL
	case remoteUIURL != "" && webhookID != "":
		apiURL = remoteUIURL + webHookPath + webhookID
	case webhookID != "" && host != "":
		url, _ := url.Parse(host)
		url = url.JoinPath(webHookPath, webhookID)
		apiURL = url.String()
	default:
		apiURL = ""
	}
	return c.Set(config.PrefApiURL, apiURL)
}
