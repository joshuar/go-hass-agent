// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

// This list is taken from:
// https://developers.home-assistant.io/docs/core/entity/sensor/

//go:generate stringer -type=SensorDeviceClass -output deviceClassStrings.go -trimprefix Sensor
const (
	Apparent_power             SensorDeviceClass = iota + 1 // Apparent power
	Aqi                                                     // Air Quality Index
	Atmospheric_pressure                                    // Atmospheric pressure.
	SensorBattery                                           // Percentage of battery that is left
	Carbon_dioxide                                          // Concentration of carbon dioxide.
	Carbon_monoxide                                         // Concentration of carbon monoxide.
	Current                                                 // Current
	Data_rate                                               // Data rate
	Data_size                                               // Data size
	Date                                                    // Date. Requires native_value to be a Python datetime.date object, or None.
	Distance                                                // Generic distance
	Duration                                                // Time period. Should not update only due to time passing. The device or service needs to give a new data point to update.
	Energy                                                  // Energy, this device class should used for sensors representing energy consumption, for example an electricity meter. Represents power over time. Not to be confused with power.
	EnergyStorage                                           // Stored energy, this device class should be used for sensors representing stored energy, for example the amount of electric energy currently stored in a battery or the capacity of a battery. Represents power over time. Not to be confused with power.
	Enum                                                    // The sensor has a limited set of (non-numeric) states. The options property must be set to a list of possible states when using this device class.
	Frequency                                               // Frequency
	Gas                                                     // Volume of gas. Gas consumption measured as energy in kWh instead of a volume should be classified as energy.
	Humidity                                                // Relative humidity
	Illuminance                                             // Light level
	Irradiance                                              // Irradiance
	Moisture                                                // Moisture
	Monetary                                                // Monetary value with a currency.
	Nitrogen_dioxide                                        // Concentration of nitrogen dioxide
	Nitrogen_monoxide                                       // Concentration of nitrogen monoxide
	Nitrous_oxide                                           // Concentration of nitrous oxide
	Ozone                                                   // Concentration of ozone
	Pm1                                                     // Concentration of particulate matter less than 1 micrometer
	Pm25                                                    // Concentration of particulate matter less than 2.5 micrometers
	Pm10                                                    // Concentration of particulate matter less than 10 micrometers
	Power_factor                                            // Power Factor
	SensorPower                                             //  Power
	Precipitation                                           // Accumulated precipitation
	Precipitation_intensity                                 // Precipitation intensity
	Pressure                                                // Pressure
	Reactive_power                                          // Reactive power
	Signal_strength                                         // Signal strength
	Sound_pressure                                          // Sound pressure
	Speed                                                   // Generic speed
	Sulphur_dioxide                                         // Concentration of sulphure dioxide
	SensorTemperature                                       // Temperature
	Timestamp                                               // Timestamp. Requires native_value to return a Python datetime.datetime object, with time zone information, or None.
	Volatile_organic_compounds                              // Concentration of volatile organic compounds
	Voltage                                                 // Voltage
	Volume                                                  // Generic stored volume, this device class should be used for sensors representing a stored volume, for example the amount of fuel in a fuel tank.
	Water                                                   // Water consumption
	Weight                                                  // Generic mass; weight is used instead of mass to fit with every day language.
	Wind_speed                                              // Wind speed
)

// SensorDeviceClass reflects the HA device class of the sensor.
type SensorDeviceClass int
