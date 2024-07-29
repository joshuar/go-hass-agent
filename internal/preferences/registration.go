// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
package preferences

import (
	"fmt"
)

type Registration struct {
	Server string `toml:"server" validate:"required,http_url"`
	Token  string `toml:"token" validate:"required"`
}

func (p *Registration) Validate() error {
	err := validate.Struct(p)
	if err != nil {
		showValidationErrors(err)

		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
