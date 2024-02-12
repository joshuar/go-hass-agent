// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-resty/resty/v2"
)

type Authenticated interface {
	Auth() string
}

type Encrypted interface {
	Secret() string
}

type GetRequest interface {
	URL() string
	ResponseBody() any
}

type PostRequest interface {
	GetRequest
	RequestBody() json.RawMessage
}

type APIError struct {
	Message    string `json:"message,omitempty"`
	Code       int    `json:"code,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("%d: %s", e.Code, e.Message)
	} else {
		return e.Message
	}
}

func NewAPIError(code int, msg string) *APIError {
	return &APIError{
		Code:    code,
		Message: msg,
	}
}

type Response struct {
	Error *APIError
	Body  any
}

// ExecuteRequest sends an API request to Home Assistant. It supports either the
// REST or WebSocket API. By default and at a minimum, request are sent as GET
// requests and need to satisfy the GetRequest interface. To send a POST,
// satisfy the PostRequest interface. To add authentication where required,
// satisfy the Auth interface. To send an encrypted request, satisfy the Secret
// interface.
func ExecuteRequest2(ctx context.Context, req any) <-chan Response {
	responseCh := make(chan Response, 1)
	defer close(responseCh)

	client := resty.New()
	if a, ok := req.(Authenticated); ok {
		client = client.SetAuthToken(a.Auth())
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		response := Response{}
		var responseErr *APIError
		var resp *resty.Response
		var err error
		switch r := req.(type) {
		case PostRequest:
			log.Trace().
				Str("method", "POST").
				Str("url", r.URL()).
				RawJSON("body", r.RequestBody()).
				Time("sent_at", time.Now()).
				Msg("Sending request.")
			resp, err = client.R().
				SetContext(requestCtx).
				SetResult(r.ResponseBody()).
				SetBody(r.RequestBody()).
				SetError(&responseErr).
				Post(r.URL())
			response.Body = r.ResponseBody()
		case GetRequest:
			log.Trace().
				Str("method", "GET").
				Str("url", r.URL()).
				Time("sent_at", time.Now()).
				Msg("Sending request.")
			resp, err = client.R().
				SetContext(requestCtx).
				SetResult(r.ResponseBody()).
				SetError(&responseErr).
				Get(r.URL())
			response.Body = r.ResponseBody()
		}
		if err != nil {
			cancel()
			response.Error = NewAPIError(0, err.Error())
			responseCh <- response
			return
		}
		log.Trace().Err(err).
			Int("statuscode", resp.StatusCode()).
			Str("status", resp.Status()).
			Str("protocol", resp.Proto()).
			Dur("time", resp.Time()).
			Time("recieved_at", resp.ReceivedAt()).
			RawJSON("body", resp.Body()).Msg("Response recieved.")
		if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
			cancel()
			responseErr.StatusCode = resp.StatusCode()
			response.Error = responseErr
			responseCh <- response
			return
		}
		responseCh <- response
	}()
	wg.Wait()
	return responseCh
}
