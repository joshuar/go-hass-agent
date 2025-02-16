// Code generated by "stringer -type=SensorDeviceClass -output deviceclass.gen.go -linecomment"; DO NOT EDIT.

package class

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[SensorClassMin-0]
	_ = x[SensorClassApparentPower-1]
	_ = x[SensorClassAqi-2]
	_ = x[SensorClassAtmosphericPressure-3]
	_ = x[SensorClassBattery-4]
	_ = x[SensorClassCarbonDioxide-5]
	_ = x[SensorClassCarbonMonoxide-6]
	_ = x[SensorClassCurrent-7]
	_ = x[SensorClassDataRate-8]
	_ = x[SensorClassDataSize-9]
	_ = x[SensorClassDate-10]
	_ = x[SensorClassDistance-11]
	_ = x[SensorClassDuration-12]
	_ = x[SensorClassEnergyStorage-13]
	_ = x[SensorClassEnum-14]
	_ = x[SensorClassFrequency-15]
	_ = x[SensorClassGas-16]
	_ = x[SensorClassHumidity-17]
	_ = x[SensorClassIlluminance-18]
	_ = x[SensorClassIrradiance-19]
	_ = x[SensorClassMoisture-20]
	_ = x[SensorClassMonetary-21]
	_ = x[SensorClassNitrogenDioxide-22]
	_ = x[SensorClassNitrogenMonoxide-23]
	_ = x[SensorClassNitrousOxide-24]
	_ = x[SensorClassOzone-25]
	_ = x[SensorClassPm1-26]
	_ = x[SensorClassPm25-27]
	_ = x[SensorClassPm10-28]
	_ = x[SensorClassPowerFactor-29]
	_ = x[SensorClassPower-30]
	_ = x[SensorClassPrecipitation-31]
	_ = x[SensorClassPrecipitationIntensity-32]
	_ = x[SensorClassPressure-33]
	_ = x[SensorClassReactivePower-34]
	_ = x[SensorClassSignalStrength-35]
	_ = x[SensorClassSoundPressure-36]
	_ = x[SensorClassSpeed-37]
	_ = x[SensorClassSulphurDioxide-38]
	_ = x[SensorClassTemperature-39]
	_ = x[SensorClassTimestamp-40]
	_ = x[SensorClassVOC-41]
	_ = x[SensorClassVoltage-42]
	_ = x[SensorClassVolume-43]
	_ = x[SensorClassWater-44]
	_ = x[SensorClassWeight-45]
	_ = x[SensorClassWindSpeed-46]
	_ = x[SensorClassMax-47]
	_ = x[BinaryClassMin-48]
	_ = x[BinaryClassBattery-49]
	_ = x[BinaryClassBatteryCharging-50]
	_ = x[BinaryClassCO-51]
	_ = x[BinaryClassCold-52]
	_ = x[BinaryClassConnectivity-53]
	_ = x[BinaryClassDoor-54]
	_ = x[BinaryClassGarageDoor-55]
	_ = x[BinaryClassGas-56]
	_ = x[BinaryClassHeat-57]
	_ = x[BinaryClassLight-58]
	_ = x[BinaryClassLock-59]
	_ = x[BinaryClassMoisture-60]
	_ = x[BinaryClassMotion-61]
	_ = x[BinaryClassMoving-62]
	_ = x[BinaryClassOccupancy-63]
	_ = x[BinaryClassOpening-64]
	_ = x[BinaryClassPlug-65]
	_ = x[BinaryClassPower-66]
	_ = x[BinaryClassPresence-67]
	_ = x[BinaryClassProblem-68]
	_ = x[BinaryClassRunning-69]
	_ = x[BinaryClassSafety-70]
	_ = x[BinaryClassSmoke-71]
	_ = x[BinaryClassSound-72]
	_ = x[BinaryClassTamper-73]
	_ = x[BinaryClassUpdate-74]
	_ = x[BinaryClassVibration-75]
	_ = x[BinaryClassWindow-76]
	_ = x[BinaryClassMax-77]
}

const _SensorDeviceClass_name = "apparent_poweraqiatmospheric_pressurebatterycarbon_dioxidecarbon_monoxidecurrentdata_ratedata_sizedatedistancedurationenergy_storageenumfrequencygashumidityilluminanceirradiancemoisturemonetarynitrogen_dioxidenitrogen_monoxidenitrous_oxideozonepm1pm25pm10power_factorpowerprecipitationprecipitation_intensitypressurereactive_powersignal_strengthsound_pressurespeedsulphure_dioxidetemperaturetimestampvocvoltagevolumewaterweightwind_speedbatterybattery_chargingcarbon_monoxidecoldconnectivitydoorgarage_doorgasheatlightlockmoisturemotionmovingoccupancyopeningplugpowerpresenceproblemrunningsafetysmokesoundtamperupdatevibrationwindow"

var _SensorDeviceClass_index = [...]uint16{0, 0, 14, 17, 37, 44, 58, 73, 80, 89, 98, 102, 110, 118, 132, 136, 145, 148, 156, 167, 177, 185, 193, 209, 226, 239, 244, 247, 251, 255, 267, 272, 285, 308, 316, 330, 345, 359, 364, 380, 391, 400, 403, 410, 416, 421, 427, 437, 437, 437, 444, 460, 475, 479, 491, 495, 506, 509, 513, 518, 522, 530, 536, 542, 551, 558, 562, 567, 575, 582, 589, 595, 600, 605, 611, 617, 626, 632, 632}

func (i SensorDeviceClass) String() string {
	if i < 0 || i >= SensorDeviceClass(len(_SensorDeviceClass_index)-1) {
		return "SensorDeviceClass(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _SensorDeviceClass_name[_SensorDeviceClass_index[i]:_SensorDeviceClass_index[i+1]]
}
