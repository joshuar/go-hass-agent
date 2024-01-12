// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/rs/zerolog/log"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
)

//go:generate moq -out mock_RegistrationInfo_test.go . RegistrationInfo
type RegistrationInfo interface {
	Server() string
	Token() string
}

type RegistrationResponse struct {
	CloudhookURL string `json:"cloudhook_url"`
	RemoteUIURL  string `json:"remote_ui_url"`
	Secret       string `json:"secret"`
	WebhookID    string `json:"webhook_id"`
}

func (rr *RegistrationResponse) GenerateAPIURL(host string) string {
	switch {
	case rr.CloudhookURL != "":
		return rr.CloudhookURL
	case rr.RemoteUIURL != "" && rr.WebhookID != "":
		return rr.RemoteUIURL + webHookPath + rr.WebhookID
	case rr.WebhookID != "":
		u, _ := url.Parse(host)
		u = u.JoinPath(webHookPath, rr.WebhookID)
		return u.String()
	default:
		return ""
	}
}

func (rr *RegistrationResponse) GenerateWebsocketURL(host string) string {
	// TODO: look into websocket http upgrade method
	u, err := url.Parse(host)
	if err != nil {
		log.Warn().Err(err).Msg("Could not parse URL.")
		return ""
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	case "ws":
		// nothing to do
	case "wss":
		// nothing to do
	default:
		u.Scheme = "ws"
	}
	u = u.JoinPath(websocketPath)
	return u.String()
}

type RegistrationRequest struct {
	AppData            any    `json:"app_data,omitempty"`
	DeviceID           string `json:"device_id"`
	AppID              string `json:"app_id"`
	AppName            string `json:"app_name"`
	AppVersion         string `json:"app_version"`
	DeviceName         string `json:"device_name"`
	Manufacturer       string `json:"manufacturer"`
	Model              string `json:"model"`
	OsName             string `json:"os_name"`
	OsVersion          string `json:"os_version"`
	SupportsEncryption bool   `json:"supports_encryption"`
}

func RegisterWithHass(ctx context.Context, cfg config.Config, device DeviceInfo) (*RegistrationResponse, error) {
	request, err := device.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var u string
	if err = cfg.Get(config.PrefHost, &u); err != nil {
		return nil, errors.New("invalid host")
	}

	serverURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	serverURL = serverURL.JoinPath("/api/mobile_app/registrations")

	var token string
	if err = cfg.Get(config.PrefToken, &token); err != nil || token == "" {
		return nil, errors.New("invalid token")
	}

	var response *RegistrationResponse
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	err = requests.
		URL(serverURL.String()).
		Header("Authorization", "Bearer "+token).
		BodyBytes(request).
		ToJSON(&response).
		Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return response, nil
}
