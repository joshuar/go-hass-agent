// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
)

func TestRegistrationResponse_GenerateAPIURL(t *testing.T) {
	type args struct {
		host       string
		ignoreURLs bool
		resp       *hass.RegistrationDetails
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
				resp: &hass.RegistrationDetails{
					CloudhookURL: "http://localhost/cloudhook",
				},
			},
			want: "http://localhost/cloudhook",
		},
		{
			name: "valid remoteuiurl",
			args: args{
				host: "http://localhost",
				resp: &hass.RegistrationDetails{
					RemoteUIURL: "http://localhost/remoteuiurl",
					WebhookID:   "foobar",
				},
			},
			want: "http://localhost/remoteuiurl" + hass.WebHookPath + "foobar",
		},
		{
			name: "webhookid only",
			args: args{
				host: "http://localhost",
				resp: &hass.RegistrationDetails{
					WebhookID: "foobar",
				},
			},
			want: "http://localhost" + hass.WebHookPath + "foobar",
		},
		{
			name: "all defined cloudhookurl",
			args: args{
				host: "http://localhost",
				resp: &hass.RegistrationDetails{
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
			if got := generateAPIURL(tt.args.host, tt.args.ignoreURLs, tt.args.resp); got != tt.want {
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
			want: "ws://localhost" + hass.WebsocketPath,
		},
		{
			name:   "wss conversion",
			fields: fields{},
			args: args{
				host: "https://localhost",
			},
			want: "wss://localhost" + hass.WebsocketPath,
		},
		{
			name:   "ws",
			fields: fields{},
			args: args{
				host: "ws://localhost",
			},
			want: "ws://localhost" + hass.WebsocketPath,
		},
		{
			name:   "wss",
			fields: fields{},
			args: args{
				host: "wss://localhost",
			},
			want: "wss://localhost" + hass.WebsocketPath,
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
