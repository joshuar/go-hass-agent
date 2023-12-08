// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

//go:generate stringer -type=battChargeState -output batteryStatesStrings.go -linecomment
const (
	stateCharging         battChargeState = iota + 1 // Charging
	stateDischarging                                 // Discharging
	stateEmpty                                       // Empty
	stateFullyCharged                                // Fully Charged
	statePendingCharge                               // Pending Charge
	statePendingDischarge                            // Pending Discharge
)

type battChargeState uint32
