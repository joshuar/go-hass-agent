// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/deviceClass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensorType"
	"github.com/joshuar/go-hass-agent/internal/hass/stateClass"
)

func TestMarshalSensorUpdate(t *testing.T) {
	mockSensorUpdate := &SensorMock{
		AttributesFunc: func() interface{} { return nil },
		StateFunc:      func() interface{} { return "aState" },
		IconFunc:       func() string { return "mdi:icon" },
		SensorTypeFunc: func() sensorType.SensorType { return sensorType.TypeSensor },
		IDFunc:         func() string { return "sensorID" },
	}
	type args struct {
		s Sensor
	}
	tests := []struct {
		name string
		args args
		want *hass.SensorUpdateInfo
	}{
		{
			name: "successful marshal",
			args: args{s: mockSensorUpdate},
			want: &hass.SensorUpdateInfo{
				StateAttributes: nil,
				State:           "aState",
				Icon:            "mdi:icon",
				Type:            "sensor",
				UniqueID:        "sensorID",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := marshalSensorUpdate(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalSensorUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalSensorRegistration(t *testing.T) {
	mockSensorUpdate := &SensorMock{
		AttributesFunc:  func() interface{} { return nil },
		StateFunc:       func() interface{} { return "aState" },
		IconFunc:        func() string { return "mdi:icon" },
		SensorTypeFunc:  func() sensorType.SensorType { return sensorType.TypeSensor },
		IDFunc:          func() string { return "sensorID" },
		DeviceClassFunc: func() deviceClass.SensorDeviceClass { return deviceClass.Duration },
		NameFunc:        func() string { return "sensorName" },
		UnitsFunc:       func() string { return "h" },
		StateClassFunc:  func() stateClass.SensorStateClass { return stateClass.StateMeasurement },
		CategoryFunc:    func() string { return "" },
	}

	type args struct {
		s Sensor
	}
	tests := []struct {
		name string
		args args
		want *hass.SensorRegistrationInfo
	}{
		{
			name: "successful marshal",
			args: args{s: mockSensorUpdate},
			want: &hass.SensorRegistrationInfo{
				StateAttributes:   nil,
				State:             "aState",
				Icon:              "mdi:icon",
				Type:              "sensor",
				UniqueID:          "sensorID",
				DeviceClass:       "Duration",
				Name:              "sensorName",
				UnitOfMeasurement: "h",
				StateClass:        "measurement",
				EntityCategory:    "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := marshalSensorRegistration(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalSensorRegistration() = %v, want %v", got, tt.want)
			}
		})
	}
}
