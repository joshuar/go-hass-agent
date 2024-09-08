// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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
		id:          "temp1_crit",
		MonitorType: Alarm,
		Attributes: []Attribute{
			{Name: "max", Value: 80},
			{Name: "crit", Value: 100},
		},
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
				Path:     "testing/data/hwmon0",
				chipName: "coretemp",
				chipID:   "hwmon0",
			},
			want: []*Sensor{tempSensor, alarmSensor},
		},
		{
			name: "fail",
			fields: fields{
				Path: "testing/data/hwmon10",
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
			got, err := c.getSensors()
			if (err != nil) != tt.wantErr {
				t.Errorf("Chip.getSensors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for i := range tt.want {
					if !slices.ContainsFunc(got, func(s *Sensor) bool {
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
			args: args{path: "testing/data/hwmon0"},
			want: &Chip{
				chipName: "coretemp",
				chipID:   "hwmon0",
				Path:     "testing/data/hwmon0",
			},
		},
		{
			name: "with a device model",
			args: args{path: "testing/data/hwmon1"},
			want: &Chip{
				chipName:    "drivetemp",
				chipID:      "hwmon1",
				Path:        "testing/data/hwmon1",
				deviceModel: "CT1000MX500SSD1",
			},
		},
		{
			name:    "fail",
			args:    args{path: "testing/data/hwmon10"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newChip(tt.args.path)
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

	chip0, err := newChip("testing/data/hwmon0")
	require.NoError(t, err)
	chip1, err := newChip("testing/data/hwmon1")
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
			args: args{path: "testing/data"},
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
			got, err := GetAllChips()
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
					Path:        "testing/data/hwmon1",
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
					Path:        "testing/data/hwmon1",
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
					Path:     "testing/data/hwmon0",
					chipName: "coretemp",
					chipID:   "hwmon0",
				},
				id: "Temp1",
			},
			want: "Hardware Sensor Coretemp Temp1",
		},
		{
			name: "no device model, label",
			fields: fields{
				Chip: &Chip{
					Path:     "testing/data/hwmon0",
					chipName: "coretemp",
					chipID:   "hwmon0",
				},
				id:    "Temp1",
				label: "Core 1",
			},
			want: "Hardware Sensor Coretemp Core 1",
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
					Path:     "testing/data/hwmon0",
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
			args: args{path: "testing/data"},
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
			got, err := GetAllSensors()
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

func Test_sensorFile_getSensorType(t *testing.T) {
	type fields struct {
		path       string
		filename   string
		sensorType string
		sensorAttr string
	}
	tests := []struct {
		name            string
		fields          fields
		wantSensorType  MonitorType
		wantScaleFactor float64
		wantUnits       string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &sensorFile{
				path:       tt.fields.path,
				filename:   tt.fields.filename,
				sensorType: tt.fields.sensorType,
				sensorAttr: tt.fields.sensorAttr,
			}
			gotSensorType, gotScaleFactor, gotUnits := f.getSensorType()
			if gotSensorType != tt.wantSensorType {
				t.Errorf("sensorFile.getSensorType() gotSensorType = %v, want %v", gotSensorType, tt.wantSensorType)
			}
			if gotScaleFactor != tt.wantScaleFactor {
				t.Errorf("sensorFile.getSensorType() gotScaleFactor = %v, want %v", gotScaleFactor, tt.wantScaleFactor)
			}
			if gotUnits != tt.wantUnits {
				t.Errorf("sensorFile.getSensorType() gotUnits = %v, want %v", gotUnits, tt.wantUnits)
			}
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

func Test_getValueAsString(t *testing.T) {
	type args struct {
		p string
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
			got, err := getValueAsString(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValueAsString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getValueAsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getValueAsFloat(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValueAsFloat(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValueAsFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getValueAsFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getValueAsBool(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValueAsBool(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValueAsBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getValueAsBool() = %v, want %v", got, tt.want)
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
		for i := 0; i < b.N; i++ {
			GetAllSensors() //nolint:errcheck
		}
	})
}
