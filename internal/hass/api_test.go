// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRequest struct {
	mock.Mock
	requestType RequestType
	data        interface{}
}

func (m *mockRequest) RequestType() RequestType {
	m.On("RequestType")
	m.Called()
	return m.requestType
}

func (m *mockRequest) RequestData() interface{} {
	m.On("RequestData")
	m.Called()
	if m.data == nil {
		return m.String()
	} else {
		return m.data
	}
}

func (m *mockRequest) ResponseHandler(b bytes.Buffer) {
	m.On("ResponseHandler")
	m.Called(b)
}

var unencryptedRequest = &mockRequest{
	requestType: RequestTypeUpdateSensorStates,
}

var encryptedRequest = &mockRequest{
	requestType: RequestTypeEncrypted,
}

func TestMarshalJSON(t *testing.T) {
	unencryptedRequest.data = ""
	unencryptedRequestJSON, err := json.Marshal(&struct {
		Type RequestType `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: RequestTypeUpdateSensorStates,
		Data: "",
	})
	assert.Nil(t, err)
	encryptedRequest.data = ""
	encryptedRequestJSON, err := json.Marshal(&struct {
		Type          RequestType `json:"type"`
		Encrypted     bool        `json:"encrypted"`
		EncryptedData interface{} `json:"encrypted_data"`
	}{
		Type:          RequestTypeEncrypted,
		Encrypted:     true,
		EncryptedData: "",
	})
	assert.Nil(t, err)
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
			args: args{request: unencryptedRequest},
			want: unencryptedRequestJSON,
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
			want: encryptedRequestJSON,
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

func TestAPIRequest(t *testing.T) {
	type args struct {
		ctx     context.Context
		request Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			APIRequest(tt.args.ctx, tt.args.request)
		})
	}
}
