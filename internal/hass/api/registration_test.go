// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterWithHass(t *testing.T) {
	okResponse := &RegistrationResponse{
		CloudhookURL: "someURL",
		RemoteUIURL:  "someURL",
		Secret:       "",
		WebhookID:    "someID",
	}
	okJSON, err := json.Marshal(okResponse)
	assert.Nil(t, err)

	notokResponse := &RegistrationResponse{
		CloudhookURL: "",
		RemoteUIURL:  "",
		Secret:       "",
		WebhookID:    "",
	}
	notokJSON, err := json.Marshal(notokResponse)
	assert.Nil(t, err)

	newMockServer := func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(authHeader)
			if token != "Bearer" {
				w.WriteHeader(http.StatusOK)
				w.Write(okJSON)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(notokJSON)
			}
		}))
	}
	mockServer := newMockServer(t)

	mockDevInfo := &DeviceInfoMock{
		MarshalJSONFunc: func() ([]byte, error) { return []byte(`{"AppName":"aDevice"}`), nil },
	}
	mockBadDevInfo := &DeviceInfoMock{
		MarshalJSONFunc: func() ([]byte, error) { return nil, errors.New("bad device") },
	}

	type args struct {
		ctx    context.Context
		server string
		token  string
		device DeviceInfo
	}
	tests := []struct {
		name    string
		args    args
		want    *RegistrationResponse
		wantErr bool
	}{
		{
			name: "successful test",
			args: args{
				ctx:    context.Background(),
				server: mockServer.URL,
				token:  "aToken",
				device: mockDevInfo,
			},
			want: okResponse,
		},
		{
			name: "bad device",
			args: args{
				ctx:    context.Background(),
				server: mockServer.URL,
				token:  "aToken",
				device: mockBadDevInfo,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "bad server url",
			args: args{
				ctx:    context.Background(),
				server: "notAURL",
				token:  "aToken",
				device: mockDevInfo,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "bad token",
			args: args{
				ctx:    context.Background(),
				server: mockServer.URL,
				token:  "",
				device: mockDevInfo,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RegisterWithHass(tt.args.ctx, tt.args.server, tt.args.token, tt.args.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterWithHass() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegisterWithHass() = %v, want %v", got, tt.want)
			}
		})
	}
}
