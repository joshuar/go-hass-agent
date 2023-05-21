// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
)

type mockSensor struct {
	mock.Mock
}

func (m *mockSensor) Attributes() interface{} {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) DeviceClass() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) Icon() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) State() interface{} {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) Type() string {
	args := m.Called()
	return args.String()
}

func (m *mockSensor) UniqueID() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) UnitOfMeasurement() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) StateClass() string {
	args := m.Called()
	return args.String()
}

func (m *mockSensor) EntityCategory() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) Registered() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockSensor) Disabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestMarshalSensorData(t *testing.T) {
	registeredSensor := new(mockSensor)
	registeredSensor.On("Attributes").Return("aString")
	registeredSensor.On("DeviceClass").Return("aString")
	registeredSensor.On("Disabled").Return(false)
	registeredSensor.On("EntityCategory").Return("aString")
	registeredSensor.On("Icon").Return("aString")
	registeredSensor.On("Name").Return("aString")
	registeredSensor.On("Registered").Return(true)
	registeredSensor.On("State").Return("aString")
	registeredSensor.On("StateClass").Return("aString")
	registeredSensor.On("Type").Return("aString")
	registeredSensor.On("UniqueID").Return("aString")
	registeredSensor.On("UnitOfMeasurement").Return("aString")

	unregisterdSensor := new(mockSensor)
	unregisterdSensor.On("Attributes").Return("aString")
	unregisterdSensor.On("DeviceClass").Return("aString")
	unregisterdSensor.On("Disabled").Return(false)
	unregisterdSensor.On("EntityCategory").Return("aString")
	unregisterdSensor.On("Icon").Return("aString")
	unregisterdSensor.On("Name").Return("aString")
	unregisterdSensor.On("Registered").Return(false)
	unregisterdSensor.On("State").Return("aString")
	unregisterdSensor.On("StateClass").Return("aString")
	unregisterdSensor.On("Type").Return("aString")
	unregisterdSensor.On("UniqueID").Return("aString")
	unregisterdSensor.On("UnitOfMeasurement").Return("aString")

	unregistered := json.RawMessage(`{"attributes":"aString","device_class":"aString","icon":"aString","name":"aString","state":"aString","type":"string","unique_id":"aString","unit_of_measurement":"aString","state_class":"string","entity_category":"aString"}`)
	registered := json.RawMessage(`[{"attributes":"aString","icon":"aString","state":"aString","type":"string","unique_id":"aString"}]`)
	type args struct {
		s Sensor
	}
	tests := []struct {
		name string
		args args
		want *json.RawMessage
	}{
		{
			name: "test unregistered sensor",
			args: args{s: unregisterdSensor},
			want: &unregistered,
		},
		{
			name: "test registered sensor",
			args: args{s: registeredSensor},
			want: &registered,
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
