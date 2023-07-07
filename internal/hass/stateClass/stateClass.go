// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package stateClass

//go:generate stringer -type=SensorStateClass -output stateClassStrings.go -linecomment
const (
	StateMeasurement     SensorStateClass = iota + 1 // measurement
	StateTotal                                       // total
	StateTotalIncreasing                             // total_increasing
)

// SensorStateClass reflects the HA state class of the sensor.
type SensorStateClass int
