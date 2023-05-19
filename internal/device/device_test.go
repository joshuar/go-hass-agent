// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDevice struct {
	mock.Mock
}

func (m *mockDevice) DeviceID() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) AppID() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) AppName() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) AppVersion() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) DeviceName() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) Manufacturer() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) Model() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) OsName() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) OsVersion() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDevice) SupportsEncryption() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockDevice) AppData() interface{} {
	args := m.Called()
	return args.String(0)
}

func TestGenerateRegistrationRequest(t *testing.T) {
	device := new(mockDevice)
	device.On("DeviceID").Return("deviceID")
	device.On("AppID").Return("appID")
	device.On("AppName").Return("appName")
	device.On("AppVersion").Return("appVersion")
	device.On("DeviceName").Return("deviceName")
	device.On("Manufacturer").Return("manufacturer")
	device.On("Model").Return("model")
	device.On("OsName").Return("osName")
	device.On("OsVersion").Return("osVersion")
	device.On("SupportsEncryption").Return(false)
	device.On("AppData").Return("")

	deviceReg := &hass.RegistrationRequest{
		DeviceID:           device.DeviceID(),
		AppID:              device.AppID(),
		AppName:            device.AppName(),
		AppVersion:         device.AppVersion(),
		DeviceName:         device.DeviceName(),
		Manufacturer:       device.Manufacturer(),
		Model:              device.Model(),
		OsName:             device.OsName(),
		OsVersion:          device.OsVersion(),
		SupportsEncryption: device.SupportsEncryption(),
		AppData:            device.AppData(),
	}
	type args struct {
		d DeviceInfo
	}
	tests := []struct {
		name string
		args args
		want *hass.RegistrationRequest
	}{
		{
			name: "default test",
			args: args{d: device},
			want: deviceReg,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateRegistrationRequest(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateRegistrationRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSensorInfo(t *testing.T) {
	wanted := &SensorInfo{
		sensorWorkers: make(map[string]func(context.Context, chan interface{})),
	}
	tests := []struct {
		name string
		want *SensorInfo
	}{
		{
			name: "default test",
			want: wanted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSensorInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSensorInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorInfo_Add(t *testing.T) {
	testName := "test"
	testFunc := func(context.Context, chan interface{}) {}
	type args struct {
		name       string
		workerFunc func(context.Context, chan interface{})
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default test",
			args: args{name: testName, workerFunc: testFunc},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewSensorInfo()
			i.Add(tt.args.name, tt.args.workerFunc)
			assert.Contains(t, i.sensorWorkers, testName)
		})
	}
}

func TestSensorInfo_Get(t *testing.T) {
	testName := "test"
	testFunc := func(context.Context, chan interface{}) {}
	testMap := make(map[string]func(context.Context, chan interface{}))
	testMap[testName] = testFunc
	type fields struct {
		sensorsName string
		sensorsFunc func(context.Context, chan interface{})
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]func(context.Context, chan interface{})
	}{
		{
			name:   "test empty",
			fields: fields{},
			want:   make(map[string]func(context.Context, chan interface{})),
		},
		{
			name:   "test exists",
			fields: fields{sensorsName: testName, sensorsFunc: testFunc},
			want:   testMap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewSensorInfo()
			if tt.fields.sensorsName != "" {
				i.Add(tt.fields.sensorsName, tt.fields.sensorsFunc)
			}
			if got := i.Get(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorInfo.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
