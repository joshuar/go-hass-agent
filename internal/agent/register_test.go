// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest,wsl
//revive:disable:unused-receiver
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func TestRegistrationResponse_GenerateAPIURL(t *testing.T) {
	type args struct {
		resp       *hass.RegistrationDetails
		host       string
		ignoreURLs bool
	}
	tests := []struct {
		name    string
		want    string
		args    args
		wantErr bool
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
			got, err := generateAPIURL(tt.args.host, tt.args.ignoreURLs, tt.args.resp)
			if got != tt.want {
				t.Errorf("RegistrationResponse.GenerateAPIURL() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)

				return
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
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
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
			got, err := generateWebsocketURL(tt.args.host)
			if got != tt.want {
				t.Errorf("RegistrationResponse.GenerateWebsocketURL() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}

//nolint:containedctx
//revive:disable:function-length
func TestAgent_performRegistration(t *testing.T) {
	preferences.SetPath(t.TempDir())
	// set a fake version as it normally gets generated on build.
	preferences.AppVersion = "v0.0.0"

	mockGoodReponse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mockResponse, err := json.Marshal(&hass.RegistrationDetails{WebhookID: "someID"})
		if err != nil {
			t.Fatal(err)
		}
		_, err = fmt.Fprint(w, string(mockResponse))
		if err != nil {
			t.Fatal(err)
		}
	}))
	mockBadResponse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mockResponse, err := json.Marshal(&hass.RegistrationDetails{})
		if err != nil {
			t.Fatal(err)
		}
		_, err = fmt.Fprint(w, string(mockResponse))
		if err != nil {
			t.Fatal(err)
		}
	}))

	type args struct {
		ctx    context.Context
		server string
		token  string
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "successful test",
			args: args{
				ctx:    context.Background(),
				server: mockGoodReponse.URL,
				token:  "someToken",
			},
		},
		{
			name: "missing server",
			args: args{
				ctx:   context.Background(),
				token: "someToken",
			},
			wantErr: true,
		},
		{
			name: "missing token",
			args: args{
				ctx:    context.Background(),
				server: mockGoodReponse.URL,
			},
			wantErr: true,
		},
		{
			name: "bad response",
			args: args{
				ctx:    context.Background(),
				server: mockBadResponse.URL,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(
				Headless(true),
				WithRegistrationInfo(tt.args.server, tt.args.token, false),
			)
			if err := agent.performRegistration(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Agent.performRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
