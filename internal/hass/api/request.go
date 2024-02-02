// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/carlmjohnson/requests"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

//go:generate stringer -type=RequestType,ResponseType -output apiTypes.go -linecomment
const (
	RequestTypeEncrypted          RequestType = iota + 1 // encrypted
	RequestTypeGetConfig                                 // get_config
	RequestTypeUpdateLocation                            // update_location
	RequestTypeRegisterSensor                            // register_sensor
	RequestTypeUpdateSensorStates                        // update_sensor_states

	ResponseTypeRegistration ResponseType = iota + 1 // registration
	ResponseTypeUpdate                               // update
)

type (
	RequestType  int
	ResponseType int
)

//go:generate moq -out mock_Request_test.go . Request
type Request interface {
	RequestType() RequestType
	RequestData() json.RawMessage
}

func marshalJSON(request Request, secret string) ([]byte, error) {
	if request == nil {
		return nil, errors.New("nil request")
	}
	if request.RequestType() != RequestTypeEncrypted {
		return json.Marshal(&UnencryptedRequest{
			Type: request.RequestType().String(),
			Data: request.RequestData(),
		})
	}
	if secret != "" && secret != "NOTSET" {
		return json.Marshal(&EncryptedRequest{
			Type:          RequestTypeEncrypted.String(),
			Encrypted:     true,
			EncryptedData: request.RequestData(),
		})
	}
	return nil, errors.New("encrypted request received without secret")
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

func ExecuteRequest(ctx context.Context, request Request) <-chan any {
	responseCh := make(chan any, 1)
	defer close(responseCh)

	prefs := preferences.FetchFromContext(ctx)

	reqJSON, err := marshalJSON(request, prefs.Secret)
	if err != nil {
		responseCh <- err
		return responseCh
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var rBuf bytes.Buffer
		err = requests.
			URL(prefs.RestAPIURL).
			BodyBytes(reqJSON).
			ToBytesBuffer(&rBuf).
			Fetch(requestCtx)
		if err != nil {
			cancel()
			responseCh <- err
		} else {
			response, err := parseResponse(request.RequestType(), &rBuf)
			if err != nil {
				responseCh <- err
			} else {
				responseCh <- response
			}
		}
	}()
	wg.Wait()
	return responseCh
}
