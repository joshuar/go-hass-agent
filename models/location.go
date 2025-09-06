// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package models

import (
	"github.com/joshuar/go-hass-agent/validation"
)

// Valid returns whether the location data is valid.
func (e *Location) Valid() bool {
	if err := validation.Validate.Struct(e); err != nil {
		return false
	}

	return true
}
