// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package battery

//go:generate stringer -type=batteryType -output batteryTypeStrings.go -linecomment
const (
	batteryTypeLinePower batteryType = iota + 1 // Line Power
	batteryTypeBattery                          // Battery
	batteryTypeUps                              // UPS
	batteryTypeMonitor                          // Monitor
	batteryTypeMouse                            // Mouse
	batteryTypeKeyboard                         // Keyboard
	batteryTypePda                              // Pda
	batteryTypePhone                            // Phone
)

type batteryType uint32
