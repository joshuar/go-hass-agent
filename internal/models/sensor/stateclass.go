// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

//go:generate go run golang.org/x/tools/cmd/stringer -type=StateClass -output stateclass.gen.go -linecomment
const (
	StateNone            StateClass = iota //
	StateMeasurement                       // measurement
	StateTotal                             // total
	StateTotalIncreasing                   // total_increasing
)

// SensorStateClass reflects the HA state class of the sensor.
type StateClass int
