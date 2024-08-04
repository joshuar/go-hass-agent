// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest,wsl,containedctx,nlreturn,dupl
//revive:disable:unused-receiver
package cpu

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

func skipCI(t *testing.T) {
	t.Helper()

	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}

func Test_getCPUFreqs(t *testing.T) {
	skipCI(t)
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid path",
			args: args{path: filepath.Join(sysFSPath, "cpu*", freqFile)},
		},
		{
			name:    "invalid path",
			args:    args{path: "/nonexistent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCPUFreqs(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCPUFreqs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Greater(t, len(got), 1)
			}
		})
	}
}

func TestNewCPUFreqWorker(t *testing.T) {
	type args struct {
		in0 context.Context
		in1 *dbusx.DBusAPI
	}
	tests := []struct {
		args    args
		want    *linux.SensorWorker
		name    string
		wantErr bool
	}{
		{
			name: "valid worker",
			want: &linux.SensorWorker{
				Value: &cpuFreqWorker{
					path: filepath.Join(sysFSPath, "cpu*", freqFile),
				},
				WorkerID: cpuFreqWorkerID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCPUFreqWorker(tt.args.in0, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCPUFreqWorker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCPUFreqWorker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cpuFreqWorker_Interval(t *testing.T) {
	type fields struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "valid interval",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cpuFreqWorker{
				path: tt.fields.path,
			}
			got := w.Interval()
			assert.Greater(t, got, cpuFreqUpdateJitter)
		})
	}
}

func Test_cpuFreqWorker_Jitter(t *testing.T) {
	type fields struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "valid jitter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cpuFreqWorker{
				path: tt.fields.path,
			}
			got := w.Jitter()
			assert.Less(t, got, cpuFreqUpdateInterval)
		})
	}
}

func Test_cpuFreqWorker_Sensors(t *testing.T) {
	skipCI(t)
	type fields struct {
		path string
	}
	type args struct {
		in0 context.Context
		in1 time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []sensor.Details
		wantErr bool
	}{
		{
			name:   "valid path",
			fields: fields{path: filepath.Join(sysFSPath, "cpu*", freqFile)},
		},
		{
			name:    "invalid path",
			fields:  fields{path: "/nonexistent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &cpuFreqWorker{
				path: tt.fields.path,
			}
			got, err := w.Sensors(tt.args.in0, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("cpuFreqWorker.Sensors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Greater(t, len(got), 1)
			}
		})
	}
}

func Test_cpuFreqSensor_Name(t *testing.T) {
	type fields struct {
		cpuFreq *cpuFreq
		Sensor  linux.Sensor
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "valid name",
			fields: fields{cpuFreq: &cpuFreq{cpu: "cpu12"}},
			want:   "Core 12 Frequency",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cpuFreqSensor{
				cpuFreq: tt.fields.cpuFreq,
				Sensor:  tt.fields.Sensor,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("cpuFreqSensor.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cpuFreqSensor_ID(t *testing.T) {
	type fields struct {
		cpuFreq *cpuFreq
		Sensor  linux.Sensor
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "valid id",
			fields: fields{cpuFreq: &cpuFreq{cpu: "cpu12"}},
			want:   "cpufreq_core12_frequency",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cpuFreqSensor{
				cpuFreq: tt.fields.cpuFreq,
				Sensor:  tt.fields.Sensor,
			}
			if got := s.ID(); got != tt.want {
				t.Errorf("cpuFreqSensor.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cpuFreqSensor_Attributes(t *testing.T) {
	validSensor := newCPUFreqSensor(
		cpuFreq{
			cpu:      "cpu12",
			governor: "performance",
			driver:   "intel_pstate",
			freq:     "999999",
		},
	)

	type fields struct {
		cpuFreq *cpuFreq
		Sensor  linux.Sensor
	}
	tests := []struct {
		want   map[string]any
		name   string
		fields fields
	}{
		{
			name:   "valid sensor",
			fields: fields{cpuFreq: validSensor.cpuFreq, Sensor: validSensor.Sensor},
			want: map[string]any{
				"governor":                   validSensor.governor,
				"driver":                     validSensor.driver,
				"native_unit_of_measurement": cpuFreqUnits,
				"data_source":                linux.DataSrcSysfs,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cpuFreqSensor{
				cpuFreq: tt.fields.cpuFreq,
				Sensor:  tt.fields.Sensor,
			}
			if got := s.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cpuFreqSensor.Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}
