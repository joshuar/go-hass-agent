// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"fmt"
	"strings"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

// Sensor represents an update for a sensor. It reflects the current state
// of the sensor at the point in time it is used. It provides a bridge between
// platform/device and HA implementations of what a sensor is.
//
//go:generate moq -out mock_Sensor_test.go . Sensor
type Sensor interface {
	Name() string
	ID() string
	Icon() string
	SensorType() sensor.SensorType
	DeviceClass() sensor.SensorDeviceClass
	StateClass() sensor.SensorStateClass
	State() interface{}
	Units() string
	Category() string
	Attributes() interface{}
}

func prettyPrintState(s Sensor) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%v", s.State())
	if s.Units() != "" {
		fmt.Fprintf(&b, " %s", s.Units())
	}
	return b.String()
}

func marshalSensorUpdate(s Sensor) *sensor.SensorUpdateInfo {
	return &sensor.SensorUpdateInfo{
		StateAttributes: s.Attributes(),
		Icon:            s.Icon(),
		State:           s.State(),
		Type:            marshalClass(s.SensorType()),
		UniqueID:        s.ID(),
	}
}

func marshalSensorRegistration(s Sensor) *sensor.SensorRegistrationInfo {
	return &sensor.SensorRegistrationInfo{
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
