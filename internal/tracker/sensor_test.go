// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tracker

import (
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/deviceClass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensorType"
	"github.com/joshuar/go-hass-agent/internal/hass/stateClass"
	"github.com/joshuar/go-hass-agent/internal/tracker/mocks"
)

func TestMarshalSensorUpdate(t *testing.T) {
	validUpdate := mocks.NewSensor(t)
	validUpdate.On("State").Return("state")
	validUpdate.On("Attributes").Return("attributes")
	validUpdate.On("Icon").Return("icon")
	validUpdate.On("SensorType").Return(sensorType.SensorType(0))
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
	validUpdate := mocks.NewSensor(t)
	validUpdate.On("State").Return("state")
	validUpdate.On("Attributes").Return("attributes")
	validUpdate.On("Icon").Return("icon")
	validUpdate.On("SensorType").Return(sensorType.SensorType(0))
	validUpdate.On("ID").Return("uniqueid")
	validUpdate.On("Name").Return("name")
	validUpdate.On("DeviceClass").Return(deviceClass.Duration)
	validUpdate.On("StateClass").Return(stateClass.SensorStateClass(0))
	validUpdate.On("Units").Return("")
	validUpdate.On("Category").Return("")
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
