// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package class

//go:generate go tool golang.org/x/tools/cmd/stringer -type=SensorStateClass -output stateclass.gen.go -linecomment
const (
	StateClassMin        SensorStateClass = iota //
	StateMeasurement                             // measurement
	StateTotal                                   // total
	StateTotalIncreasing                         // total_increasing
	StateClassMax                                //
)

// SensorStateClass reflects the HA state class of the sensor.
type SensorStateClass int

// Valid returns whether the SensorStateClass is valid.
func (c SensorStateClass) Valid() bool {
	return c > StateClassMin && c < StateClassMax
}
