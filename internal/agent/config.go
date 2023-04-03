package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
)

func (agent *Agent) loadConfig(ctx context.Context) *config.AppConfig {
	for {
		CloudhookURL := agent.App.Preferences().String("CloudhookURL")
		RemoteUIURL := agent.App.Preferences().String("RemoteUIURL")
		Host := agent.App.Preferences().String("Host")
		UseTLS := agent.App.Preferences().Bool("UseTLS")

		config := &config.AppConfig{}
		config.Secret = agent.App.Preferences().String("Secret")
		config.Token = agent.App.Preferences().String("Token")
		config.WebhookID = agent.App.Preferences().String("WebhookID")

		if UseTLS {
			config.WebSocketURL = "wss://" + Host + websocketPath
		} else {
			config.WebSocketURL = "ws://" + Host + websocketPath
		}

		switch {
		case CloudhookURL != "":
			config.APIURL = CloudhookURL
			log.Debug().Caller().
				Msgf("Using CloudhookURL %s for Home Assistant access", config.APIURL)
			return config
		case RemoteUIURL != "" && config.WebhookID != "":
			config.APIURL = RemoteUIURL + webHookPath + config.WebhookID
			log.Debug().Caller().
				Msgf("Using RemoteUIURL %s for Home Assistant access", config.APIURL)
			return config
		case config.WebhookID != "" && Host != "":
			if UseTLS {
				config.APIURL = "https://" + Host + webHookPath + config.WebhookID
			} else {
				config.APIURL = "http://" + Host + webHookPath + config.WebhookID
			}
			log.Debug().Caller().
				Msgf("Using generated URL %s for Home Assistant access", config.APIURL)
			return config
		default:
			log.Warn().Msg("No suitable existing config found! Starting new registration process")
			err := agent.runRegistrationWorker(ctx)
			if err != nil {
				log.Debug().Caller().
					Msgf("Error trying to register: %v. Exiting.", err)
				agent.Exit()
			}
		}
	}
}

func (agent *Agent) GetConfigVersion() string {
	return agent.App.Preferences().String("Version")
}

func (agent *Agent) GetDeviceDetails() (string, string) {
	return agent.App.Preferences().String("DeviceName"),
		agent.App.Preferences().String("DeviceID")
}

func (agent *Agent) saveRegistration(r *hass.RegistrationResponse, h *hass.RegistrationHost) {
	host, _ := h.Server.Get()
	useTLS, _ := h.UseTLS.Get()
	agent.App.Preferences().SetString("Host", host)
	agent.App.Preferences().SetBool("UseTLS", useTLS)
	token, _ := h.Token.Get()
	agent.App.Preferences().SetString("Token", token)
	agent.App.Preferences().SetString("Version", agent.Version)
	if r.CloudhookURL != "" {
		agent.App.Preferences().SetString("CloudhookURL", r.CloudhookURL)
	}
	if r.RemoteUIURL != "" {
		agent.App.Preferences().SetString("RemoteUIURL", r.RemoteUIURL)
	}
	if r.Secret != "" {
		agent.App.Preferences().SetString("Secret", r.Secret)
	}
	if r.WebhookID != "" {
		agent.App.Preferences().SetString("WebhookID", r.WebhookID)
	}
}
