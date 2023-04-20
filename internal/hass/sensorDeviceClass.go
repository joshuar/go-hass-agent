// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

//go:generate stringer -type=SensorDeviceClass -output sensorDeviceClassStrings.go -trimprefix Sensor

const (
	Apparent_power SensorDeviceClass = iota + 1
	Aqi
	Atmospheric_pressure
	SensorBattery
	Carbon_dioxide
	Carbon_monoxide
	Current
	Data_rate
	Data_size
	Date
	Distance
	Duration
	Energy
	Enum
	Frequency
	Gas
	Humidity
	Illuminance
	Irradiance
	Moisture
	Monetary
	Nitrogen_dioxide
	Nitrogen_monoxide
	Nitrous_oxide
	Ozone
	Pm1
	Pm25
	Pm10
	Power_factor
	SensorPower
	Precipitation
	Precipitation_intensity
	Pressure
	Reactive_power
	Signal_strength
	Sound_pressure
	Speed
	Sulphur_dioxide
	SensorTemperature
	Timestamp
	Volatile_organic_compounds
	Voltage
	Volume
	Water
	Weight
	Wind_speed
)

// SensorDeviceClass reflects the HA device class of the sensor.
type SensorDeviceClass int
