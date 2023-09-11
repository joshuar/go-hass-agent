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
	"sync"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/perimeterx/marshmallow"
)

const (
	websocketPath = "/api/websocket"
	webHookPath   = "/api/webhook/"
)

type haConfig struct {
	rawConfigProps map[string]interface{}
	haConfigProps
	mu sync.Mutex
}

type haConfigProps struct {
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

// HassConfig implements hass.Request so that it can be sent as a request to HA
// to get its data.

func (h *haConfig) RequestType() api.RequestType {
	return api.RequestTypeGetConfig
}

func (h *haConfig) RequestData() json.RawMessage {
	return nil
}

func (h *haConfig) ResponseHandler(resp bytes.Buffer, respCh chan api.Response) {
	if resp.Bytes() == nil {
		err := errors.New("no response returned")
		response := api.NewGenericResponse(err, api.RequestTypeGetConfig)
		respCh <- response
		return
	}
	h.mu.Lock()
	result, err := marshmallow.Unmarshal(resp.Bytes(), &h.haConfigProps)
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

func GetHassConfig(ctx context.Context, c api.AgentConfig) (*haConfig, error) {
	h := new(haConfig)
	respCh := make(chan api.Response, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		api.ExecuteRequest(ctx, h, c, respCh)
	}()
	response := <-respCh
	if response.Error() != nil {
		return nil, response.Error()
	}
	return h, nil
}

func (h *haConfig) GetRegisteredEntities() map[string]map[string]interface{} {
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
