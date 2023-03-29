package agent

import "github.com/rs/zerolog/log"

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

func (agent *Agent) LoadConfig() {
	// go func() {
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
			// configLoaded <- true
			return
		case RemoteUIURL != "" && agent.config.webhookID != "":
			agent.config.APIURL = RemoteUIURL + webHookPath + agent.config.webhookID
			log.Debug().Caller().
				Msgf("Using RemoteUIURL %s for Home Assistant access", agent.config.APIURL)
			// configLoaded <- true
			return
		case agent.config.webhookID != "" && Host != "":
			if UseTLS {
				agent.config.APIURL = "https://" + Host + webHookPath + agent.config.webhookID
			} else {
				agent.config.APIURL = "http://" + Host + webHookPath + agent.config.webhookID
			}
			log.Debug().Caller().
				Msgf("Using generated URL %s for Home Assistant access", agent.config.APIURL)
			// configLoaded <- true
			return
		default:
			log.Warn().Msg("No suitable existing config found! Starting new registration process")
			agent.runRegistrationWorker()
		}
	}
	// }()
}

func (agent *Agent) GetConfigVersion() string {
	return agent.App.Preferences().String("Version")
}
