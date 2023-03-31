package agent

import (
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
)

type AppConfig struct {
	APIURL       string `json:"restapi_url"`
	WebSocketURL string `json:"instance_url"`
	secret       string
	token        string
	webhookID    string
}

func (agent *Agent) loadConfig() {
	for {
		CloudhookURL := agent.App.Preferences().String("CloudhookURL")
		RemoteUIURL := agent.App.Preferences().String("RemoteUIURL")
		Host := agent.App.Preferences().String("Host")
		UseTLS := agent.App.Preferences().Bool("UseTLS")

		agent.config.secret = agent.App.Preferences().String("Secret")
		agent.config.token = agent.App.Preferences().String("Token")
		agent.config.webhookID = agent.App.Preferences().String("WebhookID")

		if UseTLS {
			agent.config.WebSocketURL = "wss://" + Host + websocketPath
		} else {
			agent.config.WebSocketURL = "ws://" + Host + websocketPath
		}

		switch {
		case CloudhookURL != "":
			agent.config.APIURL = CloudhookURL
			log.Debug().Caller().
				Msgf("Using CloudhookURL %s for Home Assistant access", agent.config.APIURL)
			return
		case RemoteUIURL != "" && agent.config.webhookID != "":
			agent.config.APIURL = RemoteUIURL + webHookPath + agent.config.webhookID
			log.Debug().Caller().
				Msgf("Using RemoteUIURL %s for Home Assistant access", agent.config.APIURL)
			return
		case agent.config.webhookID != "" && Host != "":
			if UseTLS {
				agent.config.APIURL = "https://" + Host + webHookPath + agent.config.webhookID
			} else {
				agent.config.APIURL = "http://" + Host + webHookPath + agent.config.webhookID
			}
			log.Debug().Caller().
				Msgf("Using generated URL %s for Home Assistant access", agent.config.APIURL)
			return
		default:
			log.Warn().Msg("No suitable existing config found! Starting new registration process")
			err := agent.runRegistrationWorker()
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
