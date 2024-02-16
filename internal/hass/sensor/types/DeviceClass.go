// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package types

// This list is taken from:
// https://developers.home-assistant.io/docs/core/entity/sensor/

//go:generate stringer -type=DeviceClass -output DeviceClass_generated.go -linecomment
const (
	DeviceClassApparentPower          DeviceClass = iota + 1 // Apparent Power
	DeviceClassAqi                                           // Air Quality Index
	DeviceClassAtmosphericPressure                           // Atmospheric Pressure
	DeviceClassBattery                                       // Battery Percent
	DeviceClassCarbonDioxide                                 // Carbon Dioxide Concentration.
	DeviceClassCarbonMonoxide                                // Carbon Monoxide Concentration
	DeviceClassCurrent                                       // Current
	DeviceClassDataRate                                      // Data Rate
	DeviceClassDataSize                                      // Data Size
	DeviceClassDate                                          // Date
	DeviceClassDistance                                      // Distance
	DeviceClassDuration                                      // Time Period
	DeviceClassEnergyStorage                                 // Stored Energy
	DeviceClassEnum                                          // Predefined State
	DeviceClassFrequency                                     // Frequency
	DeviceClassGas                                           // Gas Volume
	DeviceClassHumidity                                      // Relative Humidity
	DeviceClassIlluminance                                   // Light Level
	DeviceClassIrradiance                                    // Irradiance
	DeviceClassMoisture                                      // Moisture
	DeviceClassMonetary                                      // Monetary Value
	DeviceClassNitrogenDioxide                               // Nitrogen Dioxide Concentration
	DeviceClassNitrogenMonoxide                              // Nitrogen Monoxide Concentration
	DeviceClassNitrousOxide                                  // Nitrous Oxide Concentration
	DeviceClassOzone                                         // Ozone Concentration
	DeviceClassPm1                                           // PM1 Concentration
	DeviceClassPm25                                          // PM2.5 Concentration
	DeviceClassPm10                                          // PM10 Concentration
	DeviceClassPowerFactor                                   // Power Factor
	DeviceClassPower                                         // Power
	DeviceClassPrecipitation                                 // Accumulated Precipitation
	DeviceClassPrecipitationIntensity                        // Precipitation Intensity
	DeviceClassPressure                                      // Pressure
	DeviceClassReactivePower                                 // Reactive Power
	DeviceClassSignalStrength                                // Signal Strength
	DeviceClassSoundPressure                                 // Sound Pressure
	DeviceClassSpeed                                         // Speed
	DeviceClassSulphurDioxide                                // Sulphure Dioxide Concentration
	DeviceClassTemperature                                   // Temperature
	DeviceClassTimestamp                                     // Timestamp
	DeviceClassVOC                                           // VOC Concentration
	DeviceClassVoltage                                       // Voltage
	DeviceClassVolume                                        // Volume
	DeviceClassWater                                         // Water Consumption
	DeviceClassWeight                                        // Mass
	DeviceClassWindSpeed                                     // Wind speed
)

// SensorDeviceClass reflects the HA device class of the sensor.
type DeviceClass int
