// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"context"
	"net/url"
	"time"

	"github.com/carlmjohnson/requests"
)

const (
	WebsocketPath    = "/api/websocket"
	WebHookPath      = "/api/webhook/"
	registrationPath = "/api/mobile_app/registrations"
	authHeader       = "Authorization"
)

type RegistrationResponse struct {
	CloudhookURL string `json:"cloudhook_url"`
	RemoteUIURL  string `json:"remote_ui_url"`
	Secret       string `json:"secret"`
	WebhookID    string `json:"webhook_id"`
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

func RegisterWithHass(ctx context.Context, server, token string, device DeviceInfo) (*RegistrationResponse, error) {
	request, err := device.MarshalJSON()
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	serverURL = serverURL.JoinPath(registrationPath)

	var response *RegistrationResponse
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	err = requests.
		URL(serverURL.String()).
		Header(authHeader, "Bearer "+token).
		BodyBytes(request).
		ToJSON(&response).
		Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return response, nil
}
