// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/preferences"
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

func (c *Config) URL() string {
	prefs, err := preferences.Load()
	if err != nil {
		return ""
	}
	return prefs.RestAPIURL
}

func (c *Config) RequestBody() json.RawMessage {
	return json.RawMessage(`{ "type": "get_config" }`)
}

func (c *Config) ResponseBody() any { return c }

func GetConfig(ctx context.Context) (*Config, error) {
	c := new(Config)

	resp := <-api.ExecuteRequest2(ctx, c)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Body.(*Config), nil
}
