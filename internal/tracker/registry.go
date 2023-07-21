// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
)

const (
	registryStorageID = "sensorRegistry"
)

//go:generate moq -out mock_Registry_test.go . Registry
type Registry interface {
	Open(context.Context, string) error
	Close() error
	SetDisabled(string, bool) error
	SetRegistered(string, bool) error
	IsDisabled(string) bool
	IsRegistered(string) bool
}
