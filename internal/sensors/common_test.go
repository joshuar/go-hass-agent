// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensors

import (
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/stretchr/testify/mock"
)

type mockSensorUpdate struct {
	mock.Mock
}

func (m *mockSensorUpdate) Attributes() interface{} {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) DeviceClass() hass.SensorDeviceClass {
	args := m.Called()
	return args.Get(0).(hass.SensorDeviceClass)
}

func (m *mockSensorUpdate) Icon() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) State() interface{} {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) SensorType() hass.SensorType {
	args := m.Called()
	return args.Get(0).(hass.SensorType)
}

func (m *mockSensorUpdate) ID() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) Units() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) StateClass() hass.SensorStateClass {
	args := m.Called()
	return args.Get(0).(hass.SensorStateClass)
}

func (m *mockSensorUpdate) Category() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockSensorUpdate) Registered() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockSensorUpdate) Disabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockSensorUpdate) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockSensorUpdate) UnMarshalJSON(b []byte) error {
	args := m.Called(b)
	return args.Error(1)
}
