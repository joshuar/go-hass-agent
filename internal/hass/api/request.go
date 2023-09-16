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
	"github.com/joshuar/go-hass-agent/internal/agent/config"
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

func ExecuteRequest(ctx context.Context, request Request, agent Agent, responseCh chan Response) {
	var res bytes.Buffer

	defer close(responseCh)

	var apiURL, secret string
	if err := agent.GetConfig(config.PrefAPIURL, &apiURL); err != nil {
		responseCh <- NewGenericResponse(err, request.RequestType())
		return
	}
	if request.RequestType() == RequestTypeEncrypted {
		if err := agent.GetConfig(config.PrefSecret, &secret); err != nil {
			responseCh <- NewGenericResponse(err, request.RequestType())
			return
		}
	}

	reqJSON, err := marshalJSON(request, secret)
	if err != nil {
		responseCh <- NewGenericResponse(err, request.RequestType())
		return
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err = requests.
		URL(apiURL).
		BodyBytes(reqJSON).
		ToBytesBuffer(&res).
		Fetch(requestCtx)
	if err != nil {
		cancel()
		responseCh <- NewGenericResponse(err, request.RequestType())
		return
	}
	request.ResponseHandler(res, responseCh)
}
