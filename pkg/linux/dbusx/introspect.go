// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package dbusx

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/godbus/dbus/v5/introspect"
)

var ErrMethodNotFound = errors.New("method not found")

type Introspection introspect.Node

func (i Introspection) GetMethod(name string) (*Method, error) {
	for _, intr := range i.Interfaces {
		found := slices.IndexFunc(intr.Methods, func(e introspect.Method) bool {
			return strings.HasSuffix(name, e.Name)
		})

		if found != -1 {
			return &Method{
				name: name,
				intr: intr.Name,
				path: i.Name,
				obj:  &intr.Methods[found],
			}, nil
		}
	}

	return nil, ErrMethodNotFound
}

func NewIntrospection(bus *Bus, intr, path string) (*Introspection, error) {
	obj := bus.getObject(intr, path)

	node, err := introspect.Call(obj)
	if err != nil {
		return nil, fmt.Errorf("unable to introspect: %w", err)
	}

	nodeObj := Introspection(*node)

	return &nodeObj, nil
}
