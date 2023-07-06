// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

// sensorState tracks the current state of a sensor, including the sensor value
// and whether it is registered/disabled in HA.
type sensorState struct {
	data        hass.Sensor
	disableCh   chan bool
	errCh       chan error
	requestData []byte
	requestType hass.RequestType
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

// sensorState implements hass.Request so its data can be sent to the HA API

func (sensor *sensorState) RequestType() hass.RequestType {
	return sensor.requestType
}

func (sensor *sensorState) RequestData() json.RawMessage {
	return sensor.requestData
}

func (sensor *sensorState) ResponseHandler(rawResponse bytes.Buffer) {
	defer close(sensor.disableCh)
	switch {
	case rawResponse.Len() == 0 || rawResponse.String() == "{}":
		sensor.errCh <- fmt.Errorf("no response for %s request. Likely problem with request data", sensor.Name())
		return
	default:
		var r interface{}
		err := json.Unmarshal(rawResponse.Bytes(), &r)
		if err != nil {
			sensor.errCh <- errors.New("could not unmarshal response")
			return
		}
		response := r.(map[string]interface{})
		if v, ok := response["success"]; ok {
			if v.(bool) {
				close(sensor.errCh)
			}
		}
		if v, ok := response[sensor.ID()]; ok {
			status := v.(map[string]interface{})
			if !status["success"].(bool) {
				hassErr := status["error"].(map[string]interface{})
				sensor.errCh <- fmt.Errorf("code %s: %s", hassErr["code"], hassErr["message"])
			} else {
				close(sensor.errCh)
			}
			if _, ok := status["is_disabled"]; ok {
				sensor.disableCh <- true
			} else {
				sensor.disableCh <- false
			}
		}
	}
}

func newSensorState(s hass.Sensor, r Registry) *sensorState {
	update := &sensorState{
		data:      s,
		disableCh: make(chan bool, 1),
		errCh:     make(chan error, 1),
	}

	var err error
	if r.IsRegistered(s.ID()) {
		update.requestData, err = json.Marshal(hass.MarshalSensorUpdate(s))
		update.requestType = hass.RequestTypeUpdateSensorStates
	} else {
		update.requestData, err = json.Marshal(hass.MarshalSensorRegistration(s))
		update.requestType = hass.RequestTypeRegisterSensor
	}
	if err != nil {
		log.Debug().Err(err).
			Msgf("Could not marshal sensor update for %s", s.ID())
	}
	return update
}
