// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type SensorResponseBody struct {
	Error    ResponseError `json:"error,omitempty"`
	Success  bool          `json:"success"`
	Disabled bool          `json:"is_disabled,omitempty"`
}

type ResponseError struct {
	ErrorCode string `json:"code"`
	ErrorMsg  string `json:"message"`
}

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
	var r map[string]SensorResponseBody
	err := json.Unmarshal(buf.Bytes(), &r)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response (%s)", buf.String())
	}
	for sensorID, v := range r {
		if !v.Success {
			return nil, fmt.Errorf("sensor %s, code %s: %s", sensorID, v.Error.ErrorCode, v.Error.ErrorMsg)
		}
		return &SensorResponse{disabled: v.Disabled, responseType: ResponseTypeUpdate}, nil
	}
	return nil, errors.New("unknown response structure")
}

func parseResponse(t RequestType, buf *bytes.Buffer) (any, error) {
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

func parseAsMap(t any) (map[string]any, error) {
	switch v := t.(type) {
	case *bytes.Buffer:
		var r map[string]any
		err := json.Unmarshal(v.Bytes(), &r)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal response (%s)", v.String())
		}
		return r, nil
	case any:
		r, err := assertAs[map[string]any](v)
		if err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, errors.New("unsupported type")
	}
}

func assertAs[T any](thing any) (T, error) {
	var asT T
	var ok bool
	if asT, ok = thing.(T); !ok {
		return *new(T), errors.New("could not assert value")
	}
	return asT, nil
}
