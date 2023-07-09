// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/api"
	"github.com/perimeterx/marshmallow"
	"github.com/rs/zerolog/log"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
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

func NewHassConfig(ctx context.Context) (*HassConfig, error) {
	c := &HassConfig{}
	if err := c.Refresh(ctx); err != nil {
		return nil, err
	}
	return c, nil
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

func (h *HassConfig) RequestType() api.RequestType {
	return api.RequestTypeGetConfig
}

func (h *HassConfig) RequestData() json.RawMessage {
	return nil
}

func (h *HassConfig) ResponseHandler(resp bytes.Buffer, respCh chan api.Response) {
	if resp.Bytes() == nil {
		err := errors.New("no response returned")
		response := api.NewGenericResponse(err, api.RequestTypeGetConfig)
		respCh <- response
		return
	}
	h.mu.Lock()
	result, err := marshmallow.Unmarshal(resp.Bytes(), &h.hassConfigProps)
	if err != nil {
		response := api.NewGenericResponse(err, api.RequestTypeGetConfig)
		respCh <- response
		return
	}
	h.rawConfigProps = result
	h.mu.Unlock()
	response := api.NewGenericResponse(nil, api.RequestTypeGetConfig)
	respCh <- response
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
	respCh := make(chan api.Response, 1)
	defer close(respCh)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		api.ExecuteRequest(ctx, h, respCh)
	}()
	response := <-respCh
	if response.Error() != nil {
		return response.Error()
	}

	return nil
}

func (h *HassConfig) Upgrade() error {
	log.Debug().Caller().Msg("Hass configuration has no upgrades.")
	return nil
}
