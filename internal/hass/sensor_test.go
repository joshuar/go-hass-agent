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

func TestMarshalSensorUpdate(t *testing.T) {
	validUpdate := new(mockSensor)
	validUpdate.On("State").Return("state")
	validUpdate.On("Attributes").Return("attributes")
	validUpdate.On("Icon").Return("icon")
	validUpdate.On("SensorType").Return(SensorType(0))
	validUpdate.On("ID").Return("uniqueid")
	type args struct {
		s Sensor
	}
	tests := []struct {
		args args
		want *sensorUpdateInfo
		name string
	}{
		{
			name: "valid update",
			args: args{s: validUpdate},
			want: &sensorUpdateInfo{
				StateAttributes: "attributes",
				State:           "state",
				Icon:            "icon",
				UniqueID:        "uniqueid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MarshalSensorUpdate(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalSensorUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalSensorRegistration(t *testing.T) {
	validUpdate := new(mockSensor)
	validUpdate.On("State").Return("state")
	validUpdate.On("Attributes").Return("attributes")
	validUpdate.On("Icon").Return("icon")
	validUpdate.On("SensorType").Return(SensorType(0))
	validUpdate.On("ID").Return("uniqueid")
	validUpdate.On("Name").Return("name")
	validUpdate.On("DeviceClass").Return(Duration)
	validUpdate.On("StateClass").Return(SensorStateClass(0))
	validUpdate.On("Units").Return("")
	validUpdate.On("Category").Return("")
	validUpdate.On("Disabled").Return(false)
	validUpdate.On("Registered").Return(false)
	type args struct {
		s Sensor
	}
	tests := []struct {
		args args
		want *sensorRegistrationInfo
		name string
	}{
		{
			name: "valid registration",
			args: args{s: validUpdate},
			want: &sensorRegistrationInfo{
				StateAttributes: "attributes",
				DeviceClass:     "Duration",
				Icon:            "icon",
				Name:            "name",
				State:           "state",
				UniqueID:        "uniqueid",
				Disabled:        false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MarshalSensorRegistration(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalSensorRegistration() = %v, want %v", got, tt.want)
			}
		})
	}
}
