// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package types

//go:generate stringer -type=SensorType -linecomment -output sensortype_generated.go
const (
	Sensor       SensorType = iota // sensor
	BinarySensor                   // binary_sensor
)

// SensorType reflects the type of sensor, sensor or binary_sensor.
type SensorType int
