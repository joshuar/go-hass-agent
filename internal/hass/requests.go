package hass

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/carlmjohnson/requests"
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

func RequestDispatcher(requestURL string, requestsCh, responsesCh chan interface{}) {
	var wg sync.WaitGroup
	for r := range requestsCh {
		wg.Add(1)
		go func(r interface{}) {
			ctx := context.Background()
			defer wg.Done()
			// spew.Dump(r.(Request))
			req, err := MarshalJSON(r.(Request))
			if err != nil {
				log.Error().Msgf("Unable to format request: %v", err)
				responsesCh <- nil
			} else {
				var res interface{}
				err = requests.
					URL(requestURL).
					BodyBytes(req).
					ToJSON(&res).
					Fetch(ctx)
				// spew.Dump(res)
				if err != nil {
					log.Error().Msgf("Unable to send request: %v", err)
				} else {
					responsesCh <- res
				}
			}
		}(r)
	}
	wg.Wait()
}

type Conn struct {
	requestsCh, responsesCh chan interface{}
	requestURL              string
}

func NewConnection(requestURL string) *Conn {
	newConn := &Conn{
		requestsCh:  make(chan interface{}),
		responsesCh: make(chan interface{}),
		requestURL:  requestURL,
	}
	go newConn.processRequests()
	return newConn
}

func (c *Conn) processRequests() {
	var wg sync.WaitGroup
	for request := range c.requestsCh {
		wg.Add(1)
		go func(request interface{}) {
			ctx := context.Background()
			defer wg.Done()
			reqJson, err := MarshalJSON(request.(Request))
			if err != nil {
				log.Error().Msgf("Unable to format request: %v", err)
				c.responsesCh <- nil
			} else {
				var res interface{}
				err = requests.
					URL(c.requestURL).
					BodyBytes(reqJson).
					ToJSON(&res).
					Fetch(ctx)
				// spew.Dump(res)
				if err != nil {
					log.Error().Msgf("Unable to send request: %v", err)
					c.responsesCh <- nil
				} else {
					c.responsesCh <- res
				}
			}
		}(request)
	}
	wg.Wait()
}

func (c *Conn) SendRequest(request Request) interface{} {
	c.requestsCh <- request
	return <-c.responsesCh
}

type WebsocketResponse struct {
	Type    string `json:"type"`
	Success bool   `json:"success,omitempty"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
	ID           int         `json:"id,omitempty"`
	Result       interface{} `json:"result,omitempty"`
	HAVersion    string      `json:"ha_version,omitempty"`
	Notification struct {
		Message   string `json:"message,omitempty"`
		Title     string `json:"title,omitempty"`
		ConfirmID string `json:"confirm_id,omitempty"`
	} `json:"event,omitempty"`
}
