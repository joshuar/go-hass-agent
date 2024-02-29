// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
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

type RegistrationInput struct {
	Server           string `validate:"required,http_url"`
	Token            string `validate:"required"`
	IgnoreOutputURLs bool   `validate:"boolean"`
}

func (i *RegistrationInput) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(i)
}

type RegistrationDetails struct {
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

type registrationResponse struct {
	Details *RegistrationDetails
	err     error
}

func (r *registrationResponse) StoreError(err error) {
	r.err = err
}

func (r *registrationResponse) Error() string {
	return r.err.Error()
}

func (r *registrationResponse) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &r.Details)
}

func newRegistrationResponse() *registrationResponse {
	return &registrationResponse{
		err: errors.New(""),
	}
}

func RegisterWithHass(ctx context.Context, input *RegistrationInput, device DeviceInfo) (*RegistrationDetails, error) {
	req := newRegistrationRequest(device, input.Token)
	resp := newRegistrationResponse()

	serverURL, err := url.Parse(input.Server)
	if err != nil {
		return nil, err
	}
	serverURL = serverURL.JoinPath(registrationPath)
	ctx = ContextSetURL(ctx, serverURL.String())
	ctx = ContextSetClient(ctx, NewDefaultHTTPClient().SetTimeout(time.Minute))

	ExecuteRequest(ctx, req, resp)
	if errors.Is(resp, &APIError{}) || resp.Error() != "" {
		return nil, resp
	}
	return resp.Details, nil
}
