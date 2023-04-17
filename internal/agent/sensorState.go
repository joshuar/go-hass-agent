// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"

	"github.com/dgraph-io/badger/v4"
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
	switch {
	case rawResponse == nil:
		log.Debug().Caller().
			Msg("No response received. Likely failed request.")
	case len(rawResponse.(map[string]interface{})) == 0:
		log.Debug().Caller().
			Msg("No response data. Likely problem with request data.")
	default:
		response := rawResponse.(map[string]interface{})
		if v, ok := response["success"]; ok {
			if v.(bool) && !sensor.registered {
				sensor.updateRegistration()
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
				sensor.updateDisabled(true)
			}
		}
	}
}

// newSensor takes a hass.SensorUpdate sent by the platform/device and derives
// the information to encapsulate it as a sensorState.
func newSensor(newSensor hass.SensorUpdate) *sensorState {
	state, err := sensorRegistry.GetState(newSensor.ID())
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			log.Debug().
				Msgf("Adding %s to registry DB.", newSensor.Name())
			err := sensorRegistry.NewState(newSensor.ID())
			if err != nil {
				log.Debug().Err(err).
					Msgf("Could not add %s to registry DB.", newSensor.Name())
			}
		}
		log.Debug().Err(err).Msg("Could not retrieve state.")
	}
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
		registered:  state.Registered,
		disabled:    state.Disabled,
	}
	return sensor
}

// updateSensor ensures the bare minimum properties of a sensor are updated from
// a hass.SensorUpdate
func (sensor *sensorState) updateSensor(ctx context.Context, update hass.SensorUpdate) {
	sensor.state = update.State()
	sensor.attributes = update.Attributes()
	sensor.icon = update.Icon()
}

// updateRegistration updates the registration status of the sensor in the
// on-disk sensor registry for persistence/tracking between runs of
// go-hass-agent
func (sensor *sensorState) updateRegistration() {
	err := sensorRegistry.SetState(sensor.entityID, "registered", true)
	if err != nil {
		log.Debug().Err(err).
			Msgf("Could not store registration for sensor %s in DB.", sensor.name)
	} else {
		sensor.registered = true
		log.Debug().Caller().
			Msgf("Sensor %s registered with state %v", sensor.name, sensor.state)
	}
}

// updateDisabled updates the disabled status of the sensor in the on-disk
// sensor registry for persistence/tracking between runs of go-hass-agent.
func (sensor *sensorState) updateDisabled(value bool) {
	err := sensorRegistry.SetState(sensor.entityID, "disabled", value)
	if err != nil {
		log.Debug().Err(err).
			Msgf("Could not set disabled status in DB for sensor %s", sensor.name)
	} else {
		log.Debug().Caller().
			Msgf("Sensor %s disabled set to %v.", sensor.name, value)
		sensor.disabled = value
	}
}
