// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package viperconfig

import (
	"errors"
	"os"
	"path/filepath"

	fyneconfig "github.com/joshuar/go-hass-agent/internal/agent/config/fyneConfig"
	"github.com/rs/zerolog/log"
)

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
	configPath = filepath.Join(os.Getenv("HOME"), ".config", "go-hass-agent")
)

func MigrateFromFyne() error {
	var err error
	fs, err := os.Stat(filepath.Join(configPath, "go-hass-agent.toml"))
	if fs != nil && err == nil {
		log.Debug().Msg("Config already migrated. Not doing anything.")
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return errors.Join(errors.New("filesystem error"), err)
	}

	viperConfig, err := New(configPath)
	if err != nil {
		return errors.New("could not open viper config")
	}
	fyneConfig := fyneconfig.NewFyneConfig()

	for _, m := range prefs {
		var err error
		var value string
		log.Debug().
			Str("from", m.fyne).Str("to", m.viper).
			Msg("Migrating preference.")
		if err = fyneConfig.Get(m.fyne, &value); err != nil && value != "NOTSET" {
			return errors.Join(errors.New("fyne config error"), err)
		}
		if value != "NOTSET" {
			if err = viperConfig.Set(m.viper, value); err != nil {
				return errors.Join(errors.New("viper config error"), err)
			}
		}
	}
	return viperConfig.Set("hass.registered", true)
}
