// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package types

//go:generate stringer -type=SensorClass -linecomment -output SensorClass_generated.go
const (
	Sensor       SensorClass = iota + 1 // sensor
	BinarySensor                        // binary_sensor
)

// SensorClass reflects the type of sensor, sensor or binary_sensor.
type SensorClass int
