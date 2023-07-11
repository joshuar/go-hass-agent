// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"reflect"
	"testing"
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
	type args struct {
		ctx          context.Context
		registration RegistrationInfo
		device       DeviceInfo
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
			got, err := RegisterWithHass(tt.args.ctx, tt.args.registration, tt.args.device)
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
