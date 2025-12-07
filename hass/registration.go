// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/device"
	"github.com/joshuar/go-hass-agent/hass/api"
	"github.com/joshuar/go-hass-agent/validation"
)

const (
	registrationPath = "/api/mobile_app/registrations"
	websocketPath    = "/api/websocket"
	webHookPath      = "/api/webhook/"
)

// RegistrationRequest are the preferences that defines how Go Hass Agent registers
// with Home Assistant.
type RegistrationRequest struct {
	Server         string `toml:"server" form:"server"           validate:"required,http_url" help:"URL of the Home Assistant server."`
	Token          string `toml:"token"  form:"token"            validate:"required"          help:"Personal Access Token obtained from Home Assistant."`
	IgnoreHassURLS bool   `toml:"-"      form:"ignore_hass_urls" validate:"omitempty,boolean" help:"Ignore URLs returned by Home Assistant and use provided server for access." json:"-"`
}

// Valid checks whether the registration request details are valid.
func (r *RegistrationRequest) Valid() (bool, error) {
	err := validation.Validate.Struct(r)
	if err != nil {
		return false, fmt.Errorf("%w: %s", validation.ErrValidation, validation.ParseValidationErrors(err))
	}
	_, err = url.Parse(r.Server)
	if err != nil {
		return false, fmt.Errorf("%w: %w", validation.ErrValidation, err)
	}

	return true, nil
}

func (r *RegistrationRequest) Sanitise() error {
	return nil
}

// Register will register the device with Home Assistant. It uses the details entered by the user for the server/token
// and receives back from Home Assistant URLs and details needed for subsequent API requests.
func Register(ctx context.Context, id string, details *RegistrationRequest) error {
	req := newDeviceRegistration(ctx, id)
	resp := api.DeviceRegistrationResponse{}

	// Set up the api request, and the request/response bodies.
	apiReq := api.NewClient().R().SetContext(ctx)
	apiReq.SetAuthToken(details.Token)
	apiReq.SetBody(req)
	apiReq = apiReq.SetResult(&resp)

	registrationURL, err := url.Parse(details.Server)
	if err != nil {
		return fmt.Errorf("unable to register: %w", err)
	}

	registrationURL = registrationURL.JoinPath(registrationPath)
	_, err = apiReq.Post(registrationURL.String())
	if err != nil {
		return fmt.Errorf("unable to register: %w", err)
	}

	// Generate a rest API URL.
	restAPIURL, err := generateAPIURL(&resp, details)
	if err != nil {
		return fmt.Errorf("unable to register: %w", err)
	}
	// Generate a websocket API URL.
	websocketAPIURL, err := generateWebsocketURL(details.Server)
	if err != nil {
		return fmt.Errorf("unable to register: %w", err)
	}

	// Save options to config.
	err = config.Set(map[string]any{
		ConfigPrefix + "." + ConfigAPIURL:       restAPIURL,
		ConfigPrefix + "." + ConfigWebsocketURL: websocketAPIURL,
		ConfigPrefix + "." + ConfigWebhookID:    resp.WebhookID,
		ConfigPrefix + "." + ConfigSecret:       resp.Secret,
		"registration.server":                   details.Server,
		"registration.token":                    details.Token,
	})
	if err != nil {
		return fmt.Errorf("unable to register: %w", err)
	}

	return nil
}

// newDeviceRegistration generates the details required to register a device with Home Assistant.
func newDeviceRegistration(ctx context.Context, id string) *api.DeviceRegistrationRequest {
	dev := &api.DeviceRegistrationRequest{
		AppName:    config.AppName,
		AppVersion: config.AppVersion,
		AppID:      config.AppID,
		AppData:    map[string]any{"push_websocket_channel": true},
		DeviceID:   id,
	}

	var err error

	// Retrieve the name as the device name.
	dev.DeviceName, err = device.GetHostname()
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Unable to determine device hostname.",
			slog.Any("error", err))
	}

	// Retrieve the OS name and version.
	dev.OsName, dev.OsVersion, err = device.GetOSID()
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Unable to determine OS details.",
			slog.Any("error", err))
	}

	// Retrieve the hardware model and manufacturer.
	dev.Model, dev.Manufacturer, err = device.GetHWProductInfo()
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Unable to determine device hardware details.",
			slog.Any("error", err))
	}

	return dev
}

// generateAPIURL creates a URL to use for sending data back to Home
// Assistant from the registration information returned by Home Assistant. It
// follows the rules mentioned in the developer docs to generate the URL:
//
// https://developers.home-assistant.io/docs/api/native-app-integration/sending-data#sending-webhook-data-via-rest-api
func generateAPIURL(response *api.DeviceRegistrationResponse, request *RegistrationRequest) (string, error) {
	switch {
	case response.CloudhookURL != "" && !request.IgnoreHassURLS:
		return response.CloudhookURL, nil
	case response.RemoteUIURL != "" && response.WebhookID != "" && !request.IgnoreHassURLS:
		return response.RemoteUIURL + webHookPath + response.WebhookID, nil
	default:
		apiURL, err := url.Parse(request.Server)
		if err != nil {
			return "", fmt.Errorf("could not parse registration server: %w", err)
		}

		return apiURL.JoinPath(webHookPath, response.WebhookID).String(), nil
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

	return websocketURL.JoinPath(websocketPath).String(), nil
}
