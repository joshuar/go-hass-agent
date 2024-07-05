// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct,paralleltest
package hass

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
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
