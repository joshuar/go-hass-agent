// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
)

func TestSensorRegistrationInfo_RequestType(t *testing.T) {
	type fields struct {
		State             interface{}
		StateAttributes   interface{}
		UniqueID          string
		Type              string
		Name              string
		UnitOfMeasurement string
		StateClass        string
		EntityCategory    string
		Icon              string
		DeviceClass       string
		Disabled          bool
	}
	tests := []struct {
		name   string
		fields fields
		want   api.RequestType
	}{
		{
			name: "default test",
			want: api.RequestTypeRegisterSensor,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &SensorRegistrationInfo{
				State:             tt.fields.State,
				StateAttributes:   tt.fields.StateAttributes,
				UniqueID:          tt.fields.UniqueID,
				Type:              tt.fields.Type,
				Name:              tt.fields.Name,
				UnitOfMeasurement: tt.fields.UnitOfMeasurement,
				StateClass:        tt.fields.StateClass,
				EntityCategory:    tt.fields.EntityCategory,
				Icon:              tt.fields.Icon,
				DeviceClass:       tt.fields.DeviceClass,
				Disabled:          tt.fields.Disabled,
			}
			if got := reg.RequestType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorRegistrationInfo.RequestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorRegistrationInfo_RequestData(t *testing.T) {
	type fields struct {
		State             interface{}
		StateAttributes   interface{}
		UniqueID          string
		Type              string
		Name              string
		UnitOfMeasurement string
		StateClass        string
		EntityCategory    string
		Icon              string
		DeviceClass       string
		Disabled          bool
	}
	tests := []struct {
		name   string
		fields fields
		want   json.RawMessage
	}{
		{
			name: "successful test",
			fields: fields{
				Name:     "aSensor",
				Type:     "aType",
				State:    "someState",
				UniqueID: "anID",
			},
			want: json.RawMessage(`{"state":"someState","unique_id":"anID","type":"aType","name":"aSensor"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &SensorRegistrationInfo{
				State:             tt.fields.State,
				StateAttributes:   tt.fields.StateAttributes,
				UniqueID:          tt.fields.UniqueID,
				Type:              tt.fields.Type,
				Name:              tt.fields.Name,
				UnitOfMeasurement: tt.fields.UnitOfMeasurement,
				StateClass:        tt.fields.StateClass,
				EntityCategory:    tt.fields.EntityCategory,
				Icon:              tt.fields.Icon,
				DeviceClass:       tt.fields.DeviceClass,
				Disabled:          tt.fields.Disabled,
			}
			if got := reg.RequestData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorRegistrationInfo.RequestData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorUpdateInfo_RequestType(t *testing.T) {
	type fields struct {
		StateAttributes interface{}
		State           interface{}
		Icon            string
		Type            string
		UniqueID        string
	}
	tests := []struct {
		name   string
		fields fields
		want   api.RequestType
	}{
		{
			name: "default test",
			want: api.RequestTypeUpdateSensorStates,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upd := &SensorUpdateInfo{
				StateAttributes: tt.fields.StateAttributes,
				State:           tt.fields.State,
				Icon:            tt.fields.Icon,
				Type:            tt.fields.Type,
				UniqueID:        tt.fields.UniqueID,
			}
			if got := upd.RequestType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorUpdateInfo.RequestType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorUpdateInfo_RequestData(t *testing.T) {
	type fields struct {
		StateAttributes interface{}
		State           interface{}
		Icon            string
		Type            string
		UniqueID        string
	}
	tests := []struct {
		name   string
		fields fields
		want   json.RawMessage
	}{
		{
			name: "successful test",
			fields: fields{
				Type:     "aType",
				State:    "someState",
				UniqueID: "anID",
			},
			want: json.RawMessage(`{"state":"someState","type":"aType","unique_id":"anID"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upd := &SensorUpdateInfo{
				StateAttributes: tt.fields.StateAttributes,
				State:           tt.fields.State,
				Icon:            tt.fields.Icon,
				Type:            tt.fields.Type,
				UniqueID:        tt.fields.UniqueID,
			}
			if got := upd.RequestData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorUpdateInfo.RequestData() = %v, want %v", got, tt.want)
			}
		})
	}
}
