// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package linux

import (
	"reflect"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

func TestSensor_Name(t *testing.T) {
	type fields struct {
		DisplayName      string
		Value            any
		IconString       string
		UnitsString      string
		SensorSrc        string
		IsBinary         bool
		IsDiagnostic     bool
		DeviceClassValue types.DeviceClass
		StateClassValue  types.StateClass
	}

	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "known sensor type",
			fields: fields{DisplayName: "Active App"},
			want:   "Active App",
		},
		{
			name: "unset sensor type",
			want: "Unknown Sensor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &Sensor{
				Value:            tt.fields.Value,
				IconString:       tt.fields.IconString,
				UnitsString:      tt.fields.UnitsString,
				DataSource:       tt.fields.SensorSrc,
				DisplayName:      tt.fields.DisplayName,
				IsBinary:         tt.fields.IsBinary,
				IsDiagnostic:     tt.fields.IsDiagnostic,
				DeviceClassValue: tt.fields.DeviceClassValue,
				StateClassValue:  tt.fields.StateClassValue,
			}
			if got := sensor.Name(); got != tt.want {
				t.Errorf("Sensor.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensor_ID(t *testing.T) {
	type fields struct {
		Value            any
		IconString       string
		UnitsString      string
		SensorSrc        string
		DisplayName      string
		IsBinary         bool
		IsDiagnostic     bool
		DeviceClassValue types.DeviceClass
		StateClassValue  types.StateClass
	}

	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "known sensor type",
			fields: fields{DisplayName: "Active App"},
			want:   "active_app",
		},
		{
			name: "unset sensor type",
			want: "unknown_sensor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &Sensor{
				Value:            tt.fields.Value,
				IconString:       tt.fields.IconString,
				UnitsString:      tt.fields.UnitsString,
				DataSource:       tt.fields.SensorSrc,
				DisplayName:      tt.fields.DisplayName,
				IsBinary:         tt.fields.IsBinary,
				IsDiagnostic:     tt.fields.IsDiagnostic,
				DeviceClassValue: tt.fields.DeviceClassValue,
				StateClassValue:  tt.fields.StateClassValue,
			}
			if got := sensor.ID(); got != tt.want {
				t.Errorf("Sensor.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensor_State(t *testing.T) {
	type fields struct {
		Value            any
		IconString       string
		UnitsString      string
		SensorSrc        string
		IsBinary         bool
		IsDiagnostic     bool
		DeviceClassValue types.DeviceClass
		StateClassValue  types.StateClass
	}

	tests := []struct {
		want   any
		name   string
		fields fields
	}{
		{
			name:   "known value",
			fields: fields{Value: "someValue"},
			want:   "someValue",
		},
		{
			name: "unset value",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &Sensor{
				Value:            tt.fields.Value,
				IconString:       tt.fields.IconString,
				UnitsString:      tt.fields.UnitsString,
				DataSource:       tt.fields.SensorSrc,
				IsBinary:         tt.fields.IsBinary,
				IsDiagnostic:     tt.fields.IsDiagnostic,
				DeviceClassValue: tt.fields.DeviceClassValue,
				StateClassValue:  tt.fields.StateClassValue,
			}
			if got := sensor.State(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sensor.State() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensor_SensorType(t *testing.T) {
	type fields struct {
		Value            any
		IconString       string
		UnitsString      string
		SensorSrc        string
		IsBinary         bool
		IsDiagnostic     bool
		DeviceClassValue types.DeviceClass
		StateClassValue  types.StateClass
	}

	tests := []struct {
		name   string
		fields fields
		want   types.SensorClass
	}{
		{
			name: "default type",
			want: types.Sensor,
		},
		{
			name:   "binary type",
			fields: fields{IsBinary: true},
			want:   types.BinarySensor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &Sensor{
				Value:            tt.fields.Value,
				IconString:       tt.fields.IconString,
				UnitsString:      tt.fields.UnitsString,
				DataSource:       tt.fields.SensorSrc,
				IsBinary:         tt.fields.IsBinary,
				IsDiagnostic:     tt.fields.IsDiagnostic,
				DeviceClassValue: tt.fields.DeviceClassValue,
				StateClassValue:  tt.fields.StateClassValue,
			}
			if got := sensor.SensorType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sensor.SensorType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensor_Category(t *testing.T) {
	type fields struct {
		Value            any
		IconString       string
		UnitsString      string
		SensorSrc        string
		IsBinary         bool
		IsDiagnostic     bool
		DeviceClassValue types.DeviceClass
		StateClassValue  types.StateClass
	}

	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name: "default category",
			want: "",
		},
		{
			name:   "diagnostic category",
			fields: fields{IsDiagnostic: true},
			want:   "diagnostic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &Sensor{
				Value:            tt.fields.Value,
				IconString:       tt.fields.IconString,
				UnitsString:      tt.fields.UnitsString,
				DataSource:       tt.fields.SensorSrc,
				IsBinary:         tt.fields.IsBinary,
				IsDiagnostic:     tt.fields.IsDiagnostic,
				DeviceClassValue: tt.fields.DeviceClassValue,
				StateClassValue:  tt.fields.StateClassValue,
			}
			if got := sensor.Category(); got != tt.want {
				t.Errorf("Sensor.Category() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensor_Attributes(t *testing.T) {
	type fields struct {
		Value            any
		IconString       string
		UnitsString      string
		SensorSrc        string
		IsBinary         bool
		IsDiagnostic     bool
		DeviceClassValue types.DeviceClass
		StateClassValue  types.StateClass
	}

	tests := []struct {
		want   any
		name   string
		fields fields
	}{
		{
			name:   "with source",
			fields: fields{SensorSrc: DataSrcProcfs},
			want:   map[string]any{"data_source": DataSrcProcfs},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := &Sensor{
				Value:            tt.fields.Value,
				IconString:       tt.fields.IconString,
				UnitsString:      tt.fields.UnitsString,
				DataSource:       tt.fields.SensorSrc,
				IsBinary:         tt.fields.IsBinary,
				IsDiagnostic:     tt.fields.IsDiagnostic,
				DeviceClassValue: tt.fields.DeviceClassValue,
				StateClassValue:  tt.fields.StateClassValue,
			}
			if got := sensor.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sensor.Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}
