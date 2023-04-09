package agent

import (
	"context"

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
	disabled    bool
	registered  bool
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
	switch s.sensorType {
	case hass.TypeSensor:
		return "sensor"
	case hass.TypeBinary:
		return "binary_sensor"
	default:
		log.Debug().Caller().Msgf("Invalid or unknown sensor type %v", s.sensorType)
		return ""
	}
}

func (s *sensorState) UniqueID() string {
	return s.entityID
}

func (s *sensorState) UnitOfMeasurement() string {
	return s.stateUnits
}

func (s *sensorState) StateClass() string {
	if s.stateClass != 0 {
		return s.stateClass.String()
	} else {
		return ""
	}
}

func (s *sensorState) EntityCategory() string {
	return s.category
}

func (s *sensorState) Disabled() bool {
	return s.disabled
}

func (s *sensorState) Registered() bool {
	return s.registered
}

// sensorState implements hass.Request so its data can be sent to the HA API

func (sensor *sensorState) RequestType() hass.RequestType {
	if sensor.registered {
		return hass.RequestTypeUpdateSensorStates
	}
	return hass.RequestTypeRegisterSensor
}

func (sensor *sensorState) RequestData() interface{} {
	return hass.MarshalSensorData(sensor)
}

func (sensor *sensorState) ResponseHandler(rawResponse interface{}) {
	if rawResponse == nil || len(rawResponse.(map[string]interface{})) == 0 {
		log.Debug().Caller().
			Msg("No response data.")
	} else {
		response := rawResponse.(map[string]interface{})
		if v, ok := response["success"]; ok {
			if v.(bool) && !sensor.registered {
				sensor.registered = true
				log.Debug().Caller().
					Msgf("Sensor %s registered.", sensor.name)
			}
		}
		if v, ok := response[sensor.entityID]; ok {
			status := v.(map[string]interface{})
			if !status["success"].(bool) {
				error := status["error"].(map[string]interface{})
				log.Error().Msgf("Could not update sensor %s, %s: %s", sensor.name, error["code"], error["message"])
			} else {
				log.Debug().Caller().
					Msgf("Sensor %s updated. State is now: %v", sensor.name, sensor.state)
			}
			if v, ok := status["is_disabled"]; ok {
				switch v.(bool) {
				case true:
					log.Debug().Caller().
						Msgf("Sensor %s has been disabled.", sensor.name)
					sensor.disabled = true
				case false:
					log.Debug().Caller().
						Msgf("Sensor %s has been enabled.", sensor.name)
					sensor.disabled = false
				}
			}
		}
	}
}

// newSensor takes a hass.SensorUpdate sent by the platform/device and derives
// the information to encapsulate it as a sensorState.
func newSensor(newSensor hass.SensorUpdate) *sensorState {
	sensor := &sensorState{
		entityID:    newSensor.ID(),
		name:        newSensor.Name(),
		deviceClass: newSensor.DeviceClass(),
		stateClass:  newSensor.StateClass(),
		sensorType:  newSensor.SensorType(),
		state:       newSensor.State(),
		attributes:  newSensor.Attributes(),
		icon:        newSensor.Icon(),
		stateUnits:  newSensor.Units(),
		category:    newSensor.Category(),
		registered:  false,
		disabled:    false,
	}
	return sensor
}

// updateSensor ensures the bare minimum properties of a sensor are updated from
// a hass.SensorUpdate
func (sensor *sensorState) updateSensor(ctx context.Context, update hass.SensorUpdate) {
	sensor.state = update.State()
	sensor.attributes = update.Attributes()
	sensor.icon = update.Icon()
	go hass.APIRequest(ctx, sensor)
}
