package hass

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/carlmjohnson/requests"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
)

//go:generate go-enum --marshal

// ENUM(encrypted,get_config,update_location,register_sensor,update_sensor_states)
type RequestType string

type Request interface {
	RequestType() RequestType
	RequestData() interface{}
	IsEncrypted() bool
	// MarshalJSON() ([]byte, error)
}

// type Request struct {
// 	Type          RequestType
// 	Data          interface{}
// 	Encrypted     bool
// 	EncryptedData bool
// }

func MarshalJSON(request Request) ([]byte, error) {
	// data, err := request.MarshalJSON()
	// if err != nil {
	// 	return nil, err
	// } else {
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
	// }
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
	Type   string `json:"type"`
	ID     int    `json:"id"`
	Result struct {
		Body    string      `json:"body"`
		Status  string      `json:"status"`
		Headers interface{} `json:"headers"`
	} `json:"result"`
}

// func formatRequest(request Request) interface{} {
// 	if request.IsEncrypted() {
// 		return &EncryptedRequest{
// 			Type:          RequestTypeEncrypted,
// 			Encrypted:     true,
// 			EncryptedData: request.RequestData(),
// 		}
// 	} else {
// 		return &UnencryptedRequest{
// 			Type: request.RequestType(),
// 			Data: request.RequestData(),
// 		}
// 	}
// }

func RequestDispatcher(requestURL string, requestsCh, responsesCh chan interface{}) {
	var wg sync.WaitGroup
	// res := Response{}
	res := make(map[string]interface{})
	for r := range requestsCh {
		wg.Add(1)
		go func(r interface{}) {
			ctx := context.Background()
			defer wg.Done()
			req, err := MarshalJSON(r.(Request))
			spew.Dump(req)
			if err != nil {
				log.Error().Msgf("Unable to format request: %v", err)
				responsesCh <- err
			} else {
				err = requests.
					URL(requestURL).
					BodyBytes(req).
					ToJSON(&res).
					Fetch(ctx)
				spew.Dump(res)
				if err != nil {
					log.Error().Msgf("Request unsuccessful: %v", err)
					responsesCh <- err
				} else {
					log.Debug().Caller().Msg("Request successful.")
					responsesCh <- res
				}
			}
		}(r)
	}
	wg.Wait()
}
