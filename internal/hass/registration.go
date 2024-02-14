// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"net/url"
	"time"
)

const (
	registrationPath = "/api/mobile_app/registrations"
	WebsocketPath    = "/api/websocket"
	WebHookPath      = "/api/webhook/"
)

//go:generate moq -out mock_DeviceInfo_test.go . DeviceInfo
type DeviceInfo interface {
	DeviceID() string
	AppID() string
	AppName() string
	AppVersion() string
	DeviceName() string
	Manufacturer() string
	Model() string
	OsName() string
	OsVersion() string
	SupportsEncryption() bool
	AppData() any
}

type RegistrationResponse struct {
	CloudhookURL string `json:"cloudhook_url"`
	RemoteUIURL  string `json:"remote_ui_url"`
	Secret       string `json:"secret"`
	WebhookID    string `json:"webhook_id"`
}

type registrationRequest struct {
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
	Token              string `json:"-"`
	SupportsEncryption bool   `json:"supports_encryption"`
}

func (r *registrationRequest) Auth() string {
	return r.Token
}

func (r *registrationRequest) RequestBody() json.RawMessage {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return data
}

func (r *registrationRequest) ResponseBody() any {
	return &RegistrationResponse{}
}

func newRegistrationRequest(d DeviceInfo, t string) *registrationRequest {
	return &registrationRequest{
		DeviceID:           d.DeviceID(),
		AppID:              d.AppID(),
		AppName:            d.AppName(),
		AppVersion:         d.AppVersion(),
		DeviceName:         d.DeviceName(),
		Manufacturer:       d.Manufacturer(),
		Model:              d.Model(),
		OsName:             d.OsName(),
		OsVersion:          d.OsVersion(),
		SupportsEncryption: d.SupportsEncryption(),
		AppData:            d.AppData(),
		Token:              t,
	}
}

func RegisterWithHass(ctx context.Context, server, token string, device DeviceInfo) (*RegistrationResponse, error) {
	req := newRegistrationRequest(device, token)

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	serverURL = serverURL.JoinPath(registrationPath)
	ctx = ContextSetURL(ctx, serverURL.String())

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	resp := <-ExecuteRequest(ctx, req)
	if resp.Error != nil {
		return nil, resp.Error
	}
	var details *RegistrationResponse
	var ok bool
	if details, ok = resp.Body.(*RegistrationResponse); !ok {
		return nil, ErrResponseMalformed
	}
	return details, nil
}
