// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/carlmjohnson/requests"
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
	u, _ := url.Parse(host)
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
	AppData            interface{} `json:"app_data,omitempty"`
	DeviceID           string      `json:"device_id"`
	AppID              string      `json:"app_id"`
	AppName            string      `json:"app_name"`
	AppVersion         string      `json:"app_version"`
	DeviceName         string      `json:"device_name"`
	Manufacturer       string      `json:"manufacturer"`
	Model              string      `json:"model"`
	OsName             string      `json:"os_name"`
	OsVersion          string      `json:"os_version"`
	SupportsEncryption bool        `json:"supports_encryption"`
}

func RegisterWithHass(ctx context.Context, registration RegistrationInfo, device DeviceInfo) (*RegistrationResponse, error) {
	request, err := device.MarshalJSON()
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(registration.Server())
	if err != nil {
		return nil, err
	}
	url = url.JoinPath("/api/mobile_app/registrations")

	if registration.Token() == "" {
		return nil, errors.New("invalid token")
	}

	var response *RegistrationResponse
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	err = requests.
		URL(url.String()).
		Header("Authorization", "Bearer "+registration.Token()).
		BodyBytes(request).
		ToJSON(&response).
		Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return response, nil
}
