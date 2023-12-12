// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/perimeterx/marshmallow"
	"github.com/rs/zerolog/log"
)

type haConfig struct {
	rawConfigProps map[string]any
	haConfigProps
	mu sync.Mutex
}

type haConfigProps struct {
	Entities   map[string]map[string]any `json:"entities"`
	UnitSystem struct {
		Length      string `json:"length"`
		Mass        string `json:"mass"`
		Temperature string `json:"temperature"`
		Volume      string `json:"volume"`
	} `json:"unit_system"`
	ConfigDir             string   `json:"config_dir"`
	LocationName          string   `json:"location_name"`
	TimeZone              string   `json:"time_zone"`
	Version               string   `json:"version"`
	Components            []string `json:"components"`
	WhitelistExternalDirs []string `json:"whitelist_external_dirs"`
	Elevation             int      `json:"elevation"`
	Latitude              float64  `json:"latitude"`
	Longitude             float64  `json:"longitude"`
}

// HassConfig implements hass.Request so that it can be sent as a request to HA
// to get its data.

func (h *haConfig) RequestType() api.RequestType {
	return api.RequestTypeGetConfig
}

func (h *haConfig) RequestData() json.RawMessage {
	return nil
}

func (h *haConfig) extractConfig(b []byte) {
	if b == nil {
		log.Warn().Msg("No config returned.")
		return
	}
	h.mu.Lock()
	result, err := marshmallow.Unmarshal(b, &h.haConfigProps)
	if err != nil {
		log.Warn().Msg("Could not extract config structure.")
	}
	h.rawConfigProps = result
	h.mu.Unlock()
}

func GetHassConfig(ctx context.Context) (*haConfig, error) {
	h := new(haConfig)
	response := <-api.ExecuteRequest(ctx, h)
	switch r := response.(type) {
	case []byte:
		h.extractConfig(response.([]byte))
	case error:
		log.Warn().Err(r).Msg("Failed to fetch Home Assistant config.")
	default:
		log.Warn().Msgf("Unknown response type %T", r)
	}
	return h, nil
}

func (h *haConfig) GetRegisteredEntities() map[string]map[string]any {
	return h.Entities
}

func (h *haConfig) IsEntityDisabled(entity string) (bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if v, ok := h.Entities[entity]["disabled"]; ok {
		if disabledState, ok := v.(bool); !ok {
			return false, nil
		} else {
			return disabledState, nil
		}
	}
	return false, nil
}

func (h *haConfig) GetVersion() string {
	return h.Version
}
