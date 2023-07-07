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

	"github.com/joshuar/go-hass-agent/internal/request"
	"github.com/rs/zerolog/log"
)

// sensorState tracks the current state of a sensor, including the sensor value
// and whether it is registered/disabled in HA.
type sensorState struct {
	data        Sensor
	disableCh   chan bool
	errCh       chan error
	requestData []byte
	requestType request.RequestType
}

func (s *sensorState) State() interface{} {
	if s.data.State() != nil {
		return s.data.State()
	} else {
		return "Unknown"
	}
}

func (s *sensorState) Units() string {
	return s.data.Units()
}

// sensorState implements hass.Request so its data can be sent to the HA API

func (sensor *sensorState) RequestType() request.RequestType {
	return sensor.requestType
}

func (sensor *sensorState) RequestData() json.RawMessage {
	return sensor.requestData
}

func (sensor *sensorState) ResponseHandler(rawResponse bytes.Buffer) {
	defer close(sensor.disableCh)
	if rawResponse.Len() == 0 || rawResponse.String() == "{}" {
		sensor.errCh <- fmt.Errorf("no response for %s request. Likely problem with request data", sensor.data.Name())
		return
	}
	var r interface{}
	err := json.Unmarshal(rawResponse.Bytes(), &r)
	if err != nil {
		sensor.errCh <- errors.New("could not unmarshal response")
		return
	}
	response, err := assertAs[map[string]interface{}](r)
	if err != nil {
		sensor.errCh <- err
		return
	}
	if v, ok := response["success"]; ok {
		success, err := assertAs[bool](v)
		if err != nil {
			sensor.errCh <- nil
		} else {
			if success {
				close(sensor.errCh)
			} else {
				sensor.errCh <- errors.New("unsuccessful update")
			}
		}
	}
	if v, ok := response[sensor.data.ID()]; ok {
		status, err := assertAs[map[string]interface{}](v)
		if err != nil {
			sensor.errCh <- err
			return
		}
		success, err := assertAs[bool](status["success"])
		if err != nil {
			sensor.errCh <- err
			return
		} else {
			if !success {
				hassErr, err := assertAs[map[string]interface{}](status["error"])
				if err != nil {
					sensor.errCh <- errors.New("unknown error")
				} else {
					sensor.errCh <- fmt.Errorf("code %s: %s", hassErr["code"], hassErr["message"])
				}
			} else {
				close(sensor.errCh)
			}
		}
		if _, ok := status["is_disabled"]; ok {
			sensor.disableCh <- true
		} else {
			sensor.disableCh <- false
		}
	}
}

func newSensorState(s Sensor, r Registry) *sensorState {
	update := &sensorState{
		data:      s,
		disableCh: make(chan bool, 1),
		errCh:     make(chan error, 1),
	}

	var err error
	if r.IsRegistered(s.ID()) {
		update.requestData, err = json.Marshal(MarshalSensorUpdate(s))
		update.requestType = request.RequestTypeUpdateSensorStates
	} else {
		update.requestData, err = json.Marshal(MarshalSensorRegistration(s))
		update.requestType = request.RequestTypeRegisterSensor
	}
	if err != nil {
		log.Debug().Err(err).
			Msgf("Could not marshal sensor update for %s", s.ID())
	}
	return update
}

func assertAs[T any](thing interface{}) (T, error) {
	if asT, ok := thing.(T); !ok {
		return *new(T), errors.New("could not assert value")
	} else {
		return asT, nil
	}

}
