// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"bytes"
	"encoding/json"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

// sensorState tracks the current state of a sensor, including the sensor value
// and whether it is registered/disabled in HA.
type sensorState struct {
	deviceClass hass.SensorDeviceClass
	stateClass  hass.SensorStateClass
	sensorType  hass.SensorType
	state       interface{}
	stateUnits  string
	attributes  interface{}
	icon        string
	name        string
	entityID    string
	category    string
	metadata    *registryEntry
}

// sensorState implements hass.Sensor to represent a sensor in HA.
func (s *sensorState) Attributes() interface{} {
	return s.attributes
}

func (s *sensorState) DeviceClass() string {
	if s.deviceClass != 0 {
		return s.deviceClass.String()
	} else {
		return ""
	}
}

func (s *sensorState) Icon() string {
	return s.icon
}

func (s *sensorState) Name() string {
	return s.name
}

func (s *sensorState) State() interface{} {
	return s.state
}

func (s *sensorState) Type() string {
	if s.sensorType != 0 {
		return s.sensorType.String()
	} else {
		return ""
	}
	// switch s.sensorType {
	// case hass.TypeSensor:
	// 	return "sensor"
	// case hass.TypeBinary:
	// 	return "binary_sensor"
	// default:
	// 	log.Debug().Caller().Msgf("Invalid or unknown sensor type %v", s.sensorType)
	// 	return ""
	// }
}

func (s *sensorState) UniqueID() string {
	return s.entityID
}

func (s *sensorState) UnitOfMeasurement() string {
	return s.stateUnits
}

func (s *sensorState) StateClass() string {
	if s.stateClass != 0 {
		return strcase.ToCamel(s.stateClass.String())
	} else {
		return ""
	}
}

func (s *sensorState) EntityCategory() string {
	return s.category
}

func (s *sensorState) Disabled() bool {
	return s.metadata.IsDisabled()
}

func (s *sensorState) Registered() bool {
	return s.metadata.IsRegistered()
}

// sensorState implements hass.Request so its data can be sent to the HA API

func (sensor *sensorState) RequestType() hass.RequestType {
	if sensor.metadata.IsRegistered() {
		return hass.RequestTypeUpdateSensorStates
	}
	return hass.RequestTypeRegisterSensor
}

func (sensor *sensorState) RequestData() interface{} {
	return hass.MarshalSensorData(sensor)
}

func (sensor *sensorState) ResponseHandler(rawResponse bytes.Buffer) {
	switch {
	case rawResponse.Len() == 0:
		log.Debug().Caller().
			Msg("No response data. Likely problem with request data.")
	default:
		var r interface{}
		json.Unmarshal(rawResponse.Bytes(), &r)
		response := r.(map[string]interface{})
		if v, ok := response["success"]; ok {
			if v.(bool) && !sensor.metadata.IsRegistered() {
				sensor.metadata.SetRegistered(true)
				log.Debug().Caller().
					Msgf("Sensor %s registered in HA.", sensor.name)
			}
		}
		if v, ok := response[sensor.entityID]; ok {
			status := v.(map[string]interface{})
			if !status["success"].(bool) {
				error := status["error"].(map[string]interface{})
				log.Debug().Caller().
					Msgf("Could not update sensor %s, %s: %s",
						sensor.name, error["code"], error["message"])
			} else {
				log.Debug().Caller().
					Msgf("Sensor %s updated. State is now: %v",
						sensor.name, sensor.state)
			}
			if _, ok := status["is_disabled"]; ok {
				sensor.metadata.SetDisabled(true)
			} else if sensor.metadata.IsDisabled() {
				sensor.metadata.SetDisabled(false)
			}
		}
	}
}
