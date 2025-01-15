// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//nolint:paralleltest
package tracker

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

var mocksensorEntity = sensor.NewSensor(
	sensor.WithName("Mock sensor.Entity"),
	sensor.WithID("mock_sensor.Entity"),
	sensor.WithState(
		sensor.WithValue("mockValue"),
	),
)

func TestTracker_Get(t *testing.T) {
	mockSensorMap := map[string]*sensor.Entity{
		mocksensorEntity.ID: &mocksensorEntity,
	}

	type fields struct {
		sensor map[string]*sensor.Entity
	}
	type args struct {
		id string
	}
	tests := []struct {
		want    *sensor.Entity
		fields  fields
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "successful get",
			fields:  fields{sensor: mockSensorMap},
			args:    args{id: "mock_sensor.Entity"},
			wantErr: false,
			want:    &mocksensorEntity,
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
	mockSensorMap := map[string]*sensor.Entity{
		mocksensorEntity.ID: &mocksensorEntity,
	}

	type fields struct {
		sensor map[string]*sensor.Entity
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "with sensors",
			fields: fields{sensor: mockSensorMap},
			want:   []string{"mock_sensor.Entity"},
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
	newSensor := sensor.NewSensor(
		sensor.WithName("New sensor.Entity"),
		sensor.WithID("new_sensor.Entity"),
		sensor.WithState(
			sensor.WithValue("new"),
		),
	)

	existingSensor := sensor.NewSensor(
		sensor.WithName("Existing sensor.Entity"),
		sensor.WithID("existing_sensor.Entity"),
		sensor.WithState(
			sensor.WithValue("existing"),
		),
	)

	mockSensorMap := map[string]*sensor.Entity{
		existingSensor.ID: &existingSensor,
	}

	type fields struct {
		sensor map[string]*sensor.Entity
	}
	type args struct {
		sensor *sensor.Entity
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
			args:   args{sensor: &newSensor},
		},
		{
			name:   "existing sensor",
			fields: fields{sensor: mockSensorMap},
			args:   args{sensor: &existingSensor},
		},
		{
			name:    "invalid tracker",
			args:    args{sensor: &newSensor},
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
	mockSensorMap := map[string]*sensor.Entity{
		mocksensorEntity.ID: &mocksensorEntity,
	}

	type fields struct {
		sensor map[string]*sensor.Entity
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
