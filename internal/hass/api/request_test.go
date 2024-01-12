// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/stretchr/testify/assert"
)

func mockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := &UnencryptedRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.Nil(t, err)
		switch req.Type {
		case "register_sensor":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		case "encrypted":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		}
	}))
}

func Test_marshalJSON(t *testing.T) {
	mockReq := &RequestMock{
		RequestDataFunc: func() json.RawMessage {
			return json.RawMessage(`{"someField": "someValue"}`)
		},
		RequestTypeFunc: func() RequestType {
			return RequestTypeUpdateSensorStates
		},
	}
	mockEncReq := &RequestMock{
		RequestDataFunc: func() json.RawMessage {
			return json.RawMessage(`{"someField": "someValue"}`)
		},
		RequestTypeFunc: func() RequestType {
			return RequestTypeEncrypted
		},
	}

	type args struct {
		request Request
		secret  string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "unencrypted request",
			args: args{request: mockReq},
			want: []byte(`{"type":"update_sensor_states","data":{"someField":"someValue"}}`),
		},
		{
			name:    "encrypted request without secret",
			args:    args{request: mockEncReq},
			want:    nil,
			wantErr: true,
		},
		{
			name: "encrypted request with secret",
			args: args{request: mockEncReq, secret: "fakeSecret"},
			want: []byte(`{"type":"encrypted","encrypted_data":{"someField":"someValue"},"encrypted":true}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalJSON(tt.args.request, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("marshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecuteRequest(t *testing.T) {
	mockServer := mockServer(t)
	defer mockServer.Close()
	cfg, err := config.Load(t.TempDir())
	assert.Nil(t, err)
	err = cfg.Set(config.PrefAPIURL, mockServer.URL)
	assert.Nil(t, err)
	ctx := config.EmbedInContext(context.TODO(), cfg)
	mockReq := &RequestMock{
		RequestDataFunc: func() json.RawMessage {
			return json.RawMessage(`{"someField": "someValue"}`)
		},
		RequestTypeFunc: func() RequestType {
			return RequestTypeUpdateSensorStates
		},
	}
	type args struct {
		ctx     context.Context
		request Request
	}
	tests := []struct {
		name string
		args args
		want chan any
	}{
		{
			name: "default test",
			args: args{ctx: ctx, request: mockReq},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExecuteRequest(tt.args.ctx, tt.args.request)
			// if got := ExecuteRequest(tt.args.ctx, tt.args.request); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("ExecuteRequest() = %v, want %v", got, tt.want)
			// }
		})
	}
}
