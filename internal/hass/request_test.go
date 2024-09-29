// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:max-public-structs
package hass

import (
	"reflect"
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

func Test_newEntityRequest(t *testing.T) {
	locationEntity := sensor.Entity{
		State: &sensor.State{
			Value: &LocationRequest{
				Gps: []float64{0.0, 0.0},
			},
		},
	}

	entity := sensor.Entity{
		Name: "Mock Entity",
		State: &sensor.State{
			ID: "mock_entity",
		},
	}

	type args struct {
		requestType string
		entity      sensor.Entity
	}
	tests := []struct {
		want    *request
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "location request",
			args: args{requestType: requestTypeLocation, entity: locationEntity},
			want: &request{Data: locationEntity.Value, RequestType: requestTypeLocation},
		},
		{
			name: "update request",
			args: args{requestType: requestTypeUpdate, entity: entity},
			want: &request{Data: entity.State, RequestType: requestTypeUpdate},
		},
		{
			name: "registration request",
			args: args{requestType: requestTypeRegister, entity: entity},
			want: &request{Data: entity, RequestType: requestTypeRegister},
		},
		{
			name:    "no request type",
			args:    args{entity: entity},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newEntityRequest(tt.args.requestType, tt.args.entity)
			if (err != nil) != tt.wantErr {
				t.Errorf("newEntityRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newEntityRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
