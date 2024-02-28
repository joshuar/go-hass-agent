// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-resty/resty/v2"
)

var (
	ErrInvalidURL        = errors.New("invalid URL")
	ErrInvalidClient     = errors.New("invalid client")
	ErrResponseMalformed = errors.New("malformed response")
	ErrNoPrefs           = errors.New("loading preferences failed")
)

// APIError represents an error returned either by the HTTP layer or by the Home
// Assistant API. The StatusCode reflects the HTTP status code returned while
// Message and Code are additional and optional values returned from the Home
// Assistant API.
type APIError struct {
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return e.Message
}

// GetRequest is a HTTP GET request.
type GetRequest any

// PostRequest is a HTTP POST request with the request body provided by Body().
type PostRequest interface {
	RequestBody() json.RawMessage
}

// Authenticated represents a request that requires passing an authentication
// header with the value returned by Auth().
type Authenticated interface {
	Auth() string
}

// Encrypted represents a request that should be encrypted with the secret
// provided by Secret().
type Encrypted interface {
	Secret() string
}

type Response interface {
	StoreError(e error)
	error
	json.Unmarshaler
}

// ExecuteRequest sends an API request to Home Assistant. It supports either the
// REST or WebSocket API. By default and at a minimum, request are sent as GET
// requests and need to satisfy the GetRequest interface. To send a POST,
// satisfy the PostRequest interface. To add authentication where required,
// satisfy the Auth interface. To send an encrypted request, satisfy the Secret
// interface.
func ExecuteRequest(ctx context.Context, request any, response Response) {
	url := ContextGetURL(ctx)
	if url == "" {
		response.StoreError(ErrInvalidURL)
		return
	}

	client := ContextGetClient(ctx)
	if client == nil {
		response.StoreError(ErrInvalidClient)
		return
	}

	var responseErr *APIError
	var resp *resty.Response
	var err error
	cl := client.R().
		SetContext(ctx).
		SetError(&responseErr)
	if a, ok := request.(Authenticated); ok {
		cl = cl.SetAuthToken(a.Auth())
	}
	switch r := request.(type) {
	case PostRequest:
		log.Trace().
			Str("method", "POST").
			RawJSON("body", r.RequestBody()).
			Time("sent_at", time.Now()).
			Msg("Sending request.")
		resp, err = cl.
			SetBody(r.RequestBody()).
			Post(url)
	case GetRequest:
		log.Trace().
			Str("method", "GET").
			Time("sent_at", time.Now()).
			Msg("Sending request.")
		resp, err = cl.
			Get(url)
	}
	if err != nil {
		response.StoreError(err)
		return
	}
	log.Trace().Err(err).
		Int("statuscode", resp.StatusCode()).
		Str("status", resp.Status()).
		Str("protocol", resp.Proto()).
		Dur("time", resp.Time()).
		Time("received_at", resp.ReceivedAt()).
		RawJSON("body", resp.Body()).Msg("Response received.")
	if resp.IsError() {
		err := fmt.Errorf("%s (StatusCode: %d)", responseErr.Error(), resp.StatusCode())
		response.StoreError(err)
		return
	}
	if err := response.UnmarshalJSON(resp.Body()); err != nil {
		response.StoreError(errors.Join(ErrResponseMalformed, err))
	}
}

func NewDefaultHTTPClient() *resty.Client {
	return resty.New().
		SetTimeout(1 * time.Second).
		AddRetryCondition(
			// RetryConditionFunc type is for retry condition function
			// input: non-nil Response OR request execution error
			func(r *resty.Response, err error) bool {
				return r.StatusCode() == http.StatusTooManyRequests
			},
		)
}
