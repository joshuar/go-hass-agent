// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	RegistrationPath = "/api/mobile_app/registrations"
	WebsocketPath    = "/api/websocket"
	WebHookPath      = "/api/webhook/"
)

var (
	ErrInternalValidationFailed = errors.New("internal validation error")
	ErrDeviceRegistrationFailed = errors.New("device registration failed")
)

type deviceRegistrationRequest struct {
	*preferences.Device
	Token string `json:"-"`
}

func (r *deviceRegistrationRequest) Auth() string {
	return r.Token
}

func (r *deviceRegistrationRequest) RequestBody() any {
	return r.Device
}

// revive:disable:unused-receiver
func (r *deviceRegistrationRequest) Retry() bool {
	return true
}

type deviceRegistrationResponse struct {
	CloudhookURL string `json:"cloudhook_url"`
	RemoteUIURL  string `json:"remote_ui_url"`
	Secret       string `json:"secret"`
	WebhookID    string `json:"webhook_id"`
}

func newRegistrationRequest(thisDevice *preferences.Device, token string) *deviceRegistrationRequest {
	return &deviceRegistrationRequest{
		Device: thisDevice,
		Token:  token,
	}
}

func RegisterDevice(ctx context.Context, registration *preferences.Registration) error {
	// Validate provided registration details.
	if err := registration.Validate(); err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}

	registrationURL := registration.Server + RegistrationPath
	dev := preferences.NewDevice()

	// Register the device against the registration endpoint.
	response, err := api.Send[deviceRegistrationResponse](ctx, registrationURL, newRegistrationRequest(dev, registration.Token))
	if err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}

	// Generate a rest API URL.
	restAPIURL, err := generateAPIURL(&response, registration)
	if err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}
	// Generate a websocket API URL.
	websocketAPIURL, err := generateWebsocketURL(registration.Server)
	if err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}

	// Set registration status in preferences.
	preferences.SetPreferences(
		preferences.SetHassSecret(response.Secret),
		preferences.SetRestAPIURL(restAPIURL),
		preferences.SetWebsocketURL(websocketAPIURL),
		preferences.SetWebhookID(response.WebhookID),
		preferences.SetRegistered(true),
	)
	// Save preferences to disk.
	if err := preferences.Save(); err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}

	return nil
}

// generateAPIURL creates a URL to use for sending data back to Home
// Assistant from the registration information returned by Home Assistant. It
// follows the rules mentioned in the developer docs to generate the URL:
//
// https://developers.home-assistant.io/docs/api/native-app-integration/sending-data#sending-webhook-data-via-rest-api
func generateAPIURL(response *deviceRegistrationResponse, request *preferences.Registration) (string, error) {
	switch {
	case response.CloudhookURL != "" && !request.IgnoreHassURLs:
		return response.CloudhookURL, nil
	case response.RemoteUIURL != "" && response.WebhookID != "" && !request.IgnoreHassURLs:
		return response.RemoteUIURL + WebHookPath + response.WebhookID, nil
	default:
		apiURL, err := url.Parse(request.Server)
		if err != nil {
			return "", fmt.Errorf("could not parse registration server: %w", err)
		}

		return apiURL.JoinPath(WebHookPath, response.WebhookID).String(), nil
	}
}

// generateWebsocketURL creates a URL for the websocket connection. There is a
// specific format and scheme:
//
// https://developers.home-assistant.io/docs/api/websocket
func generateWebsocketURL(server string) (string, error) {
	websocketURL, err := url.Parse(server)
	if err != nil {
		return "", fmt.Errorf("could not parse registration server: %w", err)
	}

	switch websocketURL.Scheme {
	case "https":
		websocketURL.Scheme = "wss"
	case "http":
		websocketURL.Scheme = "ws"
	case "wss":
	default:
		websocketURL.Scheme = "ws"
	}

	return websocketURL.JoinPath(WebsocketPath).String(), nil
}
