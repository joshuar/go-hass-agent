// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,lll,paralleltest,wsl
//revive:disable:function-length
package hass

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	type fields struct {
		Code       any
		Message    string
		StatusCode int
	}

	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "valid error with string code",
			fields: fields{Code: "404", Message: "Not Found"},
			want:   "404: Not Found",
		},
		{
			name:   "valid error with int code",
			fields: fields{Code: 404, Message: "Not Found"},
			want:   "404: Not Found",
		},
		{
			name: "empty",
		},
		{
			name:   "valid statuscode",
			fields: fields{StatusCode: 503},
			want:   "Status: 503",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &APIError{
				Code:       tt.fields.Code,
				Message:    tt.fields.Message,
				StatusCode: tt.fields.StatusCode,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("APIError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ?: mock API level responses and test those.
//
//nolint:containedctx
func TestExecuteRequest(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/goodPost", request.URL.Path == "/goodGet":
			fmt.Fprintf(response, `{"success": true}`)
		case request.URL.Path == "/badPost", request.URL.Path == "/badGet":
			response.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(response, `{"success": false}`)
		case request.URL.Path == "/badData":
			response.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(response, `{"success": false}`)
		}
	}))

	goodPostReq := PostRequestMock{
		RequestBodyFunc: func() json.RawMessage { return json.RawMessage(`{"field":"value"}`) },
	}
	goodPostResp := &ResponseMock{
		UnmarshalJSONFunc: func(_ []byte) error { return nil },
	}
	goodGetResp := &ResponseMock{
		UnmarshalJSONFunc: func(_ []byte) error { return nil },
	}
	badPostReq := PostRequestMock{
		RequestBodyFunc: func() json.RawMessage { return json.RawMessage(`{"field":"value"}`) },
	}

	// badPostResp := &APIError{
	// 	StatusCode: 400,
	// 	Code:       "not_registered",
	// 	Message:    "Entity is not registered",
	// }

	type args struct {
		ctx      context.Context
		request  any
		response Response
		client   *resty.Client
		url      string
	}

	tests := []struct {
		args    args
		want    error
		name    string
		wantErr bool
	}{
		{
			name:    "invalid client",
			args:    args{ctx: context.TODO(), url: mockServer.URL, response: &ResponseMock{}},
			wantErr: true,
			want:    ErrInvalidClient,
		},
		{
			name: "valid post request",
			args: args{ctx: context.TODO(), client: NewDefaultHTTPClient(mockServer.URL), url: "/goodPost", request: goodPostReq, response: goodPostResp},
			want: nil,
		},
		{
			name: "valid get request",
			args: args{ctx: context.TODO(), client: NewDefaultHTTPClient(mockServer.URL), url: "/goodGet", request: "anything", response: goodGetResp},
			want: nil,
		},
		{
			name: "invalid post request",
			args: args{ctx: context.TODO(), client: NewDefaultHTTPClient(mockServer.URL), url: "/badPost", request: badPostReq, response: &ResponseMock{}},
			want: &APIError{
				StatusCode: 400,
				Message:    "400 Bad Request",
			},
		},
		{
			name: "invalid get request",
			args: args{ctx: context.TODO(), client: NewDefaultHTTPClient(mockServer.URL), url: "/badGet", request: "anything", response: &ResponseMock{}},
			want: &APIError{
				StatusCode: 400,
				Message:    "400 Bad Request",
			},
		},
		// {
		// 	name: "badData",
		// 	args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/badData"), request: badPostReq, response: &ResponseMock{}},
		// 	want: badPostResp,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExecuteRequest(tt.args.ctx, tt.args.client, tt.args.url, tt.args.request, tt.args.response)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewDefaultHTTPClient(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		args args
		want *resty.Client
		name string
	}{
		{
			name: "default",
			args: args{url: "http://localhost:8123"},
			want: resty.New().SetTimeout(defaultTimeout).
				AddRetryCondition(defaultRetry).
				SetBaseURL("http://localhost:8123"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultHTTPClient(tt.args.url); got != nil {
				assert.Equal(t, got.GetClient().Timeout, defaultTimeout)
				assert.Equal(t, got.BaseURL, tt.args.url)
			} else {
				t.Errorf("NewDefaultHTTPClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
