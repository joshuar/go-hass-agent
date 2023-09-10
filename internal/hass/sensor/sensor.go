// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/rs/zerolog/log"
)

const (
	STATE_UNKNOWN = "unknown"
)

// SensorRegistrationInfo is the JSON structure required to register a sensor
// with HA.
type SensorRegistrationInfo struct {
	State             interface{} `json:"state"`
	StateAttributes   interface{} `json:"attributes,omitempty"`
	UniqueID          string      `json:"unique_id"`
	Type              string      `json:"type"`
	Name              string      `json:"name"`
	UnitOfMeasurement string      `json:"unit_of_measurement,omitempty"`
	StateClass        string      `json:"state_class,omitempty"`
	EntityCategory    string      `json:"entity_category,omitempty"`
	Icon              string      `json:"icon,omitempty"`
	DeviceClass       string      `json:"device_class,omitempty"`
	Disabled          bool        `json:"disabled,omitempty"`
}

func (reg *SensorRegistrationInfo) RequestType() api.RequestType {
	return api.RequestTypeRegisterSensor
}

func (reg *SensorRegistrationInfo) RequestData() json.RawMessage {
	data, err := json.Marshal(reg)
	if err != nil {
		log.Debug().Err(err).
			Msg("Unable to marshal sensor to json.")
		return nil
	}
	return data
}

func (reg *SensorRegistrationInfo) ResponseHandler(res bytes.Buffer, respCh chan api.Response) {
	response, err := marshalResponse(res)
	if err != nil {
		respCh <- &SensorUpdateResponse{
			err: err,
		}
	}
	respCh <- NewSensorRegistrationResponse(response)
}

type SensorRegistrationResponse struct {
	err        error
	registered bool
}

func (r SensorRegistrationResponse) Error() error {
	return r.err
}

func (r SensorRegistrationResponse) Type() api.RequestType {
	return api.RequestTypeRegisterSensor
}

func (r SensorRegistrationResponse) Disabled() bool {
	return false
}

func (r SensorRegistrationResponse) Registered() bool {
	return r.registered
}

func NewSensorRegistrationResponse(r map[string]interface{}) *SensorRegistrationResponse {
	s := new(SensorRegistrationResponse)
	if v, ok := r["success"]; ok {
		success, err := assertAs[bool](v)
		if err != nil {
			s.err = err
			return s
		} else {
			if success {
				s.registered = true
				return s
			} else {
				s.err = errors.New("unsuccessful registration")
				return s
			}
		}
	}
	return s
}

// SensorUpdateInfo is the JSON structure required to update HA with the current
// sensor state.
type SensorUpdateInfo struct {
	StateAttributes interface{} `json:"attributes,omitempty"`
	State           interface{} `json:"state"`
	Icon            string      `json:"icon,omitempty"`
	Type            string      `json:"type"`
	UniqueID        string      `json:"unique_id"`
}

func (upd *SensorUpdateInfo) RequestType() api.RequestType {
	return api.RequestTypeUpdateSensorStates
}

func (upd *SensorUpdateInfo) RequestData() json.RawMessage {
	data, err := json.Marshal(upd)
	if err != nil {
		log.Debug().Err(err).
			Msg("Unable to marshal sensor to json.")
		return nil
	}
	return data
}

func (upd *SensorUpdateInfo) ResponseHandler(res bytes.Buffer, respCh chan api.Response) {
	response, err := marshalResponse(res)
	if err != nil {
		respCh <- &SensorUpdateResponse{
			err: err,
		}
	}
	respCh <- NewSensorUpdateResponse(upd.UniqueID, response)
}

type SensorUpdateResponse struct {
	err      error
	disabled bool
}

func (r SensorUpdateResponse) Error() error {
	return r.err
}

func (r SensorUpdateResponse) Type() api.RequestType {
	return api.RequestTypeUpdateSensorStates
}

func (r SensorUpdateResponse) Disabled() bool {
	return r.disabled
}

func (r SensorUpdateResponse) Registered() bool {
	return true
}

func NewSensorUpdateResponse(i string, r map[string]interface{}) *SensorUpdateResponse {
	s := new(SensorUpdateResponse)
	if v, ok := r[i]; ok {
		status, err := assertAs[map[string]interface{}](v)
		if err != nil {
			s.err = err
			return s
		}
		success, err := assertAs[bool](status["success"])
		if err != nil {
			s.err = err
			return s
		} else {
			if !success {
				hassErr, err := assertAs[map[string]interface{}](status["error"])
				if err != nil {
					s.err = errors.New("unknown error")
					return s
				} else {
					s.err = fmt.Errorf("code %s: %s", hassErr["code"], hassErr["message"])
					return s
				}
			}
			if _, ok := status["is_disabled"]; ok {
				s.disabled = true
			} else {
				s.disabled = false
			}
		}
	}

	return s
}

func marshalResponse(raw bytes.Buffer) (map[string]interface{}, error) {
	var r interface{}
	err := json.Unmarshal(raw.Bytes(), &r)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response (%s)", raw.String())
	}
	response, ok := r.(map[string]interface{})
	if !ok {
		return nil, errors.New("could not assert response as map")
	}
	return response, nil
}

func assertAs[T any](thing interface{}) (T, error) {
	if asT, ok := thing.(T); !ok {
		return *new(T), errors.New("could not assert value")
	} else {
		return asT, nil
	}

}
