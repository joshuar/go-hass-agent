// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"time"

	"fyne.io/fyne/v2/data/binding"
	"github.com/carlmjohnson/requests"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type RegistrationDetails struct {
	Server, Token binding.String
	Device        DeviceInfo
}

func (r *RegistrationDetails) Validate() bool {
	validate := validator.New()
	check := func(value string, validation string) bool {
		if err := validate.Var(value, validation); err != nil {
			return false
		}
		return true
	}
	if server, _ := r.Server.Get(); !check(server, "required,http_url") {
		return false
	}
	if token, _ := r.Token.Get(); !check(token, "required") {
		return false
	}
	if r.Device == nil {
		return false
	}
	return true
}

type RegistrationResponse struct {
	CloudhookURL string `json:"cloudhook_url"`
	RemoteUIURL  string `json:"remote_ui_url"`
	Secret       string `json:"secret"`
	WebhookID    string `json:"webhook_id"`
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

func RegisterWithHass(registration *RegistrationDetails) (*RegistrationResponse, error) {
	request, err := registration.Device.MarshalJSON()
	if err != nil {
		log.Debug().Err(err).Msg("Unable to generate registration request.")
		return nil, err
	}
	token, _ := registration.Token.Get()
	host, _ := registration.Server.Get()

	var response *RegistrationResponse
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err = requests.
		URL(host+"/api/mobile_app/registrations").
		Header("Authorization", "Bearer "+token).
		BodyBytes(request).
		ToJSON(&response).
		Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return response, nil
}
