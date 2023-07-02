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

func (item *registryItem) IsDisabled() bool {
	return item.data.Disabled
}

func (item *registryItem) IsRegistered() bool {
	return item.data.Registered
}

type Registry interface {
	Open(context.Context, fyne.URI) error
	Get(string) (*registryItem, error)
	Set(registryItem) error
	Close() error
}
