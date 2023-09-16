// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
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

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationResponse_GenerateAPIURL(t *testing.T) {
	type fields struct {
		CloudhookURL string
		RemoteUIURL  string
		Secret       string
		WebhookID    string
	}
	type args struct {
		host string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "valid cloudhookurl",
			fields: fields{
				CloudhookURL: "http://localhost/cloudhook",
			},
			args: args{
				host: "http://localhost",
			},
			want: "http://localhost/cloudhook",
		},
		{
			name: "valid remoteuiurl",
			fields: fields{
				RemoteUIURL: "http://localhost/remoteuiurl",
				WebhookID:   "foobar",
			},
			args: args{
				host: "http://localhost",
			},
			want: "http://localhost/remoteuiurl" + webHookPath + "foobar",
		},
		{
			name: "webhookid only",
			fields: fields{
				WebhookID: "foobar",
			},
			args: args{
				host: "http://localhost",
			},
			want: "http://localhost" + webHookPath + "foobar",
		},
		{
			name: "all defined cloudhookurl",
			fields: fields{
				CloudhookURL: "http://localhost/cloudhook",
				RemoteUIURL:  "http://localhost/remoteuiurl",
				WebhookID:    "foobar",
			},
			args: args{
				host: "http://localhost",
			},
			want: "http://localhost/cloudhook",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := &RegistrationResponse{
				CloudhookURL: tt.fields.CloudhookURL,
				RemoteUIURL:  tt.fields.RemoteUIURL,
				Secret:       tt.fields.Secret,
				WebhookID:    tt.fields.WebhookID,
			}
			if got := rr.GenerateAPIURL(tt.args.host); got != tt.want {
				t.Errorf("RegistrationResponse.GenerateAPIURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistrationResponse_GenerateWebsocketURL(t *testing.T) {
	type fields struct {
		CloudhookURL string
		RemoteUIURL  string
		Secret       string
		WebhookID    string
	}
	type args struct {
		host string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "ws conversion",
			fields: fields{},
			args: args{
				host: "http://localhost",
			},
			want: "ws://localhost" + websocketPath,
		},
		{
			name:   "wss conversion",
			fields: fields{},
			args: args{
				host: "https://localhost",
			},
			want: "wss://localhost" + websocketPath,
		},
		{
			name:   "ws",
			fields: fields{},
			args: args{
				host: "ws://localhost",
			},
			want: "ws://localhost" + websocketPath,
		},
		{
			name:   "wss",
			fields: fields{},
			args: args{
				host: "wss://localhost",
			},
			want: "wss://localhost" + websocketPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := &RegistrationResponse{
				CloudhookURL: tt.fields.CloudhookURL,
				RemoteUIURL:  tt.fields.RemoteUIURL,
				Secret:       tt.fields.Secret,
				WebhookID:    tt.fields.WebhookID,
			}
			if got := rr.GenerateWebsocketURL(tt.args.host); got != tt.want {
				t.Errorf("RegistrationResponse.GenerateWebsocketURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterWithHass(t *testing.T) {
	okResponse := &RegistrationResponse{
		CloudhookURL: "someURL",
		RemoteUIURL:  "someURL",
		Secret:       "",
		WebhookID:    "someID",
	}
	okJson, err := json.Marshal(okResponse)
	assert.Nil(t, err)

	newMockServer := func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(okJson)
		}))
	}
	mockServer := newMockServer(t)

	goodRegConfig := &AgentConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case config.PrefHost:
				*v = mockServer.URL
				return nil
			case config.PrefToken:
				*v = "aToken"
				return nil
			default:
				return errors.New("not found")
			}
		},
	}

	badRegServer := &AgentConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case config.PrefHost:
				*v = "notaurl"
				return nil
			case config.PrefToken:
				*v = "aToken"
				return nil
			default:
				return errors.New("not found")
			}
		},
	}

	badRegToken := &AgentConfigMock{
		GetFunc: func(s string, ifaceVal interface{}) error {
			v := ifaceVal.(*string)
			switch s {
			case config.PrefHost:
				*v = mockServer.URL
				return nil
			case config.PrefToken:
				*v = ""
				return nil
			default:
				return errors.New("not found")
			}
		},
	}

	mockDevInfo := &DeviceInfoMock{
		MarshalJSONFunc: func() ([]byte, error) { return []byte(`{"AppName":"aDevice"}`), nil },
	}
	mockBadDevInfo := &DeviceInfoMock{
		MarshalJSONFunc: func() ([]byte, error) { return nil, errors.New("bad device") },
	}

	type args struct {
		ctx       context.Context
		regConfig Agent
		device    DeviceInfo
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
				ctx:       context.Background(),
				regConfig: goodRegConfig,
				device:    mockDevInfo,
			},
			want: okResponse,
		},
		{
			name: "bad device",
			args: args{
				ctx:       context.Background(),
				regConfig: goodRegConfig,
				device:    mockBadDevInfo,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "bad server url",
			args: args{
				ctx:       context.Background(),
				regConfig: badRegServer,
				device:    mockDevInfo,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "bad token",
			args: args{
				ctx:       context.Background(),
				regConfig: badRegToken,
				device:    mockDevInfo,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RegisterWithHass(tt.args.ctx, tt.args.regConfig, tt.args.device)
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
