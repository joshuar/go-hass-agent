// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"context"
)

//go:generate mockery --name Registry --inpackage
type Registry interface {
	Open(context.Context, string) error
	// Get(string) (*RegistryItem, error)
	// Set(RegistryItem) error
	Close() error
	SetDisabled(string, bool) error
	SetRegistered(string, bool) error
	IsDisabled(string) bool
	IsRegistered(string) bool
}

// type RegistryItem struct {
// 	data *SensorMetadata
// 	ID   string
// }

// func (item *RegistryItem) SetDisabled(state bool) {
// 	item.data.Disabled = state
// }

// func (item *RegistryItem) IsDisabled() bool {
// 	return item.data.Disabled
// }

// func (item *RegistryItem) SetRegistered(state bool) {
// 	item.data.Registered = state
// }

// func (item *RegistryItem) IsRegistered() bool {
// 	return item.data.Registered
// }

// func (item *RegistryItem) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(item.data)
// }

// func (item *RegistryItem) UnmarshalJSON(b []byte) error {
// 	return json.Unmarshal(b, item.data)
// }

// func NewRegistryItem(id string) *RegistryItem {
// 	return &RegistryItem{
// 		ID:   id,
// 		data: new(SensorMetadata),
// 	}
// }

type SensorMetadata struct {
	Registered bool `json:"Registered"`
	Disabled   bool `json:"Disabled"`
}
