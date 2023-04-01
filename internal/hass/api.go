package hass

import (
	"context"
	"encoding/json"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog/log"
)

//go:generate go-enum --marshal

// ENUM(encrypted,get_config,update_location,register_sensor,update_sensor_states)
type RequestType string

type Request interface {
	RequestType() RequestType
	RequestData() interface{}
	IsEncrypted() bool
}

func MarshalJSON(request Request) ([]byte, error) {
	if request.IsEncrypted() {
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
	// Type    string `json:"type,omitempty"`
	// Error   struct {
	// 	Code    string `json:"code"`
	// 	Message string `json:"message"`
	// } `json:"error,omitempty"`
	// ID string `json:"id,omitempty"`
}

func APIRequest(ctx context.Context, url string, request interface{}, response func(r interface{})) {

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	reqJson, err := MarshalJSON(request.(Request))
	if err != nil {
		log.Error().Msgf("Unable to format request: %v", err)
		response(nil)
	} else {
		var res interface{}
		requestFunc := func() error {
			return requests.
				URL(url).
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
