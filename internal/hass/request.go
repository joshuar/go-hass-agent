// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
//go:generate moq -out request_mocks_test.go . PostRequest Response
package hass

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/joshuar/go-hass-agent/internal/logging"

	"github.com/go-resty/resty/v2"
)

var (
	ErrInvalidURL        = errors.New("invalid URL")
	ErrInvalidClient     = errors.New("invalid client")
	ErrResponseMalformed = errors.New("malformed response")
	ErrUnknown           = errors.New("unknown error occurred")

	defaultTimeout = 30 * time.Second
	defaultRetry   = func(r *resty.Response, _ error) bool {
		return r.StatusCode() == http.StatusTooManyRequests
	}
)

// APIError represents an error returned either by the HTTP layer or by the Home
// Assistant API. The StatusCode reflects the HTTP status code returned while
// Message and Code are additional and optional values returned from the Home
// Assistant API.
type APIError struct {
	Code       any    `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	switch {
	case e.Code != nil:
		return fmt.Sprintf("%v: %s", e.Code, e.Message)
	case e.StatusCode > 0:
		return fmt.Sprintf("Status: %d", e.StatusCode)
	default:
		return e.Message
	}
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
	json.Unmarshaler
	UnmarshalError(data []byte) error
	error
}

// ExecuteRequest sends an API request to Home Assistant. It supports either the
// REST or WebSocket API. By default and at a minimum, request are sent as GET
// requests and need to satisfy the GetRequest interface. To send a POST,
// satisfy the PostRequest interface. To add authentication where required,
// satisfy the Auth interface. To send an encrypted request, satisfy the Secret
// interface.
func ExecuteRequest(ctx context.Context, client *resty.Client, url string, request any, response Response) error {
	if client == nil {
		return ErrInvalidClient
	}

	var resp *resty.Response

	var err error

	webClient := client.R().
		SetContext(ctx)
	if a, ok := request.(Authenticated); ok {
		webClient = webClient.SetAuthToken(a.Auth())
	}

	switch req := request.(type) {
	case PostRequest:
		logging.FromContext(ctx).
			LogAttrs(ctx, logging.LevelTrace,
				"Sending request.",
				slog.String("method", "POST"),
				slog.String("body", string(req.RequestBody())),
				slog.Time("sent_at", time.Now()))

		resp, err = webClient.SetBody(req.RequestBody()).Post(url)
	case GetRequest:
		logging.FromContext(ctx).
			LogAttrs(ctx, logging.LevelTrace,
				"Sending request.",
				slog.String("method", "GET"),
				slog.Time("sent_at", time.Now()))

		resp, err = webClient.Get(url)
	}

	// If the client fails to send the request, return a wrapped error.
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}

	logging.FromContext(ctx).
		LogAttrs(ctx, logging.LevelTrace,
			"Received response.",
			slog.Int("statuscode", resp.StatusCode()),
			slog.String("status", resp.Status()),
			slog.String("protocol", resp.Proto()),
			slog.Duration("time", resp.Time()),
			slog.String("body", string(resp.Body())))

	// If the response is an error code, unmarshal it with the error method.
	if resp.IsError() {
		if err := response.UnmarshalError(resp.Body()); err != nil {
			return ErrUnknown
		}

		return response
	}
	// Otherwise for a successful response, if the response body is not an empty
	// string, unmarshal it.
	if string(resp.Body()) != "" {
		if err := response.UnmarshalJSON(resp.Body()); err != nil {
			return errors.Join(ErrResponseMalformed, err)
		}
	}

	return nil
}

func NewDefaultHTTPClient(url string) *resty.Client {
	return resty.New().
		SetTimeout(defaultTimeout).
		AddRetryCondition(defaultRetry).
		SetBaseURL(url)
}
