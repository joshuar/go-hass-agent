// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"reflect"
	"sync"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/stretchr/testify/mock"
)

type MockSensorRegistry struct {
	mock.Mock
}

func newMockSensorTracker(t *testing.T) *SensorTracker {
	fakeRegistry := newMockSensorRegistry(t)
	fakeTracker := &SensorTracker{
		sensor:   make(map[string]*sensorState),
		registry: fakeRegistry,
	}
	return fakeTracker
}

func Test_sensorTracker_add(t *testing.T) {
	s := new(mockSensorUpdate)
	s.On("Attributes").Return("")
	s.On("Category").Return("")
	s.On("DeviceClass").Return(hass.Duration)
	s.On("Disabled").Return(false)
	s.On("Registered").Return(true)
	s.On("ID").Return("default")
	s.On("Icon").Return("default")
	s.On("Name").Return("default")
	s.On("SensorType").Return(hass.TypeSensor)
	s.On("StateClass").Return(hass.StateMeasurement)
	s.On("Units").Return("")
	s.On("State").Return("default")
	type fields struct {
		mu         sync.RWMutex
		sensor     map[string]*sensorState
		registry   Registry
		hassConfig *hass.HassConfig
	}
	type args struct {
		s hass.SensorUpdate
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successful add",
			fields: fields{
				registry: newMockSensorRegistry(t),
				sensor:   make(map[string]*sensorState)},
			args: args{s: s},
		},
		{
			name: "unsuccessful add",
			fields: fields{
				registry: newMockSensorRegistry(t)},
			args:    args{s: s},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &SensorTracker{
				mu:       tt.fields.mu,
				sensor:   tt.fields.sensor,
				registry: tt.fields.registry,
			}
			if err := tracker.add(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("sensorTracker.add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sensorTracker_get(t *testing.T) {
	s := new(mockSensorUpdate)
	s.On("Attributes").Return("")
	s.On("Category").Return("")
	s.On("DeviceClass").Return(hass.Duration)
	s.On("Disabled").Return(false)
	s.On("Registered").Return(true)
	s.On("ID").Return("default")
	s.On("Icon").Return("default")
	s.On("Name").Return("default")
	s.On("SensorType").Return(hass.TypeSensor)
	s.On("StateClass").Return(hass.StateMeasurement)
	s.On("Units").Return("")
	s.On("State").Return("default")
	tracker := newMockSensorTracker(t)
	tracker.add(s)
	fakeSensorState := tracker.Get(s.ID())
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want *sensorState
	}{
		{
			name: "existing sensor",
			args: args{id: s.ID()},
			want: fakeSensorState,
		},
		{
			name: "nonexisting sensor",
			args: args{id: "nonexistent"},
			want: nil,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tracker.Get(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensorTracker.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensorTracker_exists(t *testing.T) {
	s := new(mockSensorUpdate)
	s.On("Attributes").Return("")
	s.On("Category").Return("")
	s.On("DeviceClass").Return(hass.Duration)
	s.On("Disabled").Return(false)
	s.On("Registered").Return(true)
	s.On("ID").Return("default")
	s.On("Icon").Return("default")
	s.On("Name").Return("default")
	s.On("SensorType").Return(hass.TypeSensor)
	s.On("StateClass").Return(hass.StateMeasurement)
	s.On("Units").Return("")
	s.On("State").Return("default")

	tracker := newMockSensorTracker(t)
	tracker.add(s)

	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nonexisting sensor",
			args: args{id: "nonexisting"},
			want: false,
		},
		{
			name: "existing sensor",
			args: args{id: s.ID()},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tracker.exists(tt.args.id); got != tt.want {
				t.Errorf("sensorTracker.exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
