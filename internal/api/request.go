// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/joshuar/go-hass-agent/internal/settings"
)

//go:generate stringer -type=RequestType -output requestType.go -linecomment
const (
	RequestTypeEncrypted          RequestType = iota + 1 // encrypted
	RequestTypeGetConfig                                 // get_config
	RequestTypeUpdateLocation                            // update_location
	RequestTypeRegisterSensor                            // register_sensor
	RequestTypeUpdateSensorStates                        // update_sensor_states
)

type RequestType int

//go:generate moq -out mock_Request_test.go . Request
type Request interface {
	RequestType() RequestType
	RequestData() json.RawMessage
	ResponseHandler(bytes.Buffer, chan Response)
}

func marshalJSON(request Request, secret string) ([]byte, error) {
	if request.RequestType() == RequestTypeEncrypted {
		if secret != "" {
			return json.Marshal(&EncryptedRequest{
				Type:          RequestTypeEncrypted.String(),
				Encrypted:     true,
				EncryptedData: request.RequestData(),
			})
		} else {
			return nil, errors.New("encrypted request recieved without secret")
		}
	} else {
		return json.Marshal(&UnencryptedRequest{
			Type: request.RequestType().String(),
			Data: request.RequestData(),
		})
	}
}

type UnencryptedRequest struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type EncryptedRequest struct {
	Type          string          `json:"type"`
	EncryptedData json.RawMessage `json:"encrypted_data,omitempty"`
	Encrypted     bool            `json:"encrypted"`
}

func ExecuteRequest(ctx context.Context, request Request, responseCh chan Response) {
	var res bytes.Buffer

	settingsStore, err := settings.FetchFromContext(ctx)
	if err != nil {
		responseCh <- NewGenericResponse(err, request.RequestType())
		return
	}

	var secret string
	if request.RequestType() == RequestTypeEncrypted {
		s, err := settingsStore.GetValue(settings.Secret)
		if err != nil {
			responseCh <- NewGenericResponse(err, request.RequestType())
			return
		}
		secret = s
	} else {
		secret = ""
	}

	var url string
	u, err := settingsStore.GetValue(settings.ApiURL)
	if err != nil {
		responseCh <- NewGenericResponse(err, request.RequestType())
		return
	} else {
		url = u
	}

	reqJson, err := marshalJSON(request, secret)
	if err != nil {
		responseCh <- NewGenericResponse(err, request.RequestType())
		return
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err = requests.
		URL(url).
		BodyBytes(reqJson).
		ToBytesBuffer(&res).
		Fetch(requestCtx)
	if err != nil {
		cancel()
		responseCh <- NewGenericResponse(err, request.RequestType())
		return
	}
	request.ResponseHandler(res, responseCh)
}
