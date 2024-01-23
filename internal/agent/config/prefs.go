// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import "github.com/rs/zerolog/log"

const (
	// PrefAPIURL       = "ApiURL"
	// PrefWebsocketURL = "WebSocketURL"
	// PrefCloudhookURL = "CloudhookURL"
	// PrefRemoteUIURL  = "RemoteUIURL"
	// PrefToken        = "Token"
	// PrefWebhookID    = "WebhookID"
	// PrefSecret       = "secret"
	// PrefVersion      = "Version"
	// PrefHost         = "Host"
	// PrefDeviceName   = "DeviceName"
	// PrefDeviceID     = "DeviceID"
	// PrefRegistered   = "Registered"
	// PrefMQTTServer   = "MQTTServer"
	// PrefMQTTTopic    = "MQTTTopic"
	// PrefMQTTUser     = "MQTTUser"
	// PrefMQTTPassword = "MQTTPassword"
	// PrefMQTTEnabled  = "UseMQTT"

	PrefAPIURL       = "hass.apiurl"
	PrefWebsocketURL = "hass.websocketurl"
	PrefCloudhookURL = "hass.cloudhookurl"
	PrefRemoteUIURL  = "hass.remoteuiurl"
	PrefToken        = "hass.token"
	PrefWebhookID    = "hass.webhookid"
	PrefSecret       = "hass.secret"
	PrefHost         = "hass.host"
	PrefRegistered   = "hass.registered"
	PrefVersion      = "agent.version"
	PrefDeviceName   = "device.name"
	PrefDeviceID     = "device.id"
	PrefMQTTServer   = "mqtt.server"
	PrefMQTTTopic    = "mqtt.topic"
	PrefMQTTUser     = "mqtt.user"
	PrefMQTTPassword = "mqtt.password"
	PrefMQTTEnabled  = "mqtt.enabled"
)

type MQTTPrefs struct {
	Server, User, Password string
	Enabled                bool
}

func LoadMQTTPrefs(cfg Config) *MQTTPrefs {
	s := &MQTTPrefs{
		Server: "localhost:1883",
	}
	if err := cfg.Get(PrefMQTTServer, &s.Server); err != nil {
		log.Warn().Err(err).Str("key", PrefMQTTServer).
			Msg("Could not load setting.")
	}
	if err := cfg.Get(PrefMQTTUser, &s.User); err != nil {
		log.Warn().Err(err).Str("key", PrefMQTTUser).
			Msg("Could not load setting.")
	}
	if err := cfg.Get(PrefMQTTPassword, &s.Password); err != nil {
		log.Warn().Err(err).Str("key", PrefMQTTPassword).
			Msg("Could not load setting.")
	}
	if err := cfg.Get(PrefMQTTEnabled, &s.Enabled); err != nil {
		log.Warn().Err(err).Str("key", PrefMQTTEnabled).
			Msg("Could not load setting.")
	}
	return s
}

func (s *MQTTPrefs) Save(cfg Config) {
	if err := cfg.Set(PrefMQTTServer, &s.Server); err != nil {
		log.Warn().Err(err).Str("key", PrefMQTTServer).
			Msg("Could not save setting.")
	}
	if err := cfg.Set(PrefMQTTUser, &s.User); err != nil {
		log.Warn().Err(err).Str("key", PrefMQTTUser).
			Msg("Could not save setting.")
	}
	if err := cfg.Set(PrefMQTTPassword, &s.Password); err != nil {
		log.Warn().Err(err).Str("key", PrefMQTTPassword).
			Msg("Could not save setting.")
	}
	if err := cfg.Set(PrefMQTTEnabled, &s.Enabled); err != nil {
		log.Warn().Err(err).Str("key", PrefMQTTEnabled).
			Msg("Could not save setting.")
	}
}
