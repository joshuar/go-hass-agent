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
		return fmt.Errorf("%w: %s", ErrValidationFailed, parseValidationErrors(err))
	}

	return nil
}

func (p *Registration) IsDefault() bool {
	return p.Server == defaultServer && p.Token == defaultSecret
}
