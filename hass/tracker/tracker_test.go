// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package tracker

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/sensor"
)

func TestTracker_Get(t *testing.T) {
	mocksensorEntity := sensor.NewSensor(t.Context(),
		sensor.WithName("Mock sensor.Entity"),
		sensor.WithID("mock_sensor.Entity"),
		sensor.WithState("mockValue"),
	)

	mockSensor, err := mocksensorEntity.AsSensor()
	require.NoError(t, err)

	mockSensorMap := map[string]*models.Sensor{
		"mock_sensor.Entity": &mockSensor,
	}

	type fields struct {
		sensor map[string]*models.Sensor
	}
	type args struct {
		id string
	}
	tests := []struct {
		want    *models.Sensor
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
			want:    &mockSensor,
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
	mocksensorEntity := sensor.NewSensor(t.Context(),
		sensor.WithName("Mock sensor.Entity"),
		sensor.WithID("mock_sensor.Entity"),
		sensor.WithState("mockValue"),
	)

	mockSensor, err := mocksensorEntity.AsSensor()
	require.NoError(t, err)

	mockSensorMap := map[string]*models.Sensor{
		"mock_sensor.Entity": &mockSensor,
	}

	type fields struct {
		sensor map[string]*models.Sensor
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
	newEntity := sensor.NewSensor(t.Context(),
		sensor.WithName("New sensor.Entity"),
		sensor.WithID("new_sensor.Entity"),
		sensor.WithState("new"),
	)
	newSensor, err := newEntity.AsSensor()
	require.NoError(t, err)

	existingEntity := sensor.NewSensor(t.Context(),
		sensor.WithName("Existing sensor.Entity"),
		sensor.WithID("existing_sensor.Entity"),
		sensor.WithState("existing"),
	)

	existingSensor, err := existingEntity.AsSensor()
	require.NoError(t, err)

	mockSensorMap := map[string]*models.Sensor{
		"existing_sensor.Entity": &existingSensor,
	}

	type fields struct {
		sensor map[string]*models.Sensor
	}
	type args struct {
		sensor *models.Sensor
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
	mocksensorEntity := sensor.NewSensor(t.Context(),
		sensor.WithName("Mock sensor.Entity"),
		sensor.WithID("mock_sensor.Entity"),
		sensor.WithState("mockValue"),
	)

	mockSensor, err := mocksensorEntity.AsSensor()
	require.NoError(t, err)

	mockSensorMap := map[string]*models.Sensor{
		"mock_sensor.Entity": &mockSensor,
	}

	type fields struct {
		sensor map[string]*models.Sensor
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
