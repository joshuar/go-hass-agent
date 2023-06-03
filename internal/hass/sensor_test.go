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

func (m *mockSensor) DeviceClass() SensorDeviceClass {
	args := m.Called()
	return args.Get(0).(SensorDeviceClass)
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

func (m *mockSensor) SensorType() SensorType {
	args := m.Called()
	return args.Get(0).(SensorType)
}

func (m *mockSensor) ID() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) Units() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensor) StateClass() SensorStateClass {
	args := m.Called()
	return args.Get(0).(SensorStateClass)
}

func (m *mockSensor) Category() string {
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

func (m *mockSensor) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockSensor) UnMarshalJSON(b []byte) error {
	args := m.Called(b)
	return args.Error(1)
}

func TestMarshalSensorData(t *testing.T) {
	registeredSensor := new(mockSensor)
	registeredSensor.On("Attributes").Return("aString")
	registeredSensor.On("DeviceClass").Return(Duration)
	registeredSensor.On("Disabled").Return(false)
	registeredSensor.On("Category").Return("aString")
	registeredSensor.On("Icon").Return("aString")
	registeredSensor.On("Name").Return("aString")
	registeredSensor.On("Registered").Return(true)
	registeredSensor.On("State").Return("aString")
	registeredSensor.On("StateClass").Return(SensorStateClass(0))
	registeredSensor.On("SensorType").Return(SensorType(0))
	registeredSensor.On("ID").Return("aString")
	registeredSensor.On("Units").Return("aString")

	unregisterdSensor := new(mockSensor)
	unregisterdSensor.On("Attributes").Return("aString")
	unregisterdSensor.On("DeviceClass").Return(Duration)
	unregisterdSensor.On("Disabled").Return(false)
	unregisterdSensor.On("Category").Return("aString")
	unregisterdSensor.On("Icon").Return("aString")
	unregisterdSensor.On("Name").Return("aString")
	unregisterdSensor.On("Registered").Return(false)
	unregisterdSensor.On("State").Return("aString")
	unregisterdSensor.On("StateClass").Return(SensorStateClass(0))
	unregisterdSensor.On("SensorType").Return(SensorType(0))
	unregisterdSensor.On("ID").Return("aString")
	unregisterdSensor.On("Units").Return("aString")

	unregistered := json.RawMessage(`{"attributes":"aString","device_class":"Duration","icon":"aString","name":"aString","state":"aString","type":"sensor","unique_id":"aString","unit_of_measurement":"aString","entity_category":"aString"}`)
	registered := json.RawMessage(`[{"attributes":"aString","icon":"aString","state":"aString","type":"sensor","unique_id":"aString"}]`)
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
