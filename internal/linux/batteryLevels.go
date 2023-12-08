// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

//go:generate stringer -type=batteryLevel -output batteryLevelsStrings.go -linecomment
const (
	batteryLevelNone batteryLevel = iota + 1 // None
	batteryLevelLow                          // Low
	batteryLevelCrit                         // Critical
	batteryLevelNorm                         // Normal
	batteryLevelHigh                         // High
	batteryLevelFull                         // Full
)

type batteryLevel uint32
