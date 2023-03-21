package hass

import (
	"context"

	"fyne.io/fyne/v2/data/binding"
	"github.com/carlmjohnson/requests"
	log "github.com/sirupsen/logrus"
)

type GenericRequest struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}
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

type ConfigResponse struct {
	Components   []string `json:"components"`
	ConfigDir    string   `json:"config_dir"`
	Elevation    int      `json:"elevation"`
	Latitude     float64  `json:"latitude"`
	LocationName string   `json:"location_name"`
	Longitude    float64  `json:"longitude"`
	TimeZone     string   `json:"time_zone"`
	UnitSystem   struct {
		Length      string `json:"length"`
		Mass        string `json:"mass"`
		Temperature string `json:"temperature"`
		Volume      string `json:"volume"`
	} `json:"unit_system"`
	Version               string   `json:"version"`
	WhitelistExternalDirs []string `json:"whitelist_external_dirs"`
}

func GetConfig(host string) *ConfigResponse {
	req := &GenericRequest{
		Type: "get_config",
	}
	res := &ConfigResponse{}
	ctx := context.Background()
	log.Debugf("Requesting to %s", host)
	err := requests.
		URL(host).
		BodyJSON(&req).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		log.Errorf("Unable to fetch config: %v", err)
		return nil
	} else {
		return res
	}

}
