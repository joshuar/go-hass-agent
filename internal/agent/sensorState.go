package agent

//go:generate stringer -type=deviceClassType,stateClassType -output sensorState_types.go -trimprefix deviceClass

const (
	apparent_power deviceClassType = iota + 1
	aqi
	atmospheric_pressure
	deviceClassBattery
	carbon_dioxide
	carbon_monoxide
	current
	data_rate
	data_size
	date
	distance
	duration
	energy
	enum
	frequency
	gas
	humidity
	illuminance
	irradiance
	moisture
	monetary
	nitrogen_dioxide
	nitrogen_monoxide
	nitrous_oxide
	ozone
	pm1
	pm25
	pm10
	power_factor
	deviceClassPower
	precipitation
	precipitation_intensity
	pressure
	reactive_power
	signal_strength
	sound_pressure
	speed
	sulphur_dioxide
	deviceClassTemperature
	timestamp
	volatile_organic_compounds
	voltage
	volume
	water
	weight
	wind_speed

	measurement stateClassType = iota
	total
	total_increasing
)

type deviceClassType int
type stateClassType int

type sensorState struct {
	deviceClass deviceClassType
	stateClass  stateClassType
	state       interface{}
	attributes  interface{}
	name        string
	entityID    string
	disabled    bool
	registered  bool
}

type sensorUpdate interface {
	Device() string
	Type() string
	Value() interface{}
	ExtraValues() interface{}
}
