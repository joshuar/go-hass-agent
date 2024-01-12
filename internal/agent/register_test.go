// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
)

func TestRegistrationResponse_GenerateAPIURL(t *testing.T) {
	type args struct {
		host string
		resp *api.RegistrationResponse
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid cloudhookurl",
			args: args{
				host: "http://localhost",
				resp: &api.RegistrationResponse{
					CloudhookURL: "http://localhost/cloudhook",
				},
			},
			want: "http://localhost/cloudhook",
		},
		{
			name: "valid remoteuiurl",
			args: args{
				host: "http://localhost",
				resp: &api.RegistrationResponse{
					RemoteUIURL: "http://localhost/remoteuiurl",
					WebhookID:   "foobar",
				},
			},
			want: "http://localhost/remoteuiurl" + api.WebHookPath + "foobar",
		},
		{
			name: "webhookid only",
			args: args{
				host: "http://localhost",
				resp: &api.RegistrationResponse{
					WebhookID: "foobar",
				},
			},
			want: "http://localhost" + api.WebHookPath + "foobar",
		},
		{
			name: "all defined cloudhookurl",
			args: args{
				host: "http://localhost",
				resp: &api.RegistrationResponse{
					CloudhookURL: "http://localhost/cloudhook",
					RemoteUIURL:  "http://localhost/remoteuiurl",
					WebhookID:    "foobar",
				},
			},
			want: "http://localhost/cloudhook",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateAPIURL(tt.args.host, tt.args.resp); got != tt.want {
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
			want: "ws://localhost" + api.WebsocketPath,
		},
		{
			name:   "wss conversion",
			fields: fields{},
			args: args{
				host: "https://localhost",
			},
			want: "wss://localhost" + api.WebsocketPath,
		},
		{
			name:   "ws",
			fields: fields{},
			args: args{
				host: "ws://localhost",
			},
			want: "ws://localhost" + api.WebsocketPath,
		},
		{
			name:   "wss",
			fields: fields{},
			args: args{
				host: "wss://localhost",
			},
			want: "wss://localhost" + api.WebsocketPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateWebsocketURL(tt.args.host); got != tt.want {
				t.Errorf("RegistrationResponse.GenerateWebsocketURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
