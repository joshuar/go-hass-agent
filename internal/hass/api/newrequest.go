// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/carlmjohnson/requests"
)

var (
	ErrMalformedRequest  = errors.New("malformed request body")
	ErrMalformedResponse = errors.New("could not parse response body")
	ErrFailedResponse    = errors.New("response failed")
)

type Request2 interface {
	URL() string
	Auth() string
	Body() json.RawMessage
}

type Response struct {
	Error error
	Body  json.RawMessage
}

func ExecuteRequest2(ctx context.Context, req Request2) <-chan Response {
	responseCh := make(chan Response, 1)
	defer close(responseCh)

	r := requests.
		URL(req.URL()).
		Header("Authorization", "Bearer "+req.Auth())

	if req.Body() != nil {
		reqJSON, err := json.Marshal(req.Body())
		if err != nil {
			responseCh <- Response{
				Error: ErrMalformedRequest,
			}
			return responseCh
		}
		r = r.BodyBytes(reqJSON).ContentType("application/json")
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var rBuf bytes.Buffer
		rErr := make(map[string]any)
		err := r.
			ToBytesBuffer(&rBuf).
			ErrorJSON(&rErr).
			Fetch(requestCtx)
		if len(rErr) != 0 {
			cancel()
			responseCh <- Response{
				Error: errors.New(rErr["message"].(string)),
			}
			return
		}
		if err != nil {
			cancel()
			responseCh <- Response{
				Error: err,
			}
			return
		}
		var respJSON json.RawMessage
		err = json.Unmarshal(rBuf.Bytes(), &respJSON)
		if err != nil {
			responseCh <- Response{
				Error: errors.Join(ErrMalformedResponse, err),
			}
		} else {
			responseCh <- Response{
				Body: respJSON,
			}
		}
	}()
	wg.Wait()
	return responseCh
}
