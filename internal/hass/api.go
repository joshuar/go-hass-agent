// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

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

//go:generate go-enum --marshal

// ENUM(encrypted,get_config,update_location,register_sensor,update_sensor_states)
type RequestType string

type Request interface {
	RequestType() RequestType
	RequestData() *json.RawMessage
	ResponseHandler(bytes.Buffer)
}

func MarshalJSON(request Request, secret string) ([]byte, error) {
	if request.RequestType() == RequestTypeEncrypted {
		if secret != "" {
			return json.Marshal(&struct {
				EncryptedData *json.RawMessage `json:"encrypted_data,omitempty"`
				Type          RequestType      `json:"type"`
				Encrypted     bool             `json:"encrypted"`
			}{
				Type:          RequestTypeEncrypted,
				Encrypted:     true,
				EncryptedData: request.RequestData(),
			})
		} else {
			return nil, errors.New("encrypted request recieved without secret")
		}
	} else {
		return json.Marshal(&struct {
			Data *json.RawMessage `json:"data,omitempty"`
			Type RequestType      `json:"type"`
		}{
			Type: request.RequestType(),
			Data: request.RequestData(),
		})
	}
}

type UnencryptedRequest struct {
	Data *json.RawMessage `json:"data,omitempty"`
	Type RequestType      `json:"type"`
}

type EncryptedRequest struct {
	EncryptedData *json.RawMessage `json:"encrypted_data,omitempty"`
	Type          RequestType      `json:"type"`
	Encrypted     bool             `json:"encrypted"`
}

type Response struct {
	Success bool `json:"success,omitempty"`
}

func APIRequest(ctx context.Context, request Request) {
	var res bytes.Buffer

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	secret, err := config.FetchPropertyFromContext(requestCtx, "secret")
	if err != nil {
		log.Error().Stack().Err(err).
			Msg("Could not fetch secret from agent config.")
		cancel()
		return
	}
	url, err := config.FetchPropertyFromContext(requestCtx, "apiURL")
	if err != nil {
		log.Error().Stack().Err(err).
			Msg("Could not fetch api url from agent config.")
		cancel()
		return
	}

	reqJson, err := MarshalJSON(request, secret.(string))
	if err != nil {
		log.Error().Stack().Err(err).
			Msg("Unable to format request")
		cancel()
		return
	} else {
		err := requests.
			URL(url.(string)).
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
}
