// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package types

//go:generate stringer -type=StateClass -output stateclass_generated.go -linecomment
const (
	StateClassNone            StateClass = iota //
	StateClassMeasurement                       // measurement
	StateClassTotal                             // total
	StateClassTotalIncreasing                   // total_increasing
)

// SensorStateClass reflects the HA state class of the sensor.
type StateClass int
