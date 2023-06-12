// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"

	"fyne.io/fyne/v2"
)

type registryItem struct {
	data *sensorMetadata
	id   string
}

type Registry interface {
	Open(context.Context, fyne.URI) error
	Get(string) (*registryItem, error)
	Set(registryItem) error
	Close() error
}
