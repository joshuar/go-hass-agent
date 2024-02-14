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

func (c *Config) RequestBody() json.RawMessage {
	return json.RawMessage(`{ "type": "get_config" }`)
}

func (c *Config) ResponseBody() any { return c }

func GetConfig(ctx context.Context) (*Config, error) {
	prefs, err := preferences.Load()
	if err != nil {
		return nil, errors.New("could not load preferences")
	}
	ctx = ContextSetURL(ctx, prefs.RestAPIURL)
	ctx = ContextSetClient(ctx, resty.New())

	resp := <-ExecuteRequest(ctx, new(Config))
	if resp.Error != nil {
		return nil, resp.Error
	}

	var config *Config
	var ok bool
	if config, ok = resp.Body.(*Config); !ok {
		return nil, ErrResponseMalformed
	}
	return config, nil
}
