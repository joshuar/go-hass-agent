// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sensor

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
)

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

func TestSensorRegistrationInfo_ResponseHandler(t *testing.T) {
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
	type args struct {
		res    bytes.Buffer
		respCh chan api.Response
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
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
			reg.ResponseHandler(tt.args.res, tt.args.respCh)
		})
	}
}

func TestNewSensorRegistrationResponse(t *testing.T) {
	type args struct {
		r map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want *SensorRegistrationResponse
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSensorRegistrationResponse(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSensorRegistrationResponse() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
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

func TestSensorUpdateInfo_ResponseHandler(t *testing.T) {
	type fields struct {
		StateAttributes interface{}
		State           interface{}
		Icon            string
		Type            string
		UniqueID        string
	}
	type args struct {
		res    bytes.Buffer
		respCh chan api.Response
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
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
			upd.ResponseHandler(tt.args.res, tt.args.respCh)
		})
	}
}

func TestSensorUpdateResponse_Error(t *testing.T) {
	type fields struct {
		err      error
		disabled bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := SensorUpdateResponse{
				err:      tt.fields.err,
				disabled: tt.fields.disabled,
			}
			if err := r.Error(); (err != nil) != tt.wantErr {
				t.Errorf("SensorUpdateResponse.Error() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSensorUpdateResponse_Type(t *testing.T) {
	type fields struct {
		err      error
		disabled bool
	}
	tests := []struct {
		name   string
		fields fields
		want   api.RequestType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := SensorUpdateResponse{
				err:      tt.fields.err,
				disabled: tt.fields.disabled,
			}
			if got := r.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SensorUpdateResponse.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorUpdateResponse_Disabled(t *testing.T) {
	type fields struct {
		err      error
		disabled bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := SensorUpdateResponse{
				err:      tt.fields.err,
				disabled: tt.fields.disabled,
			}
			if got := r.Disabled(); got != tt.want {
				t.Errorf("SensorUpdateResponse.Disabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensorUpdateResponse_Registered(t *testing.T) {
	type fields struct {
		err      error
		disabled bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := SensorUpdateResponse{
				err:      tt.fields.err,
				disabled: tt.fields.disabled,
			}
			if got := r.Registered(); got != tt.want {
				t.Errorf("SensorUpdateResponse.Registered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSensorUpdateResponse(t *testing.T) {
	type args struct {
		i string
		r map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want *SensorUpdateResponse
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSensorUpdateResponse(tt.args.i, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSensorUpdateResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_marshalResponse(t *testing.T) {
	type args struct {
		raw bytes.Buffer
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalResponse(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("marshalResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshalResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_assertAs(t *testing.T) {
	type args struct {
		thing interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assertAs[string](tt.args.thing)
			if (err != nil) != tt.wantErr {
				t.Errorf("assertAs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("assertAs() = %v, want %v", got, tt.want)
			}
		})
	}
}
