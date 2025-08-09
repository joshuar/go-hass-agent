// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

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

	called := m.execute(ctx, args...)
	if called.Err != nil {
		return fmt.Errorf("%s: unable to call method %s (args: %v): %w",
			m.bus.busType.String(),
			m.name,
			args,
			called.Err)
	}

	return nil
}

func (m *Method) IntrospectArgs() ([]introspect.Arg, error) {
	if m.obj == nil {
		return nil, ErrIntrospectionNotAvail
	}

	return m.obj.Args, nil
}

func (m *Method) execute(ctx context.Context, args ...any) *dbus.Call {
	obj := m.bus.getObject(m.intr, m.path)

	if len(args) > 0 {
		cleanArgs, warnings := m.sanitizeArgs(args)
		if warnings != nil {
			m.bus.traceLog("Sanitized method arguments with warnings", slog.Any("warnings", warnings))
		}

		return obj.CallWithContext(ctx, m.name, 0, cleanArgs...)
	}

	return obj.CallWithContext(ctx, m.name, 0)
}

//nolint:funlen
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
				warnings = errors.Join(warnings,
					fmt.Errorf("could not convert argument %d, using default value: %w", idx, err))
			}

			cleanArgs[idx] = value
		case "i":
			value, err := VariantToValue[int32](variant)
			if err != nil {
				warnings = errors.Join(warnings,
					fmt.Errorf("could not convert argument %d, using default value: %w", idx, err))
			}

			cleanArgs[idx] = value
		case "a{sv}":
			value, err := VariantToValue[map[string]any](variant)
			if err != nil {
				warnings = errors.Join(warnings,
					fmt.Errorf("could not convert argument %d, using default value: %w", idx, err))
			}

			cleanArgs[idx] = value
		case "as":
			value, err := VariantToValue[[]string](variant)
			if err != nil {
				warnings = errors.Join(warnings,
					fmt.Errorf("could not convert argument %d, using default value: %w", idx, err))
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

type Data[T any] struct {
	*Method
}

func (d *Data[T]) Fetch(ctx context.Context, args ...any) (T, error) {
	var stored T

	called := d.execute(ctx, args...)
	if called.Err != nil {
		return stored, fmt.Errorf("%s: unable to call method %s (args: %v): %w",
			d.bus.busType.String(),
			d.name,
			args,
			called.Err)
	}

	if err := called.Store(&stored); err != nil {
		return stored, fmt.Errorf("%s: unable to store method results: %w", d.bus.busType.String(), err)
	}

	return stored, nil
}

func NewData[T any](bus *Bus, intr, path, name string) *Data[T] {
	return &Data[T]{
		Method: &Method{
			bus:  bus,
			intr: intr,
			path: path,
			name: name,
		},
	}
}
