// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type Config struct {
	err     error
	Details *ConfigEntries
	mu      sync.Mutex
}

type ConfigEntries struct {
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
}

type units struct {
	Length      string `json:"length"`
	Mass        string `json:"mass"`
	Temperature string `json:"temperature"`
	Volume      string `json:"volume"`
}

func (c *Config) IsEntityDisabled(entity string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.Details.Entities[entity]["disabled"]; ok {
		if disabledState, ok := v.(bool); !ok {
			return false, nil
		} else {
			return disabledState, nil
		}
	}
	return false, nil
}

func (c *Config) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &c.Details)
}

func (c *Config) StoreError(e error) {
	c.err = e
}

func (c *Config) Error() string {
	return c.err.Error()
}

type configRequest struct{}

func (c *configRequest) RequestBody() json.RawMessage {
	return json.RawMessage(`{ "type": "get_config" }`)
}

func GetConfig(ctx context.Context) (*Config, error) {
	prefs, err := preferences.Load()
	if err != nil {
		return nil, errors.New("could not load preferences")
	}
	ctx = ContextSetURL(ctx, prefs.RestAPIURL)
	ctx = ContextSetClient(ctx, resty.New())

	req := &configRequest{}
	resp := &Config{}

	ExecuteRequest(ctx, req, resp)
	if errors.Is(resp, &APIError{}) {
		return nil, resp
	}

	return resp, nil
}
