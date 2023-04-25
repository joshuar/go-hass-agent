// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
)

type mockSensor struct {
	mock.Mock
	attributes interface{}
	state      interface{}
	registered bool
	disabled   bool
}

func (m *mockSensor) Attributes() interface{} {
	m.On("Attributes")
	m.Called()
	return m.attributes
}

func (m *mockSensor) DeviceClass() string {
	m.On("DeviceClass")
	args := m.Called()
	return args.String()
}

func (m *mockSensor) Icon() string {
	m.On("Icon")
	args := m.Called()
	return args.String()
}

func (m *mockSensor) Name() string {
	m.On("Name")
	args := m.Called()
	return args.String()
}

func (m *mockSensor) State() interface{} {
	m.On("State")
	m.Called()
	return m.state
}

func (m *mockSensor) Type() string {
	m.On("Type")
	args := m.Called()
	return args.String()
}

func (m *mockSensor) UniqueID() string {
	m.On("UniqueID")
	args := m.Called()
	return args.String()
}

func (m *mockSensor) UnitOfMeasurement() string {
	m.On("UnitOfMeasurement")
	args := m.Called()
	return args.String()
}

func (m *mockSensor) StateClass() string {
	m.On("StateClass")
	args := m.Called()
	return args.String()
}

func (m *mockSensor) EntityCategory() string {
	m.On("EntityCategory")
	args := m.Called()
	return args.String()
}

func (m *mockSensor) Registered() bool {
	m.On("Registered")
	m.Called()
	return m.registered
}

func (m *mockSensor) Disabled() bool {
	m.On("Disabled")
	m.Called()
	return m.disabled
}

func TestMarshalSensorData(t *testing.T) {
	type args struct {
		s Sensor
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "test unregistered sensor",
			args: args{s: &mockSensor{}},
			want: sensorRegistrationInfo{},
		},
		{
			name: "test registered sensor",
			args: args{s: &mockSensor{registered: true}},
			want: []sensorUpdateInfo{{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MarshalSensorData(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalSensorData() = %v, want %v", got, tt.want)
			}
		})
	}
}
