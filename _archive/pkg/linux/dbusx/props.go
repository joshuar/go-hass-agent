// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package dbusx

import (
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"
)

// Property represents a D-Bus property.
type Property[P any] struct {
	bus  *Bus   `validate:"required"`
	path string `validate:"required"`
	intr string `validate:"required"`
	name string `validate:"required"`
}

// Get retrieves the value of the property from D-Bus. If the property cannot be
// retrieved, a non-nil error is returned.
func (p *Property[P]) Get() (P, error) {
	var value P

	if err := valid(p); err != nil {
		return value, fmt.Errorf("invalid property: %w", err)
	}

	p.bus.traceLog("Requesting property.",
		slog.String("path", p.path),
		slog.String("dest", p.intr),
		slog.String("property", p.name))

	obj := p.bus.getObject(p.intr, p.path)

	res, err := obj.GetProperty(p.name)
	if err != nil {
		return value,
			fmt.Errorf("%s: unable to retrieve property %s from %s: %w", p.bus.busType.String(), p.name, p.intr, err)
	}

	value, err = VariantToValue[P](res)
	if err != nil {
		return value,
			fmt.Errorf("%s: unable to retrieve property %s from %s: %w", p.bus.busType.String(), p.name, p.intr, err)
	}

	return value, nil
}

// Set sets the property to the specified value.
func (p *Property[P]) Set(value P) error {
	if err := valid(p); err != nil {
		return fmt.Errorf("invalid property: %w", err)
	}

	p.bus.traceLog("Setting property.",
		slog.String("path", p.path),
		slog.String("dest", p.intr),
		slog.String("property", p.name),
		slog.Any("value", value))

	v := dbus.MakeVariant(value)
	obj := p.bus.getObject(p.intr, p.path)

	err := obj.SetProperty(p.name, v)
	if err != nil {
		return fmt.Errorf("%s: unable to set property %s (%s) to %v: %w",
			p.bus.busType.String(),
			p.name,
			p.intr,
			value,
			err)
	}

	return nil
}

func NewProperty[P any](bus *Bus, path, intr, name string) *Property[P] {
	return &Property[P]{
		bus:  bus,
		intr: intr,
		path: path,
		name: name,
	}
}

// Properties represents a signal that matches the canonical
// org.freedesktop.DBus.PropertiesChanged signature. These will have an
// interface name together with a list of changed properties (and their values)
// and invalidated property names.
type Properties struct {
	Interface   string
	Changed     map[string]dbus.Variant
	Invalidated []string
}

// ParsePropertiesChanged treats the given signal body as matching the canonical
// org.freedesktop.DBus.PropertiesChanged signature and will parse it into a
// Properties structure that is easier to use. If the signal body cannot be
// parsed an error will be returned with details of the problem. Adapted from
// https://github.com/godbus/dbus/issues/201
//
//nolint:mnd
func ParsePropertiesChanged(propsChanged []any) (*Properties, error) {
	props := &Properties{}

	var ok bool

	if len(propsChanged) != 3 {
		return nil, ErrNotPropChanged
	}

	props.Interface, ok = propsChanged[0].(string)
	if !ok {
		return nil, ErrParseInterface
	}

	props.Changed, ok = propsChanged[1].(map[string]dbus.Variant)
	if !ok {
		return nil, ErrParseNewProps
	}

	props.Invalidated, ok = propsChanged[2].([]string)
	if !ok {
		return nil, ErrParseOldProps
	}

	return props, nil
}

// HasPropertyChanged checks if the given property has been changed in the given
// signal. The given signal should match the canonical
// org.freedesktop.DBus.PropertiesChanged signature. Returned values will be a
// boolean indicating whether the property changed and the new value (or the
// default value) of the specified type for the property. If any errors
// occurred, a non-nil error will be the third return value.
func HasPropertyChanged[T any](event []any, property string) (bool, T, error) {
	var value T
	props, err := ParsePropertiesChanged(event)
	if err != nil {
		return false, value, fmt.Errorf("cannot parse event: %w", err)
	}

	if variant, changed := props.Changed[property]; changed {
		if value, err = VariantToValue[T](variant); err != nil {
			return true, value, fmt.Errorf("cannot convert variant to requested value: %w", err)
		}

		return true, value, nil
	}

	return false, value, nil
}
