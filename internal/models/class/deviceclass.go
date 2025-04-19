// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package class

//go:generate go tool golang.org/x/tools/cmd/stringer -type=SensorDeviceClass -output deviceclass.gen.go -linecomment
const (
	// For sensor entity device class descriptions, see:
	// https://developers.home-assistant.io/docs/core/entity/sensor#available-device-classes
	SensorClassMin                    SensorDeviceClass = iota //
	SensorClassApparentPower                                   // apparent_power
	SensorClassAqi                                             // aqi
	SensorClassAtmosphericPressure                             // atmospheric_pressure
	SensorClassBattery                                         // battery
	SensorClassCarbonDioxide                                   // carbon_dioxide
	SensorClassCarbonMonoxide                                  // carbon_monoxide
	SensorClassCurrent                                         // current
	SensorClassDataRate                                        // data_rate
	SensorClassDataSize                                        // data_size
	SensorClassDate                                            // date
	SensorClassDistance                                        // distance
	SensorClassDuration                                        // duration
	SensorClassEnergyStorage                                   // energy_storage
	SensorClassEnum                                            // enum
	SensorClassFrequency                                       // frequency
	SensorClassGas                                             // gas
	SensorClassHumidity                                        // humidity
	SensorClassIlluminance                                     // illuminance
	SensorClassIrradiance                                      // irradiance
	SensorClassMoisture                                        // moisture
	SensorClassMonetary                                        // monetary
	SensorClassNitrogenDioxide                                 // nitrogen_dioxide
	SensorClassNitrogenMonoxide                                // nitrogen_monoxide
	SensorClassNitrousOxide                                    // nitrous_oxide
	SensorClassOzone                                           // ozone
	SensorClassPm1                                             // pm1
	SensorClassPm25                                            // pm25
	SensorClassPm10                                            // pm10
	SensorClassPowerFactor                                     // power_factor
	SensorClassPower                                           // power
	SensorClassPrecipitation                                   // precipitation
	SensorClassPrecipitationIntensity                          // precipitation_intensity
	SensorClassPressure                                        // pressure
	SensorClassReactivePower                                   // reactive_power
	SensorClassSignalStrength                                  // signal_strength
	SensorClassSoundPressure                                   // sound_pressure
	SensorClassSpeed                                           // speed
	SensorClassSulphurDioxide                                  // sulphure_dioxide
	SensorClassTemperature                                     // temperature
	SensorClassTimestamp                                       // timestamp
	SensorClassVOC                                             // voc
	SensorClassVoltage                                         // voltage
	SensorClassVolume                                          // volume
	SensorClassWater                                           // water
	SensorClassWeight                                          // weight
	SensorClassWindSpeed                                       // wind_speed
	SensorClassMax                                             //
	// For binary sensor entity device class descriptions, see:
	//
	// https://developers.home-assistant.io/docs/core/entity/binary-sensor#available-device-classes
	BinaryClassMin             //
	BinaryClassBattery         // battery
	BinaryClassBatteryCharging // battery_charging
	BinaryClassCO              // carbon_monoxide
	BinaryClassCold            // cold
	BinaryClassConnectivity    // connectivity
	BinaryClassDoor            // door
	BinaryClassGarageDoor      // garage_door
	BinaryClassGas             // gas
	BinaryClassHeat            // heat
	BinaryClassLight           // light
	BinaryClassLock            // lock
	BinaryClassMoisture        // moisture
	BinaryClassMotion          // motion
	BinaryClassMoving          // moving
	BinaryClassOccupancy       // occupancy
	BinaryClassOpening         // opening
	BinaryClassPlug            // plug
	BinaryClassPower           // power
	BinaryClassPresence        // presence
	BinaryClassProblem         // problem
	BinaryClassRunning         // running
	BinaryClassSafety          // safety
	BinaryClassSmoke           // smoke
	BinaryClassSound           // sound
	BinaryClassTamper          // tamper
	BinaryClassUpdate          // update
	BinaryClassVibration       // vibration
	BinaryClassWindow          // window
	BinaryClassMax             //
)

// DeviceClass represents the device class of a sensor or binary sensor. It is
// an extra classification of what the entity represents, and will potentially
// enforce display and unit restrictions in Home Assistant.
type SensorDeviceClass int

// Valid returns whether the SensorDeviceClass is a valid value.
func (c SensorDeviceClass) Valid() bool {
	return c > SensorClassMin && c != SensorClassMax && c != BinaryClassMin && c < BinaryClassMax
}
