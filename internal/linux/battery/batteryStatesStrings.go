// Code generated by "stringer -type=battChargeState -output batteryStatesStrings.go -linecomment"; DO NOT EDIT.

package battery

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[stateUnknown-0]
	_ = x[stateCharging-1]
	_ = x[stateDischarging-2]
	_ = x[stateEmpty-3]
	_ = x[stateFullyCharged-4]
	_ = x[statePendingCharge-5]
	_ = x[statePendingDischarge-6]
}

const _battChargeState_name = "UnknownChargingDischargingEmptyFully ChargedPending ChargePending Discharge"

var _battChargeState_index = [...]uint8{0, 7, 15, 26, 31, 44, 58, 75}

func (i battChargeState) String() string {
	if i >= battChargeState(len(_battChargeState_index)-1) {
		return "battChargeState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _battChargeState_name[_battChargeState_index[i]:_battChargeState_index[i+1]]
}
