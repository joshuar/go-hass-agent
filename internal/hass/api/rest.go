// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go run github.com/matryer/moq -out api_mocks_test.go . PostRequest
package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	defaultTimeout = 30 * time.Second
)

// Set-up a default Resty client with a reasonable timeout and an
// exponential retry process for 429 responses from Home Assistant.
var restClient = resty.New().
	SetTimeout(defaultTimeout).
	SetRetryCount(3).
	SetRetryWaitTime(5 * time.Second).
	SetRetryMaxWaitTime(20 * time.Second).
	AddRetryCondition(func(r *resty.Response, _ error) bool {
		return r.StatusCode() == http.StatusTooManyRequests
	})

// Request is an API request to Home Assistant. It has a request body (typically
// JSON) and a boolean to indicate whether the request should be retried (with a
// default exponential backoff).
type Request interface {
	RequestBody() any
	Retry() bool
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

// Validator represents a request that should be validated before being sent.
type Validator interface {
	Validate() error
}

// ResponseError contains Home Assistant API error response details.
type ResponseError struct {
	Code    any    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *ResponseError) Error() string {
	var msg []string
	if e.Code != nil {
		msg = append(msg, fmt.Sprintf("code %v", e.Code))
	}

	if e.Message != "" {
		msg = append(msg, e.Message)
	}

	if len(msg) == 0 {
		msg = append(msg, "unknown error")
	}

	return strings.Join(msg, ": ")
}

// Send will send the given request to the specified URL. It will handle
// marshaling the request and unmarshaling the response. It can optionally set
// an Auth token for requests that require it and validate the request before
// sending. It will also handle retrying the request with an exponential backoff
// if requested.
func Send[T any](ctx context.Context, url string, details Request) (T, error) {
	var response T

	apiRequest := restClient.R().SetContext(ctx)
	apiRequest = apiRequest.SetResult(&response)

	// If the request is authenticated, set the auth header with the token.
	if a, ok := details.(Authenticated); ok {
		apiRequest = apiRequest.SetAuthToken(a.Auth())
	}

	// If the request can be validated, validate it.
	if v, ok := details.(Validator); ok {
		if err := v.Validate(); err != nil {
			return response, fmt.Errorf("invalid request: %w", err)
		}
	}

	// If request needs to be retried, retry the request on any error.
	if details.Retry() {
		apiRequest = apiRequest.AddRetryCondition(
			func(_ *resty.Response, err error) bool {
				return err != nil
			},
		)
	}

	apiRequest.SetBody(details.RequestBody())
	apiResponse, err := apiRequest.Post(url)

	logging.FromContext(ctx).
		LogAttrs(ctx, logging.LevelTrace,
			"Sending request.",
			slog.String("method", "POST"),
			slog.String("url", url),
			slog.Any("body", details),
			slog.Time("sent_at", time.Now()))

	switch {
	case err != nil:
		return response, fmt.Errorf("error sending request: %w", err)
	case apiResponse == nil:
		return response, fmt.Errorf("unknown error sending request")
	case apiResponse.IsError():
		return response, &ResponseError{Code: apiResponse.StatusCode(), Message: apiResponse.Status()}
	}

	logging.FromContext(ctx).
		LogAttrs(ctx, logging.LevelTrace,
			"Received response.",
			slog.Int("statuscode", apiResponse.StatusCode()),
			slog.String("status", apiResponse.Status()),
			slog.String("protocol", apiResponse.Proto()),
			slog.Duration("time", apiResponse.Time()),
			slog.String("body", string(apiResponse.Body())))

	return response, nil
}
