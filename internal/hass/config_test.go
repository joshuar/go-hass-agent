// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

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
		err     error
		Details *ConfigEntries
		mu      sync.Mutex
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
		// TODO: test for error.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				err:     tt.fields.err,
				Details: tt.fields.Details,
				mu:      tt.fields.mu,
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
	assert.Nil(t, err)

	invalidJSON, err := json.Marshal(`{"some":"json"}`)
	assert.Nil(t, err)

	type fields struct {
		err     error
		Details *ConfigEntries
		mu      sync.Mutex
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
				err:     tt.fields.err,
				Details: tt.fields.Details,
				mu:      tt.fields.mu,
			}
			if err := c.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Config.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_StoreError(t *testing.T) {
	type fields struct {
		err     error
		Details *ConfigEntries
		mu      sync.Mutex
	}
	type args struct {
		e error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				err:     tt.fields.err,
				Details: tt.fields.Details,
				mu:      tt.fields.mu,
			}
			c.StoreError(tt.args.e)
		})
	}
}

func TestConfig_Error(t *testing.T) {
	type fields struct {
		err     error
		Details *ConfigEntries
		mu      sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				err:     tt.fields.err,
				Details: tt.fields.Details,
				mu:      tt.fields.mu,
			}
			if got := c.Error(); got != tt.want {
				t.Errorf("Config.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_configRequest_RequestBody(t *testing.T) {
	tests := []struct {
		name string
		c    *configRequest
		want json.RawMessage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &configRequest{}
			if got := c.RequestBody(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("configRequest.RequestBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	assert.Nil(t, err)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(testConfigJSON))
	}))
	preferences.SetPath(t.TempDir())
	preferences.Save(
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
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
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
			if ok, _ := got.IsEntityDisabled("disabledEntity"); ok {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
