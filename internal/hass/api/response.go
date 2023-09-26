// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

type SensorResponse struct {
	responseType ResponseType
	disabled     bool
	registered   bool
}

func (r *SensorResponse) Type() ResponseType {
	return r.responseType
}

func (r *SensorResponse) Disabled() bool {
	return r.disabled
}

func (r *SensorResponse) Registered() bool {
	return r.registered
}

func parseRegistrationResponse(buf *bytes.Buffer) (*SensorResponse, error) {
	r, err := parseAsMap(buf)
	if err != nil {
		return nil, err
	}
	if _, ok := r["success"]; ok {
		if success, err := assertAs[bool](r["success"]); err != nil || !success {
			return nil, errors.New("unsuccessful registration")
		}
		return &SensorResponse{registered: true, responseType: ResponseTypeRegistration}, nil
	}
	return nil, errors.New("unknown response structure")
}

func parseUpdateResponse(buf *bytes.Buffer) (*SensorResponse, error) {
	r, err := parseAsMap(buf)
	if err != nil {
		return nil, err
	}
	for k, v := range r {
		log.Trace().Str("id", k).Msg("Parsing response for sensor.")
		r, err := assertAs[map[string]interface{}](v)
		if err != nil {
			return nil, err
		}
		if _, ok := r["success"]; ok {
			if success, err := assertAs[bool](r["success"]); err != nil || !success {
				if err != nil {
					return nil, err
				}
				log.Trace().Str("id", k).Msg("Unsuccessful response.")
				responseErr, err := assertAs[map[string]interface{}](r["error"])
				if err != nil {
					return nil, errors.New("unknown error")
				} else {
					return nil, fmt.Errorf("code %s: %s", responseErr["code"], responseErr["message"])
				}
			}
		}
		if _, ok := r["is_disabled"]; ok {
			log.Trace().Str("id", k).Bool("disabled", true).Msg("Successful response.")
			return &SensorResponse{disabled: true, responseType: ResponseTypeUpdate}, nil
		} else {
			log.Trace().Str("id", k).Bool("disabled", false).Msg("Successful response.")
			return &SensorResponse{disabled: false, responseType: ResponseTypeUpdate}, nil
		}
	}
	return nil, errors.New("unknown response structure")
}

func parseResponse(t RequestType, buf *bytes.Buffer) (interface{}, error) {
	switch t {
	case RequestTypeUpdateLocation:
		return buf.Bytes(), nil
	case RequestTypeGetConfig:
		return buf.Bytes(), nil
	case RequestTypeRegisterSensor:
		return parseRegistrationResponse(buf)
	case RequestTypeUpdateSensorStates:
		return parseUpdateResponse(buf)
	default:
		return nil, errors.New("unknown response")
	}
}

func parseAsMap(buf *bytes.Buffer) (map[string]interface{}, error) {
	var r interface{}
	err := json.Unmarshal(buf.Bytes(), &r)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response (%s)", buf.String())
	}
	rMap, ok := r.(map[string]interface{})
	if !ok {
		return nil, errors.New("could not parse response as map")
	}
	return rMap, nil
}

func assertAs[T any](thing interface{}) (T, error) {
	if asT, ok := thing.(T); !ok {
		return *new(T), errors.New("could not assert value")
	} else {
		return asT, nil
	}
}
