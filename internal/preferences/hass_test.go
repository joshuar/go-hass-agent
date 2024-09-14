// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:tagalign
package preferences

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreferences_SaveHassPreferences(t *testing.T) {
	ctx := AppIDToContext(context.TODO(), "go-hass-agent-test")
	existingPreferencesDir := t.TempDir()
	existingPreferences := DefaultPreferences(filepath.Join(existingPreferencesDir, "go-hass-agent-test", preferencesFile))
	err := existingPreferences.Save()
	require.NoError(t, err)

	type args struct {
		prefs   *Hass
		options *Registration
		path    string
	}
	tests := []struct {
		args             args
		name             string
		wantAPIURL       string
		wantWebSocketURL string
		wantErr          bool
	}{
		{
			name: "use cloudhook",
			args: args{
				prefs:   &Hass{CloudhookURL: "http://localhost:8123/cloudhook", WebhookID: DefaultSecret},
				options: &Registration{Server: "http://localhost", Token: DefaultSecret},
				path:    t.TempDir(),
			},
			wantAPIURL:       "http://localhost:8123/cloudhook",
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
		{
			name: "use remoteui",
			args: args{
				prefs:   &Hass{RemoteUIURL: "http://localhost:8123/remoteui", WebhookID: DefaultSecret},
				options: &Registration{Server: "http://localhost", Token: DefaultSecret},
				path:    t.TempDir(),
			},
			wantAPIURL:       "http://localhost:8123/remoteui/api/webhook/" + DefaultSecret,
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
		{
			name: "ignoreURLs",
			args: args{
				prefs:   &Hass{RemoteUIURL: "http://localhost:8123/remoteui", WebhookID: DefaultSecret},
				options: &Registration{Server: "http://localhost", Token: DefaultSecret, IgnoreHassURLs: true},
				path:    t.TempDir(),
			},
			wantAPIURL:       "http://localhost/api/webhook/" + DefaultSecret,
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
		{
			name: "no cloudhook or remoteui",
			args: args{
				prefs:   &Hass{WebhookID: DefaultSecret},
				options: &Registration{Server: "http://localhost", Token: DefaultSecret},
				path:    t.TempDir(),
			},
			wantAPIURL:       "http://localhost/api/webhook/" + DefaultSecret,
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
		{
			name: "existing preferences",
			args: args{
				prefs:   &Hass{WebhookID: "newWebHookID"},
				options: &Registration{Server: "http://localhost", Token: "newToken"},
				path:    existingPreferencesDir,
			},
			wantAPIURL:       "http://localhost/api/webhook/newWebHookID",
			wantWebSocketURL: "ws://localhost" + WebsocketPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xdg.ConfigHome = tt.args.path
			p, err := Load(ctx)
			if err != nil && !errors.Is(err, ErrNoPreferences) {
				t.Errorf("Preferences.SaveHassPreferences() error %v", err)
			}
			if err := p.SaveHassPreferences(tt.args.prefs, tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("Preferences.SaveHassPreferences() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.True(t, p.Registered)
			assert.Equal(t, tt.wantAPIURL, p.Hass.RestAPIURL)
			assert.Equal(t, tt.wantWebSocketURL, p.Hass.WebsocketURL)
		})
	}
}

func TestPreferences_generateAPIURL(t *testing.T) {
	type fields struct {
		mu           *sync.Mutex
		MQTT         *MQTT
		Registration *Registration
		Hass         *Hass
		Device       *Device
		Version      string
		file         string
		Registered   bool
	}
	tests := []struct {
		name    string
		want    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid cloudhookurl",
			fields: fields{
				Registration: &Registration{
					Server: "http://localhost",
				},
				Hass: &Hass{
					CloudhookURL: "http://localhost/cloudhook",
				},
			},
			want: "http://localhost/cloudhook",
		},
		{
			name: "valid remoteuiurl",
			fields: fields{
				Registration: &Registration{
					Server: "http://localhost",
				},
				Hass: &Hass{
					RemoteUIURL: "http://localhost/remoteuiurl",
					WebhookID:   "foobar",
				},
			},
			want: "http://localhost/remoteuiurl" + WebHookPath + "foobar",
		},
		{
			name: "webhookid only",
			fields: fields{
				Registration: &Registration{
					Server: "http://localhost",
				},
				Hass: &Hass{
					WebhookID: "foobar",
				},
			},
			want: "http://localhost" + WebHookPath + "foobar",
		},
		{
			name: "all defined cloudhookurl",
			fields: fields{
				Registration: &Registration{
					Server: "http://localhost",
				},
				Hass: &Hass{
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
			p := &Preferences{
				mu:           tt.fields.mu,
				MQTT:         tt.fields.MQTT,
				Registration: tt.fields.Registration,
				Hass:         tt.fields.Hass,
				Device:       tt.fields.Device,
				Version:      tt.fields.Version,
				file:         tt.fields.file,
				Registered:   tt.fields.Registered,
			}
			if err := p.generateAPIURL(); (err != nil) != tt.wantErr {
				t.Errorf("Preferences.generateAPIURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, p.Hass.RestAPIURL)
		})
	}
}

func TestPreferences_generateWebsocketURL(t *testing.T) {
	type fields struct {
		mu           *sync.Mutex
		MQTT         *MQTT
		Registration *Registration
		Hass         *Hass
		Device       *Device
		Version      string
		file         string
		Registered   bool
	}
	tests := []struct {
		name    string
		want    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ws conversion",
			fields: fields{
				Registration: &Registration{Server: "http://localhost"},
				Hass:         &Hass{},
			},
			want: "ws://localhost" + WebsocketPath,
		},
		{
			name: "wss conversion",
			fields: fields{
				Registration: &Registration{Server: "https://localhost"},
				Hass:         &Hass{},
			},
			want: "wss://localhost" + WebsocketPath,
		},
		{
			name: "no ws conversion",
			fields: fields{
				Registration: &Registration{Server: "ws://localhost"},
				Hass:         &Hass{},
			},
			want: "ws://localhost" + WebsocketPath,
		},
		{
			name: "no wss conversion",
			fields: fields{
				Registration: &Registration{Server: "wss://localhost"},
				Hass:         &Hass{},
			},
			want: "wss://localhost" + WebsocketPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Preferences{
				mu:           tt.fields.mu,
				MQTT:         tt.fields.MQTT,
				Registration: tt.fields.Registration,
				Hass:         tt.fields.Hass,
				Device:       tt.fields.Device,
				Version:      tt.fields.Version,
				file:         tt.fields.file,
				Registered:   tt.fields.Registered,
			}
			if err := p.generateWebsocketURL(); (err != nil) != tt.wantErr {
				t.Errorf("Preferences.generateWebsocketURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, p.Hass.WebsocketURL)
		})
	}
}
