// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go run golang.org/x/tools/cmd/stringer -type=sensorType,level,chargingState,typeDescription -output types_generated.go -linecomment
package battery

const (
	typeDesc       sensorType = iota // Battery Type
	typePercentage                   // Battery Level
	typeTemp                         // Battery Temperature
	typeVoltage                      // Battery Voltage
	typeEnergy                       // Battery Energy
	typeEnergyRate                   // Battery Power
	typeState                        // Battery State
	typeNativePath                   // Battery Path
	typeLevel                        // Battery Level
	typeModel                        // Battery Model
)

// sensorType is the type of sensor for a battery (e.g., battery level, state,
// power, etc.).
type sensorType int

const (
	levelUnknown level = iota // Unknown
	levelNone                 // None
	_
	levelLow  // Low
	levelCrit // Critical
	_
	levelNorm // Normal
	levelHigh // High
	levelFull // Full
)

// level is a description of the approximate charge level of a battery.
type level uint32

const (
	stateUnknown          chargingState = iota // Unknown
	stateCharging                              // Charging
	stateDischarging                           // Discharging
	stateEmpty                                 // Empty
	stateFullyCharged                          // Fully Charged
	statePendingCharge                         // Pending Charge
	statePendingDischarge                      // Pending Discharge
)

// chargingState is a description of the current charging state of a battery.
type chargingState uint32

const (
	linePowerType typeDescription = iota + 1 // Line Power
	batteryType                              // Battery
	upsType                                  // UPS
	monitorType                              // Monitor
	mouseType                                // Mouse
	keyboardType                             // Keyboard
	pdaType                                  // Pda
	phoneType                                // Phone
)

// typeDescription is a description of what kind of battery a battery is (e.g.,
// UPS, Phone, Line Power, etc.)
type typeDescription uint32
