// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package api contains methods and objects for interacting with the Home Assistant REST API.
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	defaultTimeout      = 30 * time.Second
	defaultRetryWait    = 5 * time.Second
	defaultRetryCount   = 5
	defaultRetryMaxWait = 20 * time.Second
)

// ErrUnknown is returned when an error occurred but the reason/cause could not be determined or was unexpected.
var ErrUnknown = errors.New("an unknown error occurred")

var ErrUnsupported = errors.New("unsupported request")

var restAPIClient *resty.Client

// Error allows an APIError to satisfy the Go Error interface.
func (e *APIError) Error() string {
	var msg []string
	if e.Code != "" {
		msg = append(msg, e.Code)
	}

	if e.Message != "" {
		msg = append(msg, e.Message)
	}

	if len(msg) == 0 {
		msg = append(msg, "unknown error")
	}

	return strings.Join(msg, ": ")
}

// HasError determines whether the response status indicates an error condition
// has occurred.
func (r *ResponseStatus) HasError() error {
	if r.Error.IsSpecified() {
		apiErr, err := r.Error.Get()
		if err != nil {
			return fmt.Errorf("%w: %w", ErrUnknown, err)
		}

		return &apiErr
	}

	if r.Error.IsNull() {
		return ErrUnknown
	}

	return nil
}

// HasSuccess determines whether the response status indicates the request was
// successful.
func (r *ResponseStatus) HasSuccess() (bool, error) {
	if r.IsSuccess.IsSpecified() {
		return r.IsSuccess.Get() //nolint:wrapcheck
	}

	if r.IsSuccess.IsNull() {
		return true, nil
	}

	return false, nil
}

// SensorDisabled determines whether the response status indicates the sensor was disabled.
func (r *ResponseStatus) SensorDisabled() bool {
	if r.IsDisabled.IsSpecified() {
		if disabled, err := r.IsDisabled.Get(); err == nil {
			return disabled
		}
	}

	return false
}

// setupClient sets up a new client for making requests to the Home Assistant rest api.
var setupClient = sync.OnceFunc(func() {
	restAPIClient = resty.New().
		SetTimeout(defaultTimeout).
		SetRetryCount(defaultRetryCount).
		SetRetryWaitTime(defaultRetryWait).
		SetRetryMaxWaitTime(defaultRetryMaxWait).
		AddRetryCondition(func(r *resty.Response, _ error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		})
})

type Request struct {
	*resty.Request

	method string
}

func NewRequest(options ...RequestOption) *Request {
	setupClient()

	req := &Request{
		Request: restAPIClient.R(),
	}
	for option := range slices.Values(options) {
		option(req)
	}

	if req.method == "" {
		req.method = http.MethodPost
	}

	return req
}

func (r *Request) Do(ctx context.Context, url string) (*Response, error) {
	r.Request = r.SetContext(ctx)

	var (
		raw *resty.Response
		err error
	)
	switch r.method {
	case http.MethodPost:
		raw, err = r.Request.Post(url)
	default:
		raw, err = r.Request.Get(url)
	}

	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return &Response{Response: raw}, nil
}

type RequestOption func(*Request)

func WithMethod(method string) RequestOption {
	return func(r *Request) {
		r.method = method
	}
}

func WithBody[T any](body T) RequestOption {
	return func(r *Request) {
		r.Request = r.SetBody(body)
	}
}

func WithResult[T any](res T) RequestOption {
	return func(r *Request) {
		r.Request = r.SetResult(res)
	}
}

func WithAuth(token string) RequestOption {
	return func(r *Request) {
		r.Request = r.SetAuthToken(token)
	}
}

func WithRetryable(value bool) RequestOption {
	return func(r *Request) {
		if value {
			r.Request = r.AddRetryCondition(
				func(_ *resty.Response, err error) bool {
					return err != nil
				},
			)
		}
	}
}

func WithTrace() RequestOption {
	return func(r *Request) {
		r.Request = r.EnableTrace()
	}
}

type Response struct {
	*resty.Response
}

func (r *RequestData) String() string {
	var str strings.Builder

	str.WriteString("Type: " + string(r.Type))
	str.WriteRune('\n')
	str.WriteString("Payload:")
	str.WriteRune('\n')
	str.WriteString(string(r.Payload.union))
	str.WriteRune('\n')
	return str.String()
}
