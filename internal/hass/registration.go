// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:errname // structs are dual-purpose response and error
package hass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	registrationPath = "/api/mobile_app/registrations"
	WebsocketPath    = "/api/websocket"
	WebHookPath      = "/api/webhook/"
)

var ErrInternalValidationFailed = errors.New("internal validation error")

type DeviceInfo struct {
	DeviceID           string  `json:"device_id"`
	AppID              string  `json:"app_id"`
	AppName            string  `json:"app_name"`
	AppVersion         string  `json:"app_version"`
	DeviceName         string  `json:"device_name"`
	Manufacturer       string  `json:"manufacturer"`
	Model              string  `json:"model"`
	OsName             string  `json:"os_name"`
	OsVersion          string  `json:"os_version"`
	AppData            AppData `json:"app_data,omitempty"`
	SupportsEncryption bool    `json:"supports_encryption"`
}

type AppData struct {
	PushWebsocket bool `json:"push_websocket_channel"`
}

type RegistrationInput struct {
	Server           string `validate:"required,http_url"`
	Token            string `validate:"required"`
	IgnoreOutputURLs bool   `validate:"boolean"`
}

func (i *RegistrationInput) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(i)
	if err != nil {
		return showValidationErrors(err)
	}

	return nil
}

type RegistrationDetails struct {
	CloudhookURL string `json:"cloudhook_url"`
	RemoteUIURL  string `json:"remote_ui_url"`
	Secret       string `json:"secret"`
	WebhookID    string `json:"webhook_id"`
}

type registrationRequest struct {
	*DeviceInfo
	Token string `json:"-"`
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

func newRegistrationRequest(info *DeviceInfo, token string) *registrationRequest {
	return &registrationRequest{
		DeviceInfo: info,
		Token:      token,
	}
}

type registrationResponse struct {
	Details *RegistrationDetails
	*APIError
}

func (r *registrationResponse) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &r.Details)
	if err != nil {
		return fmt.Errorf("failed to parse registration response: %w", err)
	}

	return nil
}

func (r *registrationResponse) UnmarshalError(data []byte) error {
	err := json.Unmarshal(data, r.APIError)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func (r *registrationResponse) Error() string {
	return r.APIError.Error()
}

//nolint:exhaustruct
func newRegistrationResponse() *registrationResponse {
	return &registrationResponse{}
}

func RegisterWithHass(ctx context.Context, input *RegistrationInput, deviceInfo *DeviceInfo) (*RegistrationDetails, error) {
	req := newRegistrationRequest(deviceInfo, input.Token)
	resp := newRegistrationResponse()

	serverURL, err := url.Parse(input.Server)
	if err != nil {
		return nil, fmt.Errorf("could not parse server URL: %w", err)
	}

	client := NewDefaultHTTPClient(serverURL.String()).SetTimeout(time.Minute)

	if err := ExecuteRequest(ctx, client, registrationPath, req, resp); err != nil {
		return nil, err
	}

	return resp.Details, nil
}

//nolint:err113,errorlint
func showValidationErrors(e error) error {
	validationErrors, ok := e.(validator.ValidationErrors)
	if !ok {
		return ErrInternalValidationFailed
	}

	var allErrors error

	for _, err := range validationErrors {
		allErrors = errors.Join(allErrors, fmt.Errorf("could validate %s input: got %s, want %s", err.Field(), err.Value(), err.Tag()))
	}

	return allErrors
}
