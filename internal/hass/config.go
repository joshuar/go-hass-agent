// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/perimeterx/marshmallow"
	"github.com/rs/zerolog/log"
)

type HassConfig struct {
	mu sync.Mutex
	hassConfigProps
	rawConfigProps map[string]interface{}
}

type hassConfigProps struct {
	Components   []string                          `json:"components"`
	Entities     map[string]map[string]interface{} `json:"entities"`
	ConfigDir    string                            `json:"config_dir"`
	Elevation    int                               `json:"elevation"`
	Latitude     float64                           `json:"latitude"`
	LocationName string                            `json:"location_name"`
	Longitude    float64                           `json:"longitude"`
	TimeZone     string                            `json:"time_zone"`
	UnitSystem   struct {
		Length      string `json:"length"`
		Mass        string `json:"mass"`
		Temperature string `json:"temperature"`
		Volume      string `json:"volume"`
	} `json:"unit_system"`
	Version               string   `json:"version"`
	WhitelistExternalDirs []string `json:"whitelist_external_dirs"`
}

func NewHassConfig(ctx context.Context) *HassConfig {
	c := &HassConfig{}
	c.Refresh(ctx)
	c.updater(ctx)
	return c
}

func (h *HassConfig) Refresh(ctx context.Context) {
	APIRequest(ctx, h)
}

func (h *HassConfig) updater(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.Refresh(ctx)
			}
		}
	}()
}

func (h *HassConfig) GetEntityState(entity string) map[string]interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()
	if v, ok := h.hassConfigProps.Entities[entity]; ok {
		return v
	}
	return nil
}

func (h *HassConfig) IsEntityDisabled(entity string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if v, ok := h.hassConfigProps.Entities[entity]; ok {
		return v["disabled"].(bool)
	} else {
		return false
	}
}

// HassConfig implements hass.Request so that it can be sent as a request to HA
// to get its data.

func (h *HassConfig) RequestType() RequestType {
	return RequestTypeGetConfig
}

func (h *HassConfig) RequestData() interface{} {
	return struct{}{}
}

func (h *HassConfig) ResponseHandler(resp bytes.Buffer) {
	h.mu.Lock()
	result, err := marshmallow.Unmarshal(resp.Bytes(), &h.hassConfigProps)
	if err != nil {
		log.Debug().Err(err).
			Msg("Couldn't unmarshal Hass config.")
		return
	}
	h.rawConfigProps = result
	h.mu.Unlock()
	log.Debug().Caller().Msg("Updated stored HA config.")
}
