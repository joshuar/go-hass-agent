// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest
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

	ctx := ContextSetClient(context.TODO(), NewDefaultHTTPClient())
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
	}

	tests := []struct {
		args    args
		want    error
		name    string
		wantErr bool
	}{
		{
			name:    "invalid URL",
			args:    args{ctx: context.TODO(), response: &ResponseMock{}},
			wantErr: true,
			want:    ErrInvalidURL,
		},
		{
			name:    "invalid Client",
			args:    args{ctx: ContextSetURL(context.TODO(), mockServer.URL), response: &ResponseMock{}},
			wantErr: true,
			want:    ErrInvalidClient,
		},
		{
			name: "goodPost",
			args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/goodPost"), request: goodPostReq, response: goodPostResp},
			want: nil,
		},
		{
			name: "goodGet",
			args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/goodGet"), request: "anything", response: goodGetResp},
			want: nil,
		},
		{
			name: "badPost",
			args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/badPost"), request: badPostReq, response: &ResponseMock{}},
			want: &APIError{
				StatusCode: 400,
				Message:    "400 Bad Request",
			},
		},
		{
			name: "badGet",
			args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/badGet"), request: "anything", response: &ResponseMock{}},
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
			got := ExecuteRequest(tt.args.ctx, tt.args.request, tt.args.response)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewDefaultHTTPClient(t *testing.T) {
	tests := []struct {
		want *resty.Client
		name string
	}{
		{
			name: "default",
			want: resty.New().SetTimeout(defaultTimeout).
				AddRetryCondition(defaultRetry),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultHTTPClient(); got != nil {
				assert.Equal(t, got.GetClient().Timeout, defaultTimeout)
			} else {
				t.Errorf("NewDefaultHTTPClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
