// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package types

// This list is taken from:
// https://developers.home-assistant.io/docs/core/entity/sensor/

//go:generate stringer -type=DeviceClass,BinarySensorDeviceClass -output DeviceClass_generated.go -linecomment
const (
	DeviceClassNone                   DeviceClass = iota //
	DeviceClassApparentPower                             // apparent_power
	DeviceClassAqi                                       // aqi
	DeviceClassAtmosphericPressure                       // atmospheric_pressure
	DeviceClassBattery                                   // battery
	DeviceClassCarbonDioxide                             // carbon_dioxide
	DeviceClassCarbonMonoxide                            // carbon_monoxide
	DeviceClassCurrent                                   // current
	DeviceClassDataRate                                  // data_rate
	DeviceClassDataSize                                  // data_size
	DeviceClassDate                                      // date
	DeviceClassDistance                                  // distance
	DeviceClassDuration                                  // duration
	DeviceClassEnergyStorage                             // energy_storage
	DeviceClassEnum                                      // enum
	DeviceClassFrequency                                 // frequency
	DeviceClassGas                                       // gas
	DeviceClassHumidity                                  // humidity
	DeviceClassIlluminance                               // illuminance
	DeviceClassIrradiance                                // irradiance
	DeviceClassMoisture                                  // moisture
	DeviceClassMonetary                                  // monetary
	DeviceClassNitrogenDioxide                           // nitrogen_dioxide
	DeviceClassNitrogenMonoxide                          // nitrogen_monoxide
	DeviceClassNitrousOxide                              // nitrous_oxide
	DeviceClassOzone                                     // ozone
	DeviceClassPm1                                       // pm1
	DeviceClassPm25                                      // pm25
	DeviceClassPm10                                      // pm10
	DeviceClassPowerFactor                               // power_factor
	DeviceClassPower                                     // power
	DeviceClassPrecipitation                             // precipitation
	DeviceClassPrecipitationIntensity                    // precipitation_intensity
	DeviceClassPressure                                  // pressure
	DeviceClassReactivePower                             // reactive_power
	DeviceClassSignalStrength                            // signal_strength
	DeviceClassSoundPressure                             // sound_pressure
	DeviceClassSpeed                                     // speed
	DeviceClassSulphurDioxide                            // sulphure_dioxide
	DeviceClassTemperature                               // temperature
	DeviceClassTimestamp                                 // timestamp
	DeviceClassVOC                                       // voc
	DeviceClassVoltage                                   // voltage
	DeviceClassVolume                                    // volume
	DeviceClassWater                                     // water
	DeviceClassWeight                                    // weight
	DeviceClassWindSpeed                                 // wind_speed
)

// SensorDeviceClass reflects the HA device class of the sensor.
type DeviceClass int

const (
	//	On means low, Off means normal.
	BinarySensorDeviceClassBattery BinarySensorDeviceClass = iota // battery
	// On means charging, Off means not charging.
	BinarySensorDeviceClassBatteryCharging // battery_charging
	// On means carbon monoxide detected, Off means no carbon monoxide (clear).
	BinarySensorDeviceClassCO // carbon_monoxide
	// On means cold, Off means normal.
	BinarySensorDeviceClassCold // cold
	// On means connected, Off means disconnected.
	BinarySensorDeviceClassConnectivity // connectivity
	// On means open, Off means closed.
	BinarySensorDeviceClassDoor // door
	// On means open, Off means closed.
	BinarySensorDeviceClassGarageDoor
	// On means gas detected, Off means no gas (clear).
	BinarySensorDeviceClassGas // gas
	// On means hot, Off means normal.
	BinarySensorDeviceClassHeat // heat
	// On means light detected, Off means no light.
	BinarySensorDeviceClassLight // light
	// On means open (unlocked), Off means closed (locked).
	BinarySensorDeviceClassLock // lock
	// On means wet, Off means dry.
	BinarySensorDeviceClassMoisture // moisture
	// On means motion detected, Off means no motion (clear).
	BinarySensorDeviceClassMotion // motion
	// On means moving, Off means not moving (stopped).
	BinarySensorDeviceClassMoving // moving
	// On means occupied, Off means not occupied (clear).
	BinarySensorDeviceClassOccupancy // occupancy
	// On means open, Off means closed.
	BinarySensorDeviceClassOpening // opening
	// On means plugged in, Off means unplugged.
	BinarySensorDeviceClassPlug // plug
	// On means power detected, Off means no power.
	BinarySensorDeviceClassPower // power
	// On means home, Off means away.
	BinarySensorDeviceClassPresence // presence
	// On means problem detected, Off means no problem (OK).
	BinarySensorDeviceClassProblem // problem
	// On means running, Off means not running.
	BinarySensorDeviceClassRunning // running
	// On means unsafe, Off means safe.
	BinarySensorDeviceClassSafety // safety
	// On means smoke detected, Off means no smoke (clear).
	BinarySensorDeviceClassSmoke // smoke
	// On means sound detected, Off means no sound (clear).
	BinarySensorDeviceClassSound // sound
	// On means tampering detected, Off means no tampering (clear).
	BinarySensorDeviceClassTamper // tamper
	// On means update available, Off means up-to-date.
	BinarySensorDeviceClassUpdate // update
	// On means vibration detected, Off means no vibration.
	BinarySensorDeviceClassVibration // vibration
	// On means open, Off means closed.
	BinarySensorDeviceClassWindow // window
)

type BinarySensorDeviceClass int
