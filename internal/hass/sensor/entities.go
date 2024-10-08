// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT
package sensor

import (
	"encoding/json"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

const (
	StateUnknown = "Unknown"
)

type State struct {
	Value      any              `json:"state" validate:"required"`
	Attributes map[string]any   `json:"attributes,omitempty" validate:"omitempty"`
	Icon       string           `json:"icon,omitempty" validate:"omitempty,startswith=mdi:"`
	ID         string           `json:"unique_id" validate:"required"`
	EntityType types.SensorType `json:"type" validate:"omitempty"`
}

//nolint:wrapcheck
func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		State      any            `json:"state" validate:"required"`
		Attributes map[string]any `json:"attributes,omitempty" validate:"omitempty"`
		Icon       string         `json:"icon,omitempty" validate:"omitempty,startswith=mdi:"`
		ID         string         `json:"unique_id" validate:"required"`
		EntityType string         `json:"type" validate:"omitempty"`
	}{
		State:      s.Value,
		Attributes: s.Attributes,
		Icon:       s.Icon,
		ID:         s.ID,
		EntityType: s.EntityType.String(),
	})
}

type Entity struct {
	*State
	Name        string            `json:"name" validate:"required"`
	Units       string            `json:"unit_of_measurement,omitempty" validate:"omitempty"`
	DeviceClass types.DeviceClass `json:"device_class,omitempty" validate:"omitempty"`
	StateClass  types.StateClass  `json:"state_class,omitempty" validate:"omitempty"`
	Category    types.Category    `json:"entity_category,omitempty" validate:"omitempty"`
}

//nolint:wrapcheck
func (e Entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		State       any            `json:"state" validate:"required"`
		Attributes  map[string]any `json:"attributes,omitempty" validate:"omitempty"`
		Icon        string         `json:"icon,omitempty" validate:"omitempty,startswith=mdi:"`
		ID          string         `json:"unique_id" validate:"required"`
		EntityType  string         `json:"type" validate:"omitempty"`
		Name        string         `json:"name" validate:"required"`
		Units       string         `json:"unit_of_measurement,omitempty" validate:"omitempty"`
		DeviceClass string         `json:"device_class,omitempty" validate:"omitempty"`
		StateClass  string         `json:"state_class,omitempty" validate:"omitempty"`
		Category    string         `json:"entity_category,omitempty" validate:"omitempty"`
	}{
		State:       e.Value,
		Attributes:  e.Attributes,
		Icon:        e.Icon,
		ID:          e.ID,
		EntityType:  e.EntityType.String(),
		Name:        e.Name,
		Units:       e.Units,
		DeviceClass: e.DeviceClass.String(),
		StateClass:  e.StateClass.String(),
		Category:    e.Category.String(),
	})
}
