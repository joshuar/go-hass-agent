package agent

import "github.com/rs/zerolog/log"

type AppConfig struct {
	RestAPIURL string `json:"restapi_url"`
	secret     string
	token      string
}

func (a *Agent) LoadConfig() {
	// go func() {
	for {
		CloudhookURL := a.App.Preferences().String("CloudhookURL")
		RemoteUIURL := a.App.Preferences().String("RemoteUIURL")
		WebhookID := a.App.Preferences().String("WebhookID")
		InstanceURL := a.App.Preferences().String("InstanceURL")

		a.config.secret = a.App.Preferences().String("Secret")
		a.config.token = a.App.Preferences().String("Token")

		switch {
		case CloudhookURL != "":
			a.config.RestAPIURL = CloudhookURL
			log.Debug().Caller().
				Msgf("Using CloudhookURL %s for Home Assistant access", a.config.RestAPIURL)
			a.configLoaded <- true
			return
		case RemoteUIURL != "" && WebhookID != "":
			a.config.RestAPIURL = RemoteUIURL + "/api/webhook/" + WebhookID
			log.Debug().Caller().
				Msgf("Using RemoteUIURL %s for Home Assistant access", a.config.RestAPIURL)
			a.configLoaded <- true
			return
		case WebhookID != "" && InstanceURL != "":
			a.config.RestAPIURL = InstanceURL + "/api/webhook/" + WebhookID
			log.Debug().Caller().
				Msgf("Using InstanceURL %s for Home Assistant access", a.config.RestAPIURL)
			a.configLoaded <- true
			return
		default:
			log.Warn().Msg("No suitable existing config found! Starting new registration process")
			a.runRegistrationWorker()
		}
	}
	// }()
}

func (a *Agent) GetConfigVersion() string {
	return a.App.Preferences().String("Version")
}
