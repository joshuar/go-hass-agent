// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package types

//go:generate go run golang.org/x/tools/cmd/stringer -type=DeviceClass -output deviceclass_generated.go -linecomment
const (
	// For sensor entity device class descriptions, see:
	// https://developers.home-assistant.io/docs/core/entity/sensor#available-device-classes
	SensorDeviceClassNone                   DeviceClass = iota //
	SensorDeviceClassApparentPower                             // apparent_power
	SensorDeviceClassAqi                                       // aqi
	SensorDeviceClassAtmosphericPressure                       // atmospheric_pressure
	SensorDeviceClassBattery                                   // battery
	SensorDeviceClassCarbonDioxide                             // carbon_dioxide
	SensorDeviceClassCarbonMonoxide                            // carbon_monoxide
	SensorDeviceClassCurrent                                   // current
	SensorDeviceClassDataRate                                  // data_rate
	SensorDeviceClassDataSize                                  // data_size
	SensorDeviceClassDate                                      // date
	SensorDeviceClassDistance                                  // distance
	SensorDeviceClassDuration                                  // duration
	SensorDeviceClassEnergyStorage                             // energy_storage
	SensorDeviceClassEnum                                      // enum
	SensorDeviceClassFrequency                                 // frequency
	SensorDeviceClassGas                                       // gas
	SensorDeviceClassHumidity                                  // humidity
	SensorDeviceClassIlluminance                               // illuminance
	SensorDeviceClassIrradiance                                // irradiance
	SensorDeviceClassMoisture                                  // moisture
	SensorDeviceClassMonetary                                  // monetary
	SensorDeviceClassNitrogenDioxide                           // nitrogen_dioxide
	SensorDeviceClassNitrogenMonoxide                          // nitrogen_monoxide
	SensorDeviceClassNitrousOxide                              // nitrous_oxide
	SensorDeviceClassOzone                                     // ozone
	SensorDeviceClassPm1                                       // pm1
	SensorDeviceClassPm25                                      // pm25
	SensorDeviceClassPm10                                      // pm10
	SensorDeviceClassPowerFactor                               // power_factor
	SensorDeviceClassPower                                     // power
	SensorDeviceClassPrecipitation                             // precipitation
	SensorDeviceClassPrecipitationIntensity                    // precipitation_intensity
	SensorDeviceClassPressure                                  // pressure
	SensorDeviceClassReactivePower                             // reactive_power
	SensorDeviceClassSignalStrength                            // signal_strength
	SensorDeviceClassSoundPressure                             // sound_pressure
	SensorDeviceClassSpeed                                     // speed
	SensorDeviceClassSulphurDioxide                            // sulphure_dioxide
	SensorDeviceClassTemperature                               // temperature
	SensorDeviceClassTimestamp                                 // timestamp
	SensorDeviceClassVOC                                       // voc
	SensorDeviceClassVoltage                                   // voltage
	SensorDeviceClassVolume                                    // volume
	SensorDeviceClassWater                                     // water
	SensorDeviceClassWeight                                    // weight
	SensorDeviceClassWindSpeed                                 // wind_speed
	SensorDeviceClassMax
	// For binary sensor entity device class descriptions, see:
	// https://developers.home-assistant.io/docs/core/entity/binary-sensor#available-device-classes
	BinarySensorDeviceClassBattery         // battery
	BinarySensorDeviceClassBatteryCharging // battery_charging
	BinarySensorDeviceClassCO              // carbon_monoxide
	BinarySensorDeviceClassCold            // cold
	BinarySensorDeviceClassConnectivity    // connectivity
	BinarySensorDeviceClassDoor            // door
	BinarySensorDeviceClassGarageDoor      // garage_door
	BinarySensorDeviceClassGas             // gas
	BinarySensorDeviceClassHeat            // heat
	BinarySensorDeviceClassLight           // light
	BinarySensorDeviceClassLock            // lock
	BinarySensorDeviceClassMoisture        // moisture
	BinarySensorDeviceClassMotion          // motion
	BinarySensorDeviceClassMoving          // moving
	BinarySensorDeviceClassOccupancy       // occupancy
	BinarySensorDeviceClassOpening         // opening
	BinarySensorDeviceClassPlug            // plug
	BinarySensorDeviceClassPower           // power
	BinarySensorDeviceClassPresence        // presence
	BinarySensorDeviceClassProblem         // problem
	BinarySensorDeviceClassRunning         // running
	BinarySensorDeviceClassSafety          // safety
	BinarySensorDeviceClassSmoke           // smoke
	BinarySensorDeviceClassSound           // sound
	BinarySensorDeviceClassTamper          // tamper
	BinarySensorDeviceClassUpdate          // update
	BinarySensorDeviceClassVibration       // vibration
	BinarySensorDeviceClassWindow          // window
	BinarySensorDeviceClassMax             //
)

// DeviceClass represents the device class of a sensor or binary sensor. It is
// an extra classification of what the entity represents, and will potentially
// enforce display and unit restrictions in Home Assistant.
type DeviceClass int
