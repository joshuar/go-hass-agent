// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
package hass

import (
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

func Test_request_Validate(t *testing.T) {
	entity := sensor.Entity{
		Name: "Mock Entity",
		State: &sensor.State{
			ID:    "mock_entity",
			Value: "mockState",
		},
	}

	invalidEntity := sensor.Entity{
		Name: "Mock Entity",
		State: &sensor.State{
			ID: "mock_entity",
		},
	}

	type fields struct {
		Data        any
		RequestType string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "valid request",
			fields: fields{RequestType: requestTypeUpdate, Data: entity},
		},
		{
			name:    "invalid request: invalid data",
			fields:  fields{RequestType: requestTypeUpdate, Data: invalidEntity},
			wantErr: true,
		},
		{
			name:    "invalid request: no data",
			fields:  fields{RequestType: requestTypeUpdate},
			wantErr: true,
		},
		{
			name:    "invalid request: unknown request type",
			fields:  fields{Data: entity},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &request{
				Data:        tt.fields.Data,
				RequestType: tt.fields.RequestType,
			}
			if err := r.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("request.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
