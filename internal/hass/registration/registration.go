// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package registration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/device/info"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
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

// RegisterDevice will register the device running Go Hass Agent with Home
// Assistant, using the details from the preferences passed in.
func RegisterDevice(ctx context.Context, registration *preferences.Registration) error {
	// Validate provided registration details.
	if err := registration.Validate(); err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}

	req := newDeviceRegistration(ctx)
	resp := api.DeviceRegistrationResponse{}

	// Set up the api request, and the request/response bodies.
	apiReq := api.NewClient().R().SetContext(ctx)
	apiReq.SetAuthToken(registration.Token)
	apiReq.SetBody(req)
	apiReq = apiReq.SetResult(&resp)

	registrationURL, err := url.Parse(registration.Server)
	if err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}

	registrationURL = registrationURL.JoinPath(RegistrationPath)

	restyResp, err := apiReq.Post(registrationURL.String())

	switch {
	case err != nil:
		return errors.Join(ErrDeviceRegistrationFailed, err)
	case restyResp.IsError():
		return errors.Join(ErrDeviceRegistrationFailed, errors.New(restyResp.Status()))
	}

	// Generate a rest API URL.
	restAPIURL, err := generateAPIURL(&resp, registration)
	if err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}
	// Generate a websocket API URL.
	websocketAPIURL, err := generateWebsocketURL(registration.Server)
	if err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}

	prefs := []preferences.SetPreference{
		preferences.SetRestAPIURL(restAPIURL),
		preferences.SetWebsocketURL(websocketAPIURL),
		preferences.SetWebhookID(resp.WebhookID),
		preferences.SetServer(registration.Server),
		preferences.SetToken(registration.Token),
		preferences.SetRegistered(true),
	}

	if resp.Secret != nil {
		prefs = append(prefs, preferences.SetHassSecret(*resp.Secret))
	}

	// Set registration status in preferences.
	if err = preferences.Set(prefs...); err != nil {
		return errors.Join(ErrDeviceRegistrationFailed, err)
	}

	return nil
}

// NewDevice creates a new device. This is used during registration with Home
// Assistant to identify the host running Go Hass Agent. While most of the
// information generated is only needed during registration, the device ID and
// Name will be stored in the preferences for later reference.
func newDeviceRegistration(ctx context.Context) *api.DeviceRegistrationRequest {
	dev := &api.DeviceRegistrationRequest{
		AppName:    preferences.AppName,
		AppVersion: preferences.AppVersion(),
		AppID:      preferences.DefaultAppID,
	}

	// Retrieve the name as the device name.
	name, err := info.GetHostname()
	if err != nil {
		logging.FromContext(ctx).Warn("Unable to determine device hostname.",
			slog.Any("error", err))
	}

	dev.DeviceName = name

	// Generate a new unique Device ID
	id, err := info.NewDeviceID()
	if err != nil {
		logging.FromContext(ctx).Warn("Unable to generate a device ID.",
			slog.Any("error", err))
	}

	dev.DeviceID = id

	// Retrieve the OS name and version.
	osName, osVersion, err := info.GetOSID()
	if err != nil {
		logging.FromContext(ctx).Warn("Unable to determine OS details.",
			slog.Any("error", err))
	}

	dev.OsName = osName
	dev.OsVersion = osVersion

	// Retrieve the hardware model and manufacturer.
	model, manufacturer, err := info.GetHWProductInfo()
	if err != nil {
		logging.FromContext(ctx).Warn("Unable to determine device hardware details.",
			slog.Any("error", err))
	}

	dev.Model = model
	dev.Manufacturer = manufacturer

	// Set the device id and name in the preferences store.
	if err := preferences.Set(
		preferences.SetDeviceID(dev.DeviceID),
		preferences.SetDeviceName(dev.DeviceName),
	); err != nil {
		logging.FromContext(ctx).Warn("Could not save device id/name.",
			slog.Any("error", err))
	}

	return dev
}

// generateAPIURL creates a URL to use for sending data back to Home
// Assistant from the registration information returned by Home Assistant. It
// follows the rules mentioned in the developer docs to generate the URL:
//
// https://developers.home-assistant.io/docs/api/native-app-integration/sending-data#sending-webhook-data-via-rest-api
func generateAPIURL(response *api.DeviceRegistrationResponse, request *preferences.Registration) (string, error) {
	switch {
	case response.CloudhookURL != nil && !request.IgnoreHassURLs:
		return *response.CloudhookURL, nil
	case response.RemoteUIURL != nil && response.WebhookID != "" && !request.IgnoreHassURLs:
		return *response.RemoteUIURL + WebHookPath + response.WebhookID, nil
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

	// Strip any port from host.
	websocketURL.Host = websocketURL.Hostname()

	return websocketURL.JoinPath(WebsocketPath).String(), nil
}
