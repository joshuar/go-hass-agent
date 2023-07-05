// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"context"

	"fyne.io/fyne/v2"
)

type RegistryItem struct {
	data *sensorMetadata
	id   string
}

func (item *RegistryItem) IsDisabled() bool {
	return item.data.Disabled
}

func (item *RegistryItem) IsRegistered() bool {
	return item.data.Registered
}

func NewRegistryItem(id string) *RegistryItem {
	return &RegistryItem{
		id:   id,
		data: new(sensorMetadata),
	}
}

//go:generate mockery --name Registry
type Registry interface {
	Open(context.Context, fyne.URI) error
	Get(string) (*RegistryItem, error)
	Set(RegistryItem) error
	Close() error
}
