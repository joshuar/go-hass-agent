package hass

import (
	"context"
	"encoding/json"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/cenkalti/backoff/v4"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/rs/zerolog/log"
)

//go:generate go-enum --marshal

// ENUM(encrypted,get_config,update_location,register_sensor,update_sensor_states)
type RequestType string

type Request interface {
	RequestType() RequestType
	RequestData() interface{}
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

func APIRequest(ctx context.Context, request interface{}, response func(r interface{})) {

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	config, validConfig := config.FromContext(requestCtx)
	if !validConfig {
		log.Debug().Caller().Msg("Could not retrieve valid config from context.")
		cancel()
		response(nil)
		return
	}

	reqJson, err := MarshalJSON(request.(Request), config.Secret)
	if err != nil {
		log.Error().Msgf("Unable to format request: %v", err)
		response(nil)
	} else {
		var res interface{}
		requestFunc := func() error {
			return requests.
				URL(config.APIURL).
				BodyBytes(reqJson).
				ToJSON(&res).
				Fetch(requestCtx)
		}
		err := backoff.Retry(requestFunc, backoff.WithContext(backoff.NewExponentialBackOff(), requestCtx))
		if err != nil {
			log.Error().Msgf("Unable to send request: %v", err)
			cancel()
			response(nil)
		} else {
			response(res)
		}
	}
}
