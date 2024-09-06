// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package dbusx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

var ErrIntrospectionNotAvail = errors.New("introspection not available on object")

//revive:disable:struct-tag
type Method struct {
	obj  *introspect.Method
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

	if len(args) > 0 {
		cleanArgs, warnings := m.sanitizeArgs(args)
		if warnings != nil {
			m.bus.traceLog("Sanitized method arguments with warnings", slog.Any("warnings", warnings))
		}

		err := obj.CallWithContext(ctx, m.name, 0, cleanArgs...).Err
		if err != nil {
			return fmt.Errorf("%s: unable to call method %s (args: %v): %w", m.bus.busType.String(), m.name, cleanArgs, err)
		}
	} else {
		err := obj.CallWithContext(ctx, m.name, 0).Err
		if err != nil {
			return fmt.Errorf("%s: unable to call method %s: %w", m.bus.busType.String(), m.name, err)
		}
	}

	return nil
}

func (m *Method) IntrospectArgs() ([]introspect.Arg, error) {
	if m.obj == nil {
		return nil, ErrIntrospectionNotAvail
	}

	return m.obj.Args, nil
}

//nolint:gocognit
func (m *Method) sanitizeArgs(args []any) ([]any, error) {
	introspection, err := NewIntrospection(m.bus, m.intr, m.path)
	if err != nil {
		return nil, fmt.Errorf("could not introspect: %w", err)
	}

	method, err := introspection.GetMethod(m.name)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve method details: %w", err)
	}

	methodArgs, err := method.IntrospectArgs()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve method arguments: %w", err)
	}

	methodArgs = slices.DeleteFunc(methodArgs, func(e introspect.Arg) bool {
		return e.Direction == "out"
	})
	cleanArgs := make([]any, len(args))

	var warnings error

	for idx, arg := range methodArgs {
		var variant dbus.Variant

		sig, err := dbus.ParseSignature(arg.Type)
		if err != nil {
			m.bus.traceLog("Could not parse argument signature. Attempting to guess signature.", slog.Any("error", err))

			variant = dbus.MakeVariant(args[idx])
		} else {
			variant = dbus.MakeVariantWithSignature(args[idx], sig)
		}

		switch arg.Type {
		case "u":
			value, err := VariantToValue[uint32](variant)
			if err != nil {
				warnings = errors.Join(warnings, fmt.Errorf("could not convert argument %d, using default value: %w", idx, err))
			}

			cleanArgs[idx] = value
		case "i":
			value, err := VariantToValue[int32](variant)
			if err != nil {
				warnings = errors.Join(warnings, fmt.Errorf("could not convert argument %d, using default value: %w", idx, err))
			}

			cleanArgs[idx] = value
		case "a{sv}":
			value, err := VariantToValue[map[string]any](variant)
			if err != nil {
				warnings = errors.Join(warnings, fmt.Errorf("could not convert argument %d, using default value: %w", idx, err))
			}

			cleanArgs[idx] = value
		case "as":
			value, err := VariantToValue[[]string](variant)
			if err != nil {
				warnings = errors.Join(warnings, fmt.Errorf("could not convert argument %d, using default value: %w", idx, err))
			}

			cleanArgs[idx] = value
		default:
			cleanArgs[idx] = variant.Value()
		}
	}

	return cleanArgs, warnings
}

func NewMethod(bus *Bus, intr, path, name string) *Method {
	return &Method{
		bus:  bus,
		intr: intr,
		path: path,
		name: name,
	}
}
