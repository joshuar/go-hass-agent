// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package battery

//go:generate go run golang.org/x/tools/cmd/stringer -type=battChargeState -output batteryStatesStrings.go -linecomment
const (
	stateUnknown          battChargeState = iota // Unknown
	stateCharging                                // Charging
	stateDischarging                             // Discharging
	stateEmpty                                   // Empty
	stateFullyCharged                            // Fully Charged
	statePendingCharge                           // Pending Charge
	statePendingDischarge                        // Pending Discharge
)

type battChargeState uint32
