// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
)

//go:generate mockery --name Registry
type Registry interface {
	Open(context.Context, string) error
	Close() error
	SetDisabled(string, bool) error
	SetRegistered(string, bool) error
	IsDisabled(string) bool
	IsRegistered(string) bool
}

type SensorMetadata struct {
	Registered bool `json:"Registered"`
	Disabled   bool `json:"Disabled"`
}
