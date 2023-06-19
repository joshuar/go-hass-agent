// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

// sensorState tracks the current state of a sensor, including the sensor value
// and whether it is registered/disabled in HA.
type sensorState struct {
	data     hass.SensorUpdate
	metadata *sensorMetadata
}

type sensorMetadata struct {
	Registered bool `json:"Registered"`
	Disabled   bool `json:"Disabled"`
}

// sensorState implements hass.Sensor to represent a sensor in HA.

func (s *sensorState) DeviceClass() hass.SensorDeviceClass {
	return s.data.DeviceClass()
}

func (s *sensorState) StateClass() hass.SensorStateClass {
	return s.data.StateClass()
}

func (s *sensorState) SensorType() hass.SensorType {
	return s.data.SensorType()
}

func (s *sensorState) Icon() string {
	return s.data.Icon()
}

func (s *sensorState) Name() string {
	return s.data.Name()
}

func (s *sensorState) State() interface{} {
	if s.data.State() != nil {
		return s.data.State()
	} else {
		return "Unknown"
	}
}

func (s *sensorState) Attributes() interface{} {
	return s.data.Attributes()
}

func (s *sensorState) ID() string {
	return s.data.ID()
}

func (s *sensorState) Units() string {
	return s.data.Units()
}

func (s *sensorState) Category() string {
	return s.data.Category()
}

func (s *sensorState) Disabled() bool {
	if s.metadata != nil {
		return s.metadata.Disabled
	} else {
		return false
	}
}

func (s *sensorState) Registered() bool {
	if s.metadata != nil {
		return s.metadata.Registered
	} else {
		return false
	}
}

func (s *sensorState) MarshalJSON() ([]byte, error) {
	if s.Registered() {
		m := hass.MarshalSensorUpdate(s)
		return json.Marshal(m)
	} else {
		m := hass.MarshalSensorRegistration(s)
		return json.Marshal(m)
	}
}

func (s *sensorState) UnMarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s)
}

// sensorState implements hass.Request so its data can be sent to the HA API

func (sensor *sensorState) RequestType() hass.RequestType {
	if sensor.metadata.Registered {
		return hass.RequestTypeUpdateSensorStates
	}
	return hass.RequestTypeRegisterSensor
}

func (sensor *sensorState) RequestData() *json.RawMessage {
	data, _ := sensor.MarshalJSON()
	raw := json.RawMessage(data)
	return &raw
}

func (sensor *sensorState) ResponseHandler(rawResponse bytes.Buffer) {
	switch {
	case rawResponse.Len() == 0 || rawResponse.String() == "{}":
		log.Debug().Caller().
			Msgf("No response for %s request. Likely problem with request data.", sensor.Name())
	default:
		var r interface{}
		err := json.Unmarshal(rawResponse.Bytes(), &r)
		if err != nil {
			log.Debug().Caller().Err(err).
				Msg("Could not unmarshal response.")
			return
		}
		response := r.(map[string]interface{})
		if v, ok := response["success"]; ok {
			if v.(bool) && !sensor.metadata.Registered {
				sensor.metadata.Registered = true
				log.Debug().Caller().
					Msgf("Sensor %s registered in HA.",
						sensor.Name())
			}
		}
		if v, ok := response[sensor.ID()]; ok {
			status := v.(map[string]interface{})
			if !status["success"].(bool) {
				hassErr := status["error"].(map[string]interface{})
				err := fmt.Errorf("%s: %s", hassErr["code"], hassErr["message"])
				log.Debug().Caller().Err(err).
					Msgf("Could not update sensor %s.", sensor.Name())
			} else {
				log.Debug().Caller().
					Msgf("Sensor %s updated (%s). State is now: %v %s",
						sensor.Name(),
						sensor.ID(),
						sensor.State(),
						sensor.Units())
			}
			if _, ok := status["is_disabled"]; ok {
				sensor.metadata.Disabled = true
			} else if sensor.metadata.Disabled {
				sensor.metadata.Disabled = false
			}
		}
	}
}
