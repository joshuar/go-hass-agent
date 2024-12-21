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

var (
	client *resty.Client

	// defaultRetryFunc defines how we retry requests. By default, requests are
	// only retried when Home Assistant responds with 429.
	defaultRetryFunc = func(r *resty.Response, _ error) bool {
		return r.StatusCode() == http.StatusTooManyRequests
	}
)

func init() {
	client = resty.New().
		SetTimeout(defaultTimeout).
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second).
		AddRetryCondition(defaultRetryFunc)
}

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

type Validator interface {
	Validate() error
}

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

func Send[T any](ctx context.Context, url string, details Request) (T, error) {
	var response T

	requestClient := client.R().SetContext(ctx)
	requestClient = requestClient.SetResult(&response)

	// If the request is authenticated, set the auth header with the token.
	if a, ok := details.(Authenticated); ok {
		requestClient = requestClient.SetAuthToken(a.Auth())
	}

	// If the request can be validated, validate it.
	if v, ok := details.(Validator); ok {
		if err := v.Validate(); err != nil {
			return response, fmt.Errorf("invalid request: %w", err)
		}
	}

	if details.Retry() {
		// If request needs to be retried, retry the request on any error.
		logging.FromContext(ctx).Debug("Will retry requests.", slog.Any("body", details))
		requestClient = requestClient.AddRetryCondition(
			func(r *resty.Response, err error) bool {
				if err != nil {
					logging.FromContext(ctx).Debug("Retrying request.", slog.Any("body", details))
					return true
				}
				return false
			},
		)
	}

	requestClient.SetBody(details.RequestBody())
	responseObj, err := requestClient.Post(url)

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
	case responseObj == nil:
		return response, fmt.Errorf("unknown error sending request")
	case responseObj.IsError():
		return response, &ResponseError{Code: responseObj.StatusCode(), Message: responseObj.Status()}
	}

	logging.FromContext(ctx).
		LogAttrs(ctx, logging.LevelTrace,
			"Received response.",
			slog.Int("statuscode", responseObj.StatusCode()),
			slog.String("status", responseObj.Status()),
			slog.String("protocol", responseObj.Proto()),
			slog.Duration("time", responseObj.Time()),
			slog.String("body", string(responseObj.Body())))

	return response, nil
}
