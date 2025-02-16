// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package api

import "errors"

var (
	ErrInvalidEntityConfig = errors.New("entity has invalid config")
	ErrInvalidConfig       = errors.New("invalid config")
)

// GetVersion returns the version of Home Assistant from the config response, or
// "Unknown" if it was not set.
func (c *ConfigResponse) GetVersion() string {
	if c.Version != nil {
		return *c.Version
	}

	return "Unknown"
}

// IsEntityDisabled returns whether the entity with the given ID has been
// disabled in Home Assistant. If the disabled status could not be determined,
// it will return a non-nil error.
func (c *ConfigResponse) IsEntityDisabled(id string) (bool, error) {
	if c.Entities == nil {
		return false, ErrInvalidConfig
	}

	entities := *c.Entities

	if v, ok := entities[id]["disabled"]; ok {
		disabledState, ok := v.(bool)
		if !ok {
			return false, nil
		}

		return disabledState, nil
	}

	return false, nil
}
