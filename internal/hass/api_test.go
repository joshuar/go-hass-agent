// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/stretchr/testify/mock"
)

type mockRequest struct {
	mock.Mock
}

func (m *mockRequest) RequestType() RequestType {
	args := m.Called()
	return args.Get(0).(RequestType)
}

func (m *mockRequest) RequestData() *json.RawMessage {
	args := m.Called()
	return args.Get(0).(*json.RawMessage)
}

func (m *mockRequest) ResponseHandler(b bytes.Buffer) {
	m.Called(b)
}

func TestMarshalJSON(t *testing.T) {
	requestData := json.RawMessage(`{"someField": "someValue"}`)
	request := new(mockRequest)
	request.On("RequestType").Return(RequestTypeUpdateSensorStates)
	request.On("RequestData").Return(&requestData)

	encryptedRequest := new(mockRequest)
	encryptedRequest.On("RequestType").Return(RequestTypeEncrypted)
	encryptedRequest.On("RequestData").Return(&requestData)

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
			want: []byte(`{"data":{"someField":"someValue"},"type":"update_sensor_states"}`),
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
			want: []byte(`{"encrypted_data":{"someField":"someValue"},"type":"encrypted","encrypted":true}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalJSON(tt.args.request, tt.args.secret)
			spew.Dump(MarshalJSON(tt.args.request, tt.args.secret))
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

type mockConfig struct {
	mock.Mock
}

func (m *mockConfig) Get(property string) (interface{}, error) {
	args := m.Called(property)
	return args.String(0), args.Error(1)
}

func (m *mockConfig) Set(property string, value interface{}) error {
	args := m.Called(property, value)
	return args.Error(1)
}

func (m *mockConfig) Validate() error {
	args := m.Called()
	return args.Error(1)
}

func (m *mockConfig) Refresh() error {
	args := m.Called()
	return args.Error(1)
}

func TestAPIRequest(t *testing.T) {
	server := mockServer(t)
	defer server.Close()

	mockConfig := &mockConfig{}
	mockConfig.On("Get", "apiURL").Return(server.URL, nil)
	mockConfig.On("Get", "secret").Return("", nil)

	mockCtx := config.StoreInContext(context.Background(), mockConfig)

	requestData := json.RawMessage(`{"someField": "someValue"}`)
	request := new(mockRequest)
	request.On("RequestType").Return(RequestTypeUpdateSensorStates)
	request.On("RequestData").Return(&requestData)
	request.On("ResponseHandler", *bytes.NewBufferString(`{"success":true}`)).Return()

	type args struct {
		ctx     context.Context
		request Request
	}
	tests := []struct {
		args args
		name string
	}{
		{
			name: "successful test",
			args: args{
				ctx:     mockCtx,
				request: Request(request),
			},
		},
		{
			name: "invalid context",
			args: args{
				ctx:     context.Background(),
				request: Request(request),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			APIRequest(tt.args.ctx, tt.args.request)
		})
	}
}
