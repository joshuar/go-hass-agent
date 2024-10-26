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

	defaultRetryFunc = func(r *resty.Response, _ error) bool {
		return r.StatusCode() == http.StatusTooManyRequests
	}
)

func init() {
	client = resty.New().
		SetTimeout(defaultTimeout).
		AddRetryCondition(defaultRetryFunc)
}

type RawRequest interface {
	RequestBody() any
}

// Request is a HTTP POST request with the request body provided by Body().
type Request interface {
	RequestType() string
	RequestData() any
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

type requestBody struct {
	Data        any    `json:"data"`
	RequestType string `json:"type"`
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

func Send[T any](ctx context.Context, url string, details any) (T, error) {
	var (
		response    T
		responseErr ResponseError
		responseObj *resty.Response
	)

	requestClient := client.R().SetContext(ctx)
	requestClient = requestClient.SetError(&responseErr)
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

	switch request := details.(type) {
	case Request:
		body := &requestBody{
			RequestType: request.RequestType(),
			Data:        request.RequestData(),
		}
		logging.FromContext(ctx).
			LogAttrs(ctx, logging.LevelTrace,
				"Sending request.",
				slog.String("method", "POST"),
				slog.String("url", url),
				slog.Any("body", body),
				slog.Time("sent_at", time.Now()))

		responseObj, _ = requestClient.SetBody(body).Post(url) //nolint:errcheck // error is checked with responseObj.IsError()
	case RawRequest:
		logging.FromContext(ctx).
			LogAttrs(ctx, logging.LevelTrace,
				"Sending request.",
				slog.String("method", "POST"),
				slog.String("url", url),
				slog.Any("body", request),
				slog.Time("sent_at", time.Now()))

		responseObj, _ = requestClient.SetBody(request).Post(url) //nolint:errcheck // error is checked with responseObj.IsError()
	}

	logging.FromContext(ctx).
		LogAttrs(ctx, logging.LevelTrace,
			"Received response.",
			slog.Int("statuscode", responseObj.StatusCode()),
			slog.String("status", responseObj.Status()),
			slog.String("protocol", responseObj.Proto()),
			slog.Duration("time", responseObj.Time()),
			slog.String("body", string(responseObj.Body())))

	if responseObj.IsError() {
		return response, &ResponseError{Code: responseObj.StatusCode(), Message: responseObj.Status()}
	}

	return response, nil
}
