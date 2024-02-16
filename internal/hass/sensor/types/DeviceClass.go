// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package types

// This list is taken from:
// https://developers.home-assistant.io/docs/core/entity/sensor/

//go:generate stringer -type=DeviceClass -output DeviceClass_generated.go -linecomment
const (
	DeviceClassApparentPower          DeviceClass = iota + 1 // apparent_power
	DeviceClassAqi                                           // aqi
	DeviceClassAtmosphericPressure                           // atmospheric_pressure
	DeviceClassBattery                                       // battery
	DeviceClassCarbonDioxide                                 // carbon_dioxide
	DeviceClassCarbonMonoxide                                // carbon_monoxide
	DeviceClassCurrent                                       // current
	DeviceClassDataRate                                      // data_rate
	DeviceClassDataSize                                      // data_size
	DeviceClassDate                                          // date
	DeviceClassDistance                                      // distance
	DeviceClassDuration                                      // duration
	DeviceClassEnergyStorage                                 // energy_storage
	DeviceClassEnum                                          // enum
	DeviceClassFrequency                                     // frequency
	DeviceClassGas                                           // gas
	DeviceClassHumidity                                      // humidity
	DeviceClassIlluminance                                   // illuminance
	DeviceClassIrradiance                                    // irradiance
	DeviceClassMoisture                                      // moisture
	DeviceClassMonetary                                      // monetary
	DeviceClassNitrogenDioxide                               // nitrogen_dioxide
	DeviceClassNitrogenMonoxide                              // nitrogen_monoxide
	DeviceClassNitrousOxide                                  // nitrous_oxide
	DeviceClassOzone                                         // ozone
	DeviceClassPm1                                           // pm1
	DeviceClassPm25                                          // pm25
	DeviceClassPm10                                          // pm10
	DeviceClassPowerFactor                                   // power_factor
	DeviceClassPower                                         // power
	DeviceClassPrecipitation                                 // precipitation
	DeviceClassPrecipitationIntensity                        // precipitation_intensity
	DeviceClassPressure                                      // pressure
	DeviceClassReactivePower                                 // reactive_power
	DeviceClassSignalStrength                                // signal_strength
	DeviceClassSoundPressure                                 // sound_pressure
	DeviceClassSpeed                                         // speed
	DeviceClassSulphurDioxide                                // sulphure_dioxide
	DeviceClassTemperature                                   // temperature
	DeviceClassTimestamp                                     // timestamp
	DeviceClassVOC                                           // voc
	DeviceClassVoltage                                       // voltage
	DeviceClassVolume                                        // volume
	DeviceClassWater                                         // water
	DeviceClassWeight                                        // weight
	DeviceClassWindSpeed                                     // wind_speed
)

// SensorDeviceClass reflects the HA device class of the sensor.
type DeviceClass int
