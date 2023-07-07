// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"github.com/joshuar/go-hass-agent/internal/hass/deviceClass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensorType"
	"github.com/joshuar/go-hass-agent/internal/hass/stateClass"
)

// Sensor represents an update for a sensor. It reflects the current state
// of the sensor at the point in time it is used. It provides a bridge between
// platform/device and HA implementations of what a sensor is.
//
//go:generate mockery --name Sensor
type Sensor interface {
	Name() string
	ID() string
	Icon() string
	SensorType() sensorType.SensorType
	DeviceClass() deviceClass.SensorDeviceClass
	StateClass() stateClass.SensorStateClass
	State() interface{}
	Units() string
	Category() string
	Attributes() interface{}
}

// sensorRegistrationInfo is the JSON structure required to register a sensor
// with HA.
type sensorRegistrationInfo struct {
	StateAttributes   interface{} `json:"attributes,omitempty"`
	DeviceClass       string      `json:"device_class,omitempty"`
	Icon              string      `json:"icon,omitempty"`
	Name              string      `json:"name"`
	State             interface{} `json:"state"`
	Type              string      `json:"type"`
	UniqueID          string      `json:"unique_id"`
	UnitOfMeasurement string      `json:"unit_of_measurement,omitempty"`
	StateClass        string      `json:"state_class,omitempty"`
	EntityCategory    string      `json:"entity_category,omitempty"`
	Disabled          bool        `json:"disabled,omitempty"`
}

// sensorUpdateInfo is the JSON structure required to update HA with the current
// sensor state.
type sensorUpdateInfo struct {
	StateAttributes interface{} `json:"attributes,omitempty"`
	Icon            string      `json:"icon,omitempty"`
	State           interface{} `json:"state"`
	Type            string      `json:"type"`
	UniqueID        string      `json:"unique_id"`
}

func MarshalSensorUpdate(s Sensor) *sensorUpdateInfo {
	return &sensorUpdateInfo{
		StateAttributes: s.Attributes(),
		Icon:            s.Icon(),
		State:           s.State(),
		Type:            marshalClass(s.SensorType()),
		UniqueID:        s.ID(),
	}
}

func MarshalSensorRegistration(s Sensor) *sensorRegistrationInfo {
	return &sensorRegistrationInfo{
		StateAttributes:   s.Attributes(),
		DeviceClass:       marshalClass(s.DeviceClass()),
		Icon:              s.Icon(),
		Name:              s.Name(),
		State:             s.State(),
		Type:              marshalClass(s.SensorType()),
		UniqueID:          s.ID(),
		UnitOfMeasurement: s.Units(),
		StateClass:        marshalClass(s.StateClass()),
		EntityCategory:    s.Category(),
		Disabled:          false,
	}
}

type ComparableStringer interface {
	comparable
	String() string
}

func returnZero[T any](s ...T) T {
	var zero T
	return zero
}

func marshalClass[C ComparableStringer](class C) string {
	if class == returnZero[C](class) {
		return ""
	} else {
		return class.String()
	}
}
