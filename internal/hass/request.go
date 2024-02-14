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
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-resty/resty/v2"
)

var (
	ErrInvalidURL        = errors.New("invalid URL")
	ErrResponseMalformed = errors.New("malformed response")
	ErrNoPrefs           = errors.New("loading preferences failed")
)

type Authenticated interface {
	Auth() string
}

type Encrypted interface {
	Secret() string
}

type GetRequest interface {
	ResponseBody() any
}

type PostRequest interface {
	GetRequest
	RequestBody() json.RawMessage
}

type APIError struct {
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	} else {
		return e.Message
	}
}

func NewAPIError(code, msg string) *APIError {
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
func ExecuteRequest(ctx context.Context, req any) <-chan Response {
	responseCh := make(chan Response, 1)
	defer close(responseCh)

	url := ContextGetURL(ctx)
	if url == "" {
		responseCh <- Response{
			Error: NewAPIError("", ErrInvalidURL.Error()),
		}
		return responseCh
	}

	client := ContextGetClient(ctx)
	if client == nil {
		responseCh <- Response{
			Error: NewAPIError("", "invalid client"),
		}
		return responseCh
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var responseErr *APIError
		var resp *resty.Response
		var err error
		response := Response{}
		cl := client.R().
			SetContext(requestCtx).
			SetError(&responseErr)
		if a, ok := req.(Authenticated); ok {
			cl = cl.SetAuthToken(a.Auth())
		}
		switch r := req.(type) {
		case PostRequest:
			log.Trace().
				Str("method", "POST").
				Str("url", url).
				RawJSON("body", r.RequestBody()).
				Time("sent_at", time.Now()).
				Msg("Sending request.")
			result := r.ResponseBody()
			resp, err = cl.
				SetResult(result).
				SetBody(r.RequestBody()).
				Post(url)
			response.Body = result
		case GetRequest:
			log.Trace().
				Str("method", "GET").
				Str("url", url).
				Time("sent_at", time.Now()).
				Msg("Sending request.")
			result := r.ResponseBody()
			resp, err = cl.
				SetResult(result).
				Get(url)
			response.Body = result
		}
		if err != nil {
			cancel()
			response.Error = NewAPIError("", err.Error())
			responseCh <- response
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
