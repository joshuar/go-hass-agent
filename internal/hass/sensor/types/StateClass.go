// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package types

//go:generate go run golang.org/x/tools/cmd/stringer -type=StateClass -output StatedClass_generated.go -linecomment
const (
	StateClassMeasurement     StateClass = iota + 1 // measurement
	StateClassTotal                                 // total
	StateClassTotalIncreasing                       // total_increasing
)

// SensorStateClass reflects the HA state class of the sensor.
type StateClass int
