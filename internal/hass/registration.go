// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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

//nolint:interfacebloat
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

func newRegistrationRequest(info DeviceInfo, token string) *registrationRequest {
	return &registrationRequest{
		DeviceID:           info.DeviceID(),
		AppID:              info.AppID(),
		AppName:            info.AppName(),
		AppVersion:         info.AppVersion(),
		DeviceName:         info.DeviceName(),
		Manufacturer:       info.Manufacturer(),
		Model:              info.Model(),
		OsName:             info.OsName(),
		OsVersion:          info.OsVersion(),
		SupportsEncryption: info.SupportsEncryption(),
		AppData:            info.AppData(),
		Token:              token,
	}
}

type registrationResponse struct {
	Details *RegistrationDetails
}

func (r *registrationResponse) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &r.Details)
	if err != nil {
		return fmt.Errorf("failed to parse registration response: %w", err)
	}

	return nil
}

//nolint:exhaustruct
func newRegistrationResponse() *registrationResponse {
	return &registrationResponse{}
}

func RegisterWithHass(ctx context.Context, input *RegistrationInput, device DeviceInfo) (*RegistrationDetails, error) {
	req := newRegistrationRequest(device, input.Token)
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
