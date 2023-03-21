package hass

import (
	"context"

	"fyne.io/fyne/v2/data/binding"
	"github.com/carlmjohnson/requests"
	log "github.com/sirupsen/logrus"
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
	ctx := context.Background()
	err := requests.
		URL(host+"/api/mobile_app/registrations").
		Header("Authorization", "Bearer "+token).
		BodyJSON(&rr).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		log.Errorf("Unable to register: %v", err)
		return nil
	} else {
		return res
	}
}
