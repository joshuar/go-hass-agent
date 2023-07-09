// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	bytes "bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/config"
)

func TestMarshalJSON(t *testing.T) {
	requestData := json.RawMessage(`{"someField": "someValue"}`)
	request := NewMockRequest(t)
	request.On("RequestType").Return(RequestTypeUpdateSensorStates)
	request.On("RequestData").Return(requestData)

	encryptedRequest := NewMockRequest(t)
	encryptedRequest.On("RequestType").Return(RequestTypeEncrypted)
	encryptedRequest.On("RequestData").Return(requestData)

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
			args: args{request: request},
			want: []byte(`{"type":"update_sensor_states","data":{"someField":"someValue"}}`),
		},
		{
			name:    "encrypted request without secret",
			args:    args{request: encryptedRequest},
			want:    nil,
			wantErr: true,
		},
		{
			name: "encrypted request with secret",
			args: args{request: encryptedRequest, secret: "fakeSecret"},
			want: []byte(`{"type":"encrypted","encrypted_data":{"someField":"someValue"},"encrypted":true}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalJSON(tt.args.request, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
}

func TestExecuteRequest(t *testing.T) {
	server := mockServer(t)
	defer server.Close()

	goodConfig := config.NewMockConfig(t)
	goodConfig.On("Get", "apiURL").Return(server.URL, nil)
	goodConfig.On("Get", "secret").Return("", nil)
	goodCtx := config.StoreInContext(context.Background(), goodConfig)

	responseCh := make(chan Response, 1)
	defer close(responseCh)

	requestData := json.RawMessage(`{"someField": "someValue"}`)
	request := NewMockRequest(t)
	request.On("RequestType").Return(RequestTypeRegisterSensor)
	request.On("RequestData").Return(requestData)
	request.On("ResponseHandler", *bytes.NewBufferString(`{"success":true}`), responseCh).Return()

	encryptedRequest := NewMockRequest(t)
	encryptedRequest.On("RequestType").Return(RequestTypeEncrypted)
	encryptedRequest.On("RequestData").Return(requestData)
	encryptedRequest.On("ResponseHandler", *bytes.NewBufferString(`{"success":true}`), responseCh).Return()

	type args struct {
		ctx        context.Context
		request    Request
		responseCh chan Response
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "successful request",
			args: args{
				ctx:        goodCtx,
				request:    request,
				responseCh: responseCh,
			},
		},
		{
			name: "bad encrypted request",
			args: args{
				ctx:        goodCtx,
				request:    encryptedRequest,
				responseCh: responseCh,
			},
		},
		{
			name: "bad context",
			args: args{
				ctx:        context.Background(),
				request:    request,
				responseCh: make(chan Response, 1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExecuteRequest(tt.args.ctx, tt.args.request, tt.args.responseCh)
		})
	}
}
