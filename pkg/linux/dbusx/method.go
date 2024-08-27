// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"context"
	"fmt"
)

//revive:disable:struct-tag
type Method struct {
	bus  *Bus   `validate:"required"`
	path string `validate:"required"`
	intr string `validate:"required"`
	name string `validate:"required"`
}

func (m *Method) Call(ctx context.Context, args ...any) error {
	if err := valid(m); err != nil {
		return fmt.Errorf("invalid method: %w", err)
	}

	obj := m.bus.getObject(m.intr, m.path)

	var err error

	if len(args) > 0 {
		err = obj.CallWithContext(ctx, m.name, 0, args...).Err
	} else {
		err = obj.CallWithContext(ctx, m.name, 0).Err
	}

	if err != nil {
		return fmt.Errorf("%s: unable to call method %s (args: %v): %w", m.bus.busType.String(), m.name, args, err)
	}

	return nil
}

func NewMethod(bus *Bus, intr, path, name string) *Method {
	return &Method{
		bus:  bus,
		intr: intr,
		path: path,
		name: name,
	}
}
