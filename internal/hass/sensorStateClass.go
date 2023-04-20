// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

//go:generate stringer -type=SensorStateClass -output sensorStateClassStrings.go -trimprefix State
const (
	StateMeasurement SensorStateClass = iota + 1
	StateTotal
	StateTotalIncreasing
)

// SensorStateClass reflects the HA state class of the sensor.
type SensorStateClass int
