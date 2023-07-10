// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"reflect"
	"testing"

	"fyne.io/fyne/v2/data/binding"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationDetails_Validate(t *testing.T) {
	mockDeviceInfo := &DeviceInfoMock{}

	type fields struct {
		Server string
		Token  string
		Device DeviceInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "hostname",
			fields: fields{
				Server: "localhost",
				Token:  "abcde.abcde_abcde",
				Device: mockDeviceInfo,
			},
			want: false,
		},
		{
			name: "hostname and port",
			fields: fields{
				Server: "localhost:8123",
				Token:  "abcde.abcde_abcde",
				Device: mockDeviceInfo,
			},
			want: false,
		},
		{
			name: "url",
			fields: fields{
				Server: "http://localhost",
				Token:  "abcde.abcde_abcde",
				Device: mockDeviceInfo,
			},
			want: true,
		},
		{
			name: "url with port",
			fields: fields{
				Server: "http://localhost:8123",
				Token:  "abcde.abcde_abcde",
				Device: mockDeviceInfo,
			},
			want: true,
		},
		{
			name: "url with trailing slash",
			fields: fields{
				Server: "http://localhost/",
				Token:  "abcde.abcde_abcde",
				Device: mockDeviceInfo,
			},
			want: true,
		},
		{
			name: "invalid url",
			fields: fields{
				Server: "asdegasg://localhost//",
				Token:  "abcde.abcde_abcde",
				Device: mockDeviceInfo,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			server := binding.NewString()
			err = server.Set(tt.fields.Server)
			assert.Nil(t, err)
			token := binding.NewString()
			err = token.Set(tt.fields.Server)
			assert.Nil(t, err)

			r := &RegistrationDetails{
				Server: server,
				Token:  token,
				Device: tt.fields.Device,
			}
			if got := r.Validate(); got != tt.want {
				t.Errorf("RegistrationDetails.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	type args struct {
		registration *RegistrationDetails
	}
	tests := []struct {
		name    string
		args    args
		want    *RegistrationResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RegisterWithHass(tt.args.registration)
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
