// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
)

type Config struct {
	Entities              map[string]map[string]any `json:"entities"`
	UnitSystem            units                     `json:"unit_system"`
	ConfigDir             string                    `json:"config_dir"`
	LocationName          string                    `json:"location_name"`
	TimeZone              string                    `json:"time_zone"`
	Version               string                    `json:"version"`
	Components            []string                  `json:"components"`
	WhitelistExternalDirs []string                  `json:"whitelist_external_dirs"`
	Elevation             int                       `json:"elevation"`
	Latitude              float64                   `json:"latitude"`
	Longitude             float64                   `json:"longitude"`
	mu                    sync.Mutex                `json:"-"`
}

type units struct {
	Length      string `json:"length"`
	Mass        string `json:"mass"`
	Temperature string `json:"temperature"`
	Volume      string `json:"volume"`
}

// HassConfig implements hass.Request so that it can be sent as a request to HA
// to get its data.

func (c *Config) RequestType() api.RequestType {
	return api.RequestTypeGetConfig
}

func (c *Config) RequestData() json.RawMessage {
	return nil
}

func (c *Config) IsEntityDisabled(entity string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.Entities[entity]["disabled"]; ok {
		if disabledState, ok := v.(bool); !ok {
			return false, nil
		} else {
			return disabledState, nil
		}
	}
	return false, nil
}

func (c *Config) extractConfig(b []byte) {
	if b == nil {
		log.Warn().Msg("No config returned.")
		return
	}
	c.mu.Lock()
	err := json.Unmarshal(b, &c)
	if err != nil {
		log.Warn().Msg("Could not extract config structure.")
	}
	c.mu.Unlock()
}

func GetConfig(ctx context.Context) (*Config, error) {
	h := new(Config)
	response := <-api.ExecuteRequest(ctx, h)
	switch r := response.(type) {
	case []byte:
		h.extractConfig(r)
	case error:
		log.Warn().Err(r).Msg("Failed to fetch Home Assistant config.")
	default:
		log.Warn().Msgf("Unknown response type %T", r)
	}
	return h, nil
}
