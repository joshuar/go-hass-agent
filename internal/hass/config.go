// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/perimeterx/marshmallow"
	"github.com/rs/zerolog/log"
)

type HassConfig struct {
	rawConfigProps map[string]interface{}
	hassConfigProps
	mu sync.Mutex
}

type hassConfigProps struct {
	Entities   map[string]map[string]interface{} `json:"entities"`
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

func NewHassConfig(ctx context.Context) *HassConfig {
	c := &HassConfig{}
	if err := c.Refresh(ctx); err != nil {
		log.Debug().Err(err).
			Msg("Could not fetch config from Home Assistant.")
		return nil
	}
	return c
}

func (h *HassConfig) GetEntityState(entity string) map[string]interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()
	if v, ok := h.Entities[entity]; ok {
		return v
	}
	return nil
}

func (h *HassConfig) IsEntityDisabled(entity string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if v, ok := h.Entities[entity]["disabled"]; ok {
		if disabledState, ok := v.(bool); !ok {
			return false
		} else {
			return disabledState
		}
	}
	return false
}

// HassConfig implements hass.Request so that it can be sent as a request to HA
// to get its data.

func (h *HassConfig) RequestType() RequestType {
	return requestTypeGetConfig
}

func (h *HassConfig) RequestData() json.RawMessage {
	return nil
}

func (h *HassConfig) ResponseHandler(resp bytes.Buffer) {
	if resp.Bytes() == nil {
		log.Debug().
			Msg("No response returned.")
		return
	}
	h.mu.Lock()
	result, err := marshmallow.Unmarshal(resp.Bytes(), &h.hassConfigProps)
	if err != nil {
		log.Debug().Err(err).
			Msg("Couldn't unmarshal Hass config.")
		return
	}
	h.rawConfigProps = result
	h.mu.Unlock()
}

// HassConfig implements config.Config

func (c *HassConfig) Get(property string) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if value, ok := c.rawConfigProps[property]; ok {
		return value, nil
	} else {
		return nil, fmt.Errorf("config does not have an option %s", property)
	}
}

func (c *HassConfig) Set(property string, value interface{}) error {
	log.Debug().Caller().Msg("Hass configuration is not settable.")
	return nil
}

func (c *HassConfig) Validate() error {
	log.Debug().Caller().Msg("Hass configuration has no validation.")
	return nil
}

func (h *HassConfig) Refresh(ctx context.Context) error {
	APIRequest(ctx, h)
	return nil
}

func (h *HassConfig) Upgrade() error {
	log.Debug().Caller().Msg("Hass configuration has no upgrades.")
	return nil
}
