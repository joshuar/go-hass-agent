// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest
package hass

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

func TestConfig_IsEntityDisabled(t *testing.T) {
	testConfig := &ConfigEntries{
		Entities: map[string]map[string]any{
			"disabledEntity": {
				"disabled": true,
			},
			"enabledEntity": {
				"disabled": false,
			},
		},
	}

	type fields struct {
		Details *ConfigEntries
	}

	type args struct {
		entity string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "is disabled",
			args:    args{entity: "disabledEntity"},
			fields:  fields{Details: testConfig},
			want:    true,
			wantErr: false,
		},
		{
			name:    "is enabled",
			args:    args{entity: "enabledEntity"},
			fields:  fields{Details: testConfig},
			want:    false,
			wantErr: false,
		},
		// ?: test for error.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Details: tt.fields.Details,
			}
			got, err := c.IsEntityDisabled(tt.args.entity)

			if (err != nil) != tt.wantErr {
				t.Errorf("Config.IsEntityDisabled() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("Config.IsEntityDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_UnmarshalJSON(t *testing.T) {
	testConfig := &ConfigEntries{}
	testConfigJSON, err := json.Marshal(testConfig)
	require.NoError(t, err)

	invalidJSON, err := json.Marshal(`{"some":"json"}`)
	require.NoError(t, err)

	type fields struct {
		Details *ConfigEntries
	}

	type args struct {
		b []byte
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "successful test",
			args:    args{b: testConfigJSON},
			wantErr: false,
		},
		{
			name:    "unsuccessful test",
			args:    args{b: invalidJSON},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Details: tt.fields.Details,
			}

			if err := c.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Config.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//nolint:containedctx,musttag
func TestGetConfig(t *testing.T) {
	testConfig := &Config{
		Details: &ConfigEntries{
			Entities: map[string]map[string]any{
				"disabledEntity": {
					"disabled": true,
				},
			},
		},
	}
	testConfigJSON, err := json.Marshal(testConfig)
	require.NoError(t, err)

	mockServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, err = fmt.Fprint(response, string(testConfigJSON))
		if err != nil {
			t.Fatal(err)
		}
	}))

	preferences.SetPath(t.TempDir())
	err = preferences.Save(
		preferences.SetHost(mockServer.URL),
		preferences.SetToken("testToken"),
		preferences.SetCloudhookURL(""),
		preferences.SetRemoteUIURL(""),
		preferences.SetWebhookID("testID"),
		preferences.SetSecret(""),
		preferences.SetRestAPIURL(mockServer.URL),
		preferences.SetWebsocketURL(mockServer.URL),
		preferences.SetDeviceName("testDevice"),
		preferences.SetDeviceID("testID"),
		preferences.SetVersion("6.4.0"),
		preferences.SetRegistered(true),
	)
	require.NoError(t, err)

	type args struct {
		ctx context.Context
	}

	tests := []struct {
		args    args
		want    *Config
		name    string
		wantErr bool
	}{
		{
			name:    "successful test",
			args:    args{ctx: context.TODO()},
			want:    testConfig,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfig(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			ok, err := got.IsEntityDisabled("disabledEntity")
			if ok {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}

			require.NoError(t, err)
		})
	}
}
