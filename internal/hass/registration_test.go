// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hass

import (
	"testing"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

func Test_generateAPIURL(t *testing.T) {
	webhookID := "testID"
	token := "testToken"

	type args struct {
		response *deviceRegistrationResponse
		request  *preferences.Registration
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "use cloudhook",
			args: args{
				response: &deviceRegistrationResponse{CloudhookURL: "http://localhost:8123/cloudhook", WebhookID: webhookID},
				request:  &preferences.Registration{Server: "http://localhost", Token: token},
			},
			want: "http://localhost:8123/cloudhook",
		},
		{
			name: "use remoteui",
			args: args{
				response: &deviceRegistrationResponse{RemoteUIURL: "http://localhost:8123/remoteui", WebhookID: webhookID},
				request:  &preferences.Registration{Server: "http://localhost", Token: token},
			},
			want: "http://localhost:8123/remoteui/api/webhook/" + webhookID,
		},
		{
			name: "ignoreURLs",
			args: args{
				response: &deviceRegistrationResponse{RemoteUIURL: "http://localhost:8123/remoteui", WebhookID: webhookID},
				request:  &preferences.Registration{Server: "http://localhost", Token: token, IgnoreHassURLs: true},
			},
			want: "http://localhost/api/webhook/" + webhookID,
		},
		{
			name: "no cloudhook or remoteui",
			args: args{
				response: &deviceRegistrationResponse{WebhookID: webhookID},
				request:  &preferences.Registration{Server: "http://localhost", Token: token},
			},
			want: "http://localhost/api/webhook/" + webhookID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateAPIURL(tt.args.response, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateAPIURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateAPIURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateWebsocketURL(t *testing.T) {
	type args struct {
		server string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "http",
			args: args{server: "http://localhost"},
			want: "ws://localhost" + WebsocketPath,
		},
		{
			name: "https",
			args: args{server: "https://localhost"},
			want: "wss://localhost" + WebsocketPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateWebsocketURL(tt.args.server)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateWebsocketURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateWebsocketURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
