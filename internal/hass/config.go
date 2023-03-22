package hass

import (
	"context"

	"github.com/carlmjohnson/requests"
	"github.com/rs/zerolog/log"
)

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
		Type: RequestTypeGetConfig,
	}
	res := &ConfigResponse{}
	ctx := context.Background()
	err := requests.
		URL(host).
		BodyJSON(&req).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		log.Error().Caller().
			Msgf("Unable to fetch config: %v", err)
		return nil
	} else {
		log.Debug().Msg("Configuration fetched successfully")
		return res
	}
}
