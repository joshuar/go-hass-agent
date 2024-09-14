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
	sensor, _, _ := newMockDetails(t)

	mockSensorMap := map[string]Details{
		sensor.ID(): sensor,
	}

	type fields struct {
		sensor map[string]Details
	}
	type args struct {
		id string
	}
	tests := []struct {
		want    Details
		fields  fields
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "successful get",
			fields:  fields{sensor: mockSensorMap},
			args:    args{id: sensor.ID()},
			wantErr: false,
			want:    sensor,
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
	sensor, _, _ := newMockDetails(t)

	mockSensorMap := map[string]Details{
		sensor.ID(): sensor,
	}

	type fields struct {
		sensor map[string]Details
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "with sensors",
			fields: fields{sensor: mockSensorMap},
			want:   []string{sensor.ID()},
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
	sensor, _, _ := newMockDetails(t)

	mockSensorMap := map[string]Details{
		sensor.ID(): sensor,
	}

	type fields struct {
		sensor map[string]Details
	}
	type args struct {
		sensor Details
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
			args:   args{sensor: sensor},
		},
		{
			name:   "existing sensor",
			fields: fields{sensor: mockSensorMap},
			args:   args{sensor: sensor},
		},
		{
			name:    "invalid tracker",
			args:    args{sensor: sensor},
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
	sensor, _, _ := newMockDetails(t)

	mockSensorMap := map[string]Details{
		sensor.ID(): sensor,
	}

	type fields struct {
		sensor map[string]Details
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
