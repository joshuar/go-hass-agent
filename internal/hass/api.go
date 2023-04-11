// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
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
	RequestData() interface{}
	ResponseHandler(interface{})
}

func MarshalJSON(request Request, secret string) ([]byte, error) {
	if secret != "" {
		return json.Marshal(&struct {
			Type          RequestType `json:"type"`
			Encrypted     bool        `json:"encrypted"`
			EncryptedData interface{} `json:"encrypted_data"`
		}{
			Type:          RequestTypeEncrypted,
			Encrypted:     true,
			EncryptedData: request.RequestData(),
		})
	} else {
		return json.Marshal(&struct {
			Type RequestType `json:"type"`
			Data interface{} `json:"data"`
		}{
			Type: request.RequestType(),
			Data: request.RequestData(),
		})
	}
}

type UnencryptedRequest struct {
	Type RequestType `json:"type"`
	Data interface{} `json:"data"`
}

type EncryptedRequest struct {
	Type          RequestType `json:"type"`
	Encrypted     bool        `json:"encrypted"`
	EncryptedData interface{} `json:"encrypted_data"`
}

type Response struct {
	Success bool `json:"success,omitempty"`
}

func APIRequest(ctx context.Context, request Request) {
	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	config, validConfig := config.FromContext(requestCtx)
	if !validConfig {
		log.Error().Caller().
			Msg("Could not retrieve valid config from context.")
		cancel()
		request.ResponseHandler(nil)
		return
	}

	reqJson, err := MarshalJSON(request, config.Secret)
	if err != nil {
		log.Error().Stack().Err(err).
			Msg("Unable to format request")
		request.ResponseHandler(nil)
	} else {
		var res interface{}
		err := requests.
			URL(config.APIURL).
			BodyBytes(reqJson).
			ToJSON(&res).
			Fetch(requestCtx)

		// requestFunc := func() error {
		// 	return requests.
		// 		URL(config.APIURL).
		// 		BodyBytes(reqJson).
		// 		ToJSON(&res).
		// 		Fetch(requestCtx)
		// }
		// retryNotifyFunc := func(e error, d time.Duration) {
		// 	log.Debug().Msgf("Retrying request %s in %v seconds.", string(reqJson), d.Seconds())
		// }
		// err := backoff.RetryNotify(requestFunc, backoff.NewExponentialBackOff(), retryNotifyFunc)
		if err != nil {
			log.Error().Stack().Err(err).
				Msgf("Unable to send request with body:\n\t%s\n\t", reqJson)
			cancel()
			request.ResponseHandler(nil)
		} else {
			request.ResponseHandler(res)
		}
	}
}
