// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
//revive:disable:unused-receiver
package cpu

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"testing"

	"github.com/joshuar/go-hass-agent/internal/linux"
)

func generateRuns(b *testing.B) []int {
	b.Helper()

	var runs []int

	runs = append(runs, 1, 2)

	if runtime.NumCPU() > 4 {
		runs = append(runs, 4)
	}

	if runtime.NumCPU() > 8 {
		runs = append(runs, 8, runtime.NumCPU()/2)
	}

	runs = append(runs, runtime.NumCPU())

	return runs
}

func skipCI(t *testing.T) {
	t.Helper()

	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
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
	skipCI(t)

	validSensor := newCPUFreqSensor("cpu0")

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

func Benchmark_cpuFreqWorker_Sensors(b *testing.B) {
	runs := generateRuns(b)

	for _, v := range runs {
		var cpus []string
		for i := range runs {
			cpus = append(cpus, "cpu"+strconv.Itoa(i))
		}
		b.Run(fmt.Sprintf("num cpus %d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, cpu := range cpus {
					getCPUFreqs(cpu)
				}
			}
		})
	}
}
