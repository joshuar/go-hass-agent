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

type mockSensorUpdate struct {
	mock.Mock
	id         string
	state      interface{}
	icon       string
	attributes interface{}
}

func (m *mockSensorUpdate) Name() string {
	m.On("Name")
	args := m.Called()
	return args.String()
}

func (m *mockSensorUpdate) ID() string {
	m.On("ID")
	args := m.Called()
	if m.id == "" {
		return args.String()
	} else {
		return m.id
	}
}

func (m *mockSensorUpdate) Icon() string {
	m.On("Icon")
	args := m.Called()
	if m.icon == "" {
		return args.String()
	} else {
		return m.icon
	}
}

func (m *mockSensorUpdate) SensorType() hass.SensorType {
	m.On("SensorType")
	m.Called()
	return hass.TypeSensor
}

func (m *mockSensorUpdate) DeviceClass() hass.SensorDeviceClass {
	m.On("DeviceClass")
	m.Called()
	return 0
}

func (m *mockSensorUpdate) StateClass() hass.SensorStateClass {
	m.On("StateClass")
	m.Called()
	return 0
}

func (m *mockSensorUpdate) State() interface{} {
	m.On("State")
	args := m.Called()
	if m.state == nil {
		return args.String()
	} else {
		return m.state
	}

}

func (m *mockSensorUpdate) Units() string {
	m.On("Units")
	args := m.Called()
	return args.String()
}

func (m *mockSensorUpdate) Category() string {
	m.On("Category")
	args := m.Called()
	return args.String()
}

func (m *mockSensorUpdate) Attributes() interface{} {
	m.On("Attributes")
	m.Called()
	return m.attributes
}

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
			args: args{s: &mockSensorUpdate{}},
		},
		{
			name: "unsuccessful add",
			fields: fields{
				registry: newMockSensorRegistry(t)},
			args:    args{s: &mockSensorUpdate{}},
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
	fakeSensorUpdate := &mockSensorUpdate{}
	tracker := newMockSensorTracker(t)
	tracker.add(fakeSensorUpdate)
	fakeSensorState := tracker.Get(fakeSensorUpdate.ID())
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
			args: args{id: fakeSensorUpdate.ID()},
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

func Test_sensorTracker_update(t *testing.T) {
	fakeSensorUpdate := &mockSensorUpdate{}
	fakeSensorStates := make(map[string]*sensorState)
	fakeSensorStates[fakeSensorUpdate.ID()] = marshalSensorState(fakeSensorUpdate)
	type fields struct {
		mu         sync.RWMutex
		sensor     map[string]*sensorState
		registry   *badgerDBRegistry
		hassConfig *hass.HassConfig
	}
	type args struct {
		s hass.SensorUpdate
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "try to update nonexistent sensor",
			fields: fields{sensor: make(map[string]*sensorState)},
			args: args{s: &mockSensorUpdate{
				state: "foo",
				icon:  "bar",
			}},
		},
		{
			name:   "try to update existing sensor",
			fields: fields{sensor: fakeSensorStates},
			args: args{s: &mockSensorUpdate{
				state: "foo",
				icon:  "bar",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &SensorTracker{
				mu:       tt.fields.mu,
				sensor:   tt.fields.sensor,
				registry: tt.fields.registry,
			}
			tracker.update(tt.args.s)
		})
	}
}

func Test_sensorTracker_exists(t *testing.T) {
	fakeSensorUpdate := &mockSensorUpdate{}
	tracker := newMockSensorTracker(t)
	tracker.add(fakeSensorUpdate)

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
			args: args{id: fakeSensorUpdate.ID()},
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
