// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/preferences"
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

func TestAgent_performRegistration(t *testing.T) {
	preferences.SetPath(t.TempDir())

	mockGoodReponse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse, err := json.Marshal(&hass.RegistrationDetails{WebhookID: "someID"})
		assert.Nil(t, err)
		fmt.Fprintf(w, string(mockResponse))
	}))
	mockBadResponse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse, err := json.Marshal(&hass.RegistrationDetails{})
		assert.Nil(t, err)
		fmt.Fprintf(w, string(mockResponse))
	}))

	type fields struct {
		ui      UI
		done    chan struct{}
		Options *Options
	}
	type args struct {
		ctx    context.Context
		server string
		token  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "successful test",
			args:   args{ctx: context.Background(), server: mockGoodReponse.URL, token: "someToken"},
			fields: fields{Options: &Options{Headless: true}},
		},
		{
			name:    "missing server",
			args:    args{ctx: context.Background(), token: "someToken"},
			fields:  fields{Options: &Options{Headless: true}},
			wantErr: true,
		},
		{
			name:    "missing token",
			args:    args{ctx: context.Background(), server: mockGoodReponse.URL},
			fields:  fields{Options: &Options{Headless: true}},
			wantErr: true,
		},
		{
			name:    "bad response",
			args:    args{ctx: context.Background(), server: mockBadResponse.URL},
			fields:  fields{Options: &Options{Headless: true}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ui:      tt.fields.ui,
				done:    tt.fields.done,
				Options: tt.fields.Options,
			}
			if err := agent.performRegistration(tt.args.ctx, tt.args.server, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("Agent.performRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
