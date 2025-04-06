// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package models contains the common objects and methods on these objects, used
// by the agent and its internal packages.
package models

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml api.yaml

import (
	"cmp"
	"time"
)

// Option is a functional option for type T. Any concrete type can define its
// own functional options using this.
//
// type MyType Option[*MyType].
type Option[T any] func(T) error

type StateValue interface {
	cmp.Ordered | time.Time
}
