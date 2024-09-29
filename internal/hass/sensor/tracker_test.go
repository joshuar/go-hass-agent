// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package sensor

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracker_Get(t *testing.T) {
	mockEntity := &Entity{
		Name: "Mock Entity",
		State: &State{
			ID: "mock_entity",
		},
	}

	mockSensorMap := map[string]*Entity{
		mockEntity.ID: mockEntity,
	}

	type fields struct {
		sensor map[string]*Entity
	}
	type args struct {
		id string
	}
	tests := []struct {
		want    *Entity
		fields  fields
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "successful get",
			fields:  fields{sensor: mockSensorMap},
			args:    args{id: "mock_entity"},
			wantErr: false,
			want:    mockEntity,
		},
		{
			name:    "unsuccessful get",
			fields:  fields{sensor: mockSensorMap},
			args:    args{id: "doesntExist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
			}
			got, err := tr.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tracker.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tracker.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTracker_SensorList(t *testing.T) {
	mockEntity := &Entity{
		Name: "Mock Entity",
		State: &State{
			ID: "mock_entity",
		},
	}

	mockSensorMap := map[string]*Entity{
		mockEntity.ID: mockEntity,
	}

	type fields struct {
		sensor map[string]*Entity
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "with sensors",
			fields: fields{sensor: mockSensorMap},
			want:   []string{"mock_entity"},
		},
		{
			name: "without sensors",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
			}
			if got := tr.SensorList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tracker.SensorList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTracker_Add(t *testing.T) {
	newEntity := &Entity{
		Name: "New Entity",
		State: &State{
			ID: "new_entity",
		},
	}

	existingEntity := &Entity{
		Name: "Existing Entity",
		State: &State{
			ID: "existing_entity",
		},
	}

	mockSensorMap := map[string]*Entity{
		existingEntity.ID: existingEntity,
	}

	type fields struct {
		sensor map[string]*Entity
	}
	type args struct {
		sensor *Entity
	}
	tests := []struct {
		args    args
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name:   "new sensor",
			fields: fields{sensor: mockSensorMap},
			args:   args{sensor: newEntity},
		},
		{
			name:   "existing sensor",
			fields: fields{sensor: mockSensorMap},
			args:   args{sensor: existingEntity},
		},
		{
			name:    "invalid tracker",
			args:    args{sensor: newEntity},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
			}
			if err := tr.Add(tt.args.sensor); (err != nil) != tt.wantErr {
				t.Errorf("Tracker.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTracker_Reset(t *testing.T) {
	mockEntity := &Entity{
		Name: "Mock Entity",
		State: &State{
			ID: "mock_entity",
		},
	}

	mockSensorMap := map[string]*Entity{
		mockEntity.ID: mockEntity,
	}

	type fields struct {
		sensor map[string]*Entity
	}
	tests := []struct {
		fields fields
		name   string
	}{
		{
			name:   "reset",
			fields: fields{sensor: mockSensorMap},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tracker{
				sensor: tt.fields.sensor,
			}
			tr.Reset()
			assert.Nil(t, tr.sensor)
		})
	}
}
