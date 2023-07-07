// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/rs/zerolog/log"
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

//go:generate mockery --name Request --inpackage
type Request interface {
	RequestType() RequestType
	RequestData() json.RawMessage
	ResponseHandler(bytes.Buffer)
}

func MarshalJSON(request Request, secret string) ([]byte, error) {
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

type Response struct {
	Success bool `json:"success,omitempty"`
}

func APIRequest(ctx context.Context, request Request) {
	var res bytes.Buffer

	secret, err := config.FetchPropertyFromContext(ctx, "secret")
	if err != nil {
		log.Error().Stack().Err(err).
			Msg("Could not fetch secret from agent config.")
		return
	}
	url, err := config.FetchPropertyFromContext(ctx, "apiURL")
	if err != nil {
		log.Error().Stack().Err(err).
			Msg("Could not fetch api url from agent config.")
		return
	}
	urlString, ok := url.(string)
	if !ok { // type assertion failed
		log.Error().Stack().
			Msg("API URL does not appear to be valid.")
		return
	}

	reqJson, err := MarshalJSON(request, secret.(string))
	if err != nil {
		log.Error().Stack().Err(err).
			Msg("Unable to format request")
		return
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err = requests.
		URL(urlString).
		BodyBytes(reqJson).
		ToBytesBuffer(&res).
		Fetch(requestCtx)
	if err != nil {
		log.Error().Stack().Err(err).
			Msgf("Unable to send request with body:\n\t%s\n\t", reqJson)
		cancel()
		return
	}
	request.ResponseHandler(res)
}
