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
	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog/log"
)

type RegistrationHost struct {
	Server, Token binding.String
	UseTLS        binding.Bool
}

type RegistrationResponse struct {
	CloudhookURL string `json:"cloudhook_url"`
	RemoteUIURL  string `json:"remote_ui_url"`
	Secret       string `json:"secret"`
	WebhookID    string `json:"webhook_id"`
}

type RegistrationRequest struct {
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
	AppData            interface{} `json:"app_data,omitempty"`
}

func RegisterWithHass(ri *RegistrationHost, rr *RegistrationRequest) *RegistrationResponse {
	res := &RegistrationResponse{}

	token, _ := ri.Token.Get()

	var host string
	server, _ := ri.Server.Get()
	if v, _ := ri.UseTLS.Get(); v {
		host = "https://" + server
	} else {
		host = "http://" + server
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	requestFunc := func() error {
		return requests.
			URL(host+"/api/mobile_app/registrations").
			Header("Authorization", "Bearer "+token).
			BodyJSON(&rr).
			ToJSON(&res).
			Fetch(ctx)
	}
	retryNotifyFunc := func(e error, d time.Duration) {
		log.Debug().Err(e).
			Msgf("Retrying registration request in %v seconds.", d.Seconds())
	}
	err := backoff.RetryNotify(requestFunc, backoff.NewExponentialBackOff(), retryNotifyFunc)

	if err != nil {
		log.Error().Msgf("Unable to register: %v", err)
		return nil
	} else {
		return res
	}
}
