// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package hass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrLoadPrefsFailed = errors.New("could not load preferences")

type Config struct {
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
		disabledState, ok := v.(bool)
		if !ok {
			return false, nil
		}

		return disabledState, nil
	}

	return false, nil
}

func (c *Config) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &c.Details)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	return nil
}

type configRequest struct{}

func (c *configRequest) RequestBody() json.RawMessage {
	return json.RawMessage(`{ "type": "get_config" }`)
}

//nolint:exhaustruct
func GetConfig(ctx context.Context) (*Config, error) {
	prefs, err := preferences.ContextGetPrefs(ctx)
	if err != nil {
		return nil, ErrLoadPrefsFailed
	}

	client := NewDefaultHTTPClient(prefs.RestAPIURL)

	req := &configRequest{}
	resp := &Config{}

	if err := ExecuteRequest(ctx, client, "", req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
