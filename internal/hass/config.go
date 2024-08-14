// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:errname // structs are dual-purpose response and error
//revive:disable:unused-receiver
package hass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrInvalidEntityConfig = errors.New("entity has invalid config")
	ErrInvalidConfig       = errors.New("invalid config")
)

type Config struct {
	Details *ConfigEntries
	*APIError
	mu sync.Mutex
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
	if c.Details == nil {
		return false, ErrInvalidConfig
	}

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

func (c *Config) UnmarshalError(data []byte) error {
	err := json.Unmarshal(data, c.APIError)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func (c *Config) Error() string {
	return c.APIError.Error()
}

type configRequest struct{}

func (c *configRequest) RequestBody() json.RawMessage {
	return json.RawMessage(`{ "type": "get_config" }`)
}

func GetConfig(ctx context.Context, url string) (*Config, error) {
	client := NewDefaultHTTPClient(url)

	req := &configRequest{}
	resp := &Config{}

	if err := ExecuteRequest(ctx, client, "", req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
