// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

//go:generate stringer -type=SensorType -linecomment -output sensorTypeStrings.go
const (
	TypeSensor SensorType = iota + 1 // sensor
	TypeBinary                       // binary_sensor
)

// SensorType reflects the type of sensor, sensor or binary_sensor.
type SensorType int
