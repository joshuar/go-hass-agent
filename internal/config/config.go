package config

import (
	"github.com/kirsle/configdir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	Name    = "go-hass-agent"
	Version = "0.0.1"
)

var cfgFile string

type appConfig struct {
	CloudhookURL string `json:"cloudhook_url"`
	RemoteUIURL  string `json:"remote_ui_url"`
	secret       string
	WebhookID    string `json:"webhook_id"`
	deviceID     string
	appID        string
}

func InitialiseConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(Name)
		viper.SetConfigType("yaml")
		viper.AddConfigPath(configdir.LocalConfig(Name))
	}

	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %s", viper.ConfigFileUsed())
	} else {
		log.Warn(err.Error())
	}
}


