// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package hwmon

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChip_String(t *testing.T) {
	type fields struct {
		chipName    string
		chipID      string
		deviceModel string
		HWMonPath   string
		Sensors     []*Sensor
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name: "with device model",
			fields: fields{
				deviceModel: "PM9A1 NVMe Samsung 1024GB",
				chipID:      "hwmon1",
				chipName:    "nvme",
			},
			want: "PM9A1 NVMe Samsung 1024GB",
		},
		{
			name: "with name",
			fields: fields{
				chipID:   "hwmon1",
				chipName: "nvme",
			},
			want: "nvme",
		},
		{
			name: "without model or name",
			fields: fields{
				chipID: "hwmon1",
			},
			want: "hwmon1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Chip{
				chipName:    tt.fields.chipName,
				chipID:      tt.fields.chipID,
				deviceModel: tt.fields.deviceModel,
				Path:        tt.fields.HWMonPath,
				Sensors:     tt.fields.Sensors,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("Chip.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChip_getSensors(t *testing.T) {
	tempSensor := &Sensor{
		value:       36,
		label:       "Package id 0",
		id:          "temp1",
		units:       "°C",
		scaleFactor: 1000,
		MonitorType: Temp,
		Attributes: []Attribute{
			{Name: "max", Value: 80},
			{Name: "crit", Value: 100},
		},
	}
	alarmSensor := &Sensor{
		value:       false,
		label:       "Temp1 Crit Alarm",
		id:          "temp1_crit_alarm",
		MonitorType: Alarm,
	}

	type fields struct {
		chipName    string
		chipID      string
		deviceModel string
		Path        string
		Sensors     []*Sensor
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*Sensor
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				Path:     "testdata/hwmon0",
				chipName: "coretemp",
				chipID:   "hwmon0",
			},
			want: []*Sensor{tempSensor, alarmSensor},
		},
		{
			name: "fail",
			fields: fields{
				Path: "testdata/hwmon10",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Chip{
				chipName:    tt.fields.chipName,
				chipID:      tt.fields.chipID,
				deviceModel: tt.fields.deviceModel,
				Path:        tt.fields.Path,
				Sensors:     tt.fields.Sensors,
			}
			for i := range tt.want {
				tt.want[i].Chip = c
			}
			got, err := c.getSensors(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("Chip.getSensors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for i := range tt.want {
					if !slices.ContainsFunc(got, func(s *Sensor) bool {
						t.Logf("Chip.getSensors() = %v, want %v", s.id, tt.want[i].id)
						return s.id == tt.want[i].id
					}) {
						t.Errorf("Chip.getSensors() = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}

func Test_newChip(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		want    *Chip
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "without device model",
			args: args{path: "testdata/hwmon0"},
			want: &Chip{
				chipName: "coretemp",
				chipID:   "hwmon0",
				Path:     "testdata/hwmon0",
			},
		},
		{
			name: "with a device model",
			args: args{path: "testdata/hwmon1"},
			want: &Chip{
				chipName:    "drivetemp",
				chipID:      "hwmon1",
				Path:        "testdata/hwmon1",
				deviceModel: "CT1000MX500SSD1",
			},
		},
		{
			name:    "fail",
			args:    args{path: "testdata/hwmon10"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newChip(t.Context(), tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("newChip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want.chipID, got.chipID)
				assert.Equal(t, tt.want.chipName, got.chipName)
				assert.Equal(t, tt.want.Path, got.Path)
				assert.Equal(t, tt.want.deviceModel, got.deviceModel)
			}
		})
	}
}

func TestGetAllChips(t *testing.T) {
	var chips []*Chip

	chip0, err := newChip(t.Context(), "testdata/hwmon0")
	require.NoError(t, err)
	chip1, err := newChip(t.Context(), "testdata/hwmon1")
	require.NoError(t, err)

	chips = append(chips, chip0, chip1)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Chip
		wantErr bool
	}{
		{
			name: "success",
			args: args{path: "testdata"},
			want: chips,
		},
		{
			name:    "fail",
			args:    args{path: "/nonexistent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HWMonPath = tt.args.path
			got, err := GetAllChips(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllChips() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Len(t, got, 2)
			}
		})
	}
}

func TestSensor_Name(t *testing.T) {
	type fields struct {
		Chip        *Chip
		value       any
		label       string
		id          string
		units       string
		Attributes  []Attribute
		scaleFactor float64
		MonitorType MonitorType
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name: "device model, no label",
			fields: fields{
				Chip: &Chip{
					deviceModel: "CT1000MX500SSD1",
					Path:        "testdata/hwmon1",
					chipName:    "drivetemp",
					chipID:      "hwmon1",
				},
				id: "Temp1",
			},
			want: "CT1000MX500SSD1 Temp1",
		},
		{
			name: "device model, label",
			fields: fields{
				Chip: &Chip{
					deviceModel: "CT1000MX500SSD1",
					Path:        "testdata/hwmon1",
					chipName:    "drivetemp",
					chipID:      "hwmon1",
				},
				id:    "Temp1",
				label: "Drive Temp 1",
			},
			want: "CT1000MX500SSD1 Drive Temp 1",
		},
		{
			name: "no device model, no label",
			fields: fields{
				Chip: &Chip{
					Path:     "testdata/hwmon0",
					chipName: "coretemp",
					chipID:   "hwmon0",
				},
				id: "Temp1",
			},
			want: "Hardware Sensor coretemp Temp1",
		},
		{
			name: "no device model, label",
			fields: fields{
				Chip: &Chip{
					Path:     "testdata/hwmon0",
					chipName: "coretemp",
					chipID:   "hwmon0",
				},
				id:    "Temp1",
				label: "Core 1",
			},
			want: "Hardware Sensor coretemp Core 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sensor{
				Chip:        tt.fields.Chip,
				value:       tt.fields.value,
				label:       tt.fields.label,
				id:          tt.fields.id,
				units:       tt.fields.units,
				Attributes:  tt.fields.Attributes,
				scaleFactor: tt.fields.scaleFactor,
				MonitorType: tt.fields.MonitorType,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("Sensor.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensor_ID(t *testing.T) {
	type fields struct {
		Chip        *Chip
		value       any
		label       string
		id          string
		units       string
		Attributes  []Attribute
		scaleFactor float64
		MonitorType MonitorType
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name: "valid",
			fields: fields{
				Chip: &Chip{
					Path:     "testdata/hwmon0",
					chipName: "coretemp",
					chipID:   "hwmon0",
				},
				id: "Temp1",
			},
			want: "hwmon_0_coretemp_temp_1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sensor{
				Chip:        tt.fields.Chip,
				value:       tt.fields.value,
				label:       tt.fields.label,
				id:          tt.fields.id,
				units:       tt.fields.units,
				Attributes:  tt.fields.Attributes,
				scaleFactor: tt.fields.scaleFactor,
				MonitorType: tt.fields.MonitorType,
			}
			if got := s.ID(); got != tt.want {
				t.Errorf("Sensor.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAllSensors(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Sensor
		wantErr bool
	}{
		{
			name: "success",
			args: args{path: "testdata"},
		},
		{
			name:    "fail",
			args:    args{path: "/nonexistent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HWMonPath = tt.args.path
			got, err := GetAllSensors(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllSensors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Len(t, got, 3)
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("GetAllSensors() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_getFileContents(t *testing.T) {
	type args struct {
		path string
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
			got, err := getFileContents(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFileContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getFileContents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSensorFiles(t *testing.T) {
	type args struct {
		hwMonPath string
	}
	tests := []struct {
		name    string
		args    args
		want    []sensorFile
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSensorFiles(tt.args.hwMonPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSensorFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSensorFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Benchmark_GetAllSensors(b *testing.B) {
	b.Run(fmt.Sprintf("run %d", b.N), func(b *testing.B) {
		for b.Loop() {
			_, err := GetAllSensors(b.Context())
			if err != nil {
				b.Log("problem getting sensors: %w", err)
			}
		}
	})
}
