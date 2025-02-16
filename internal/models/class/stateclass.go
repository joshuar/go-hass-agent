// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package class

//go:generate go run golang.org/x/tools/cmd/stringer -type=SensorStateClass -output stateclass.gen.go -linecomment
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
