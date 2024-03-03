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
		fields fields
		want   string
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

type mockResponse struct {
	Body any
	err  error
}

func (m *mockResponse) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.Body)
}

func (m *mockResponse) StoreError(e error) {
	m.err = e
}

func (m *mockResponse) Error() string {
	return m.err.Error()
}

func TestExecuteRequest(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/goodPost", r.URL.Path == "/goodGet":
			fmt.Fprintf(w, `{"success": true}`)
		case r.URL.Path == "/badPost", r.URL.Path == "/badGet":
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"success": false}`)
		case r.URL.Path == "/badData":
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"success": false}`)
		}
	}))
	ctx := ContextSetClient(context.TODO(), NewDefaultHTTPClient())

	goodPostReq := PostRequestMock{
		RequestBodyFunc: func() json.RawMessage { return json.RawMessage(`{"field":"value"}`) },
	}

	badPostReq := PostRequestMock{
		RequestBodyFunc: func() json.RawMessage { return json.RawMessage(`{"field":"value"}`) },
	}

	type args struct {
		ctx      context.Context
		request  any
		response Response
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    Response
	}{
		{
			name:    "invalid URL",
			args:    args{ctx: context.TODO(), response: &mockResponse{err: errors.New("")}},
			wantErr: true,
			want:    &mockResponse{err: ErrInvalidURL},
		},
		{
			name:    "invalid Client",
			args:    args{ctx: ContextSetURL(context.TODO(), mockServer.URL), response: &mockResponse{err: errors.New("")}},
			wantErr: true,
			want:    &mockResponse{err: ErrInvalidClient},
		},
		{
			name: "goodPost",
			args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/goodPost"), request: goodPostReq, response: &mockResponse{err: errors.New("")}},
			want: &mockResponse{err: errors.New("")},
		},
		{
			name: "goodGet",
			args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/goodGet"), request: "anything", response: &mockResponse{err: errors.New("")}},
			want: &mockResponse{err: errors.New("")},
		},
		{
			name: "badPost",
			args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/badPost"), request: badPostReq, response: &mockResponse{err: errors.New("")}},
			want: &mockResponse{err: errors.New("400 Bad Request")},
		},
		{
			name: "badGet",
			args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/badGet"), request: "anything", response: &mockResponse{err: errors.New("")}},
			want: &mockResponse{err: errors.New("400 Bad Request")},
		},
		// {
		// 	name: "badData",
		// 	args: args{ctx: ContextSetURL(ctx, mockServer.URL+"/badData"), request: badPostReq, response: &mockResponse{err: errors.New("")}},
		// 	want: &mockResponse{err: errors.New("400 Bad Request")},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExecuteRequest(tt.args.ctx, tt.args.request, tt.args.response)
			assert.Equal(t, tt.args.response.Error(), tt.want.Error())
			// TODO: mock API level responses and test those
		})
	}
}

func TestNewDefaultHTTPClient(t *testing.T) {
	tests := []struct {
		name string
		want *resty.Client
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
