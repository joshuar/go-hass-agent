// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
package cpu

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tklauser/go-sysconf"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

func Test_cpuUsageSensor_generateValues(t *testing.T) {
	skipCI(t)

	clktck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	require.NoError(t, err)

	validValues := make([]string, 0, 10)

	for range 10 {
		validValues = append(validValues, strconv.Itoa(rand.Intn(999999))) //nolint:gosec
	}

	type fields struct {
		cpuID           string
		usageAttributes map[string]any
		Sensor          linux.Sensor
	}
	type args struct {
		details []string
		clktk   int64
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   int
	}{
		{
			name:   "valid values",
			args:   args{clktk: clktck, details: validValues},
			fields: fields{cpuID: "cpu"},
			want:   len(validValues) + 1,
		},
		{
			name:   "invalid values",
			args:   args{clktk: clktck, details: make([]string, 0)},
			fields: fields{cpuID: "cpu"},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cpuUsageSensor{
				cpuID:           tt.fields.cpuID,
				usageAttributes: tt.fields.usageAttributes,
				Sensor:          tt.fields.Sensor,
			}
			s.generateValues(tt.args.clktk, tt.args.details)
			assert.Len(t, s.usageAttributes, tt.want)
		})
	}
}

func Test_cpuUsageSensor_Name(t *testing.T) {
	type fields struct {
		cpuID           string
		usageAttributes map[string]any
		Sensor          linux.Sensor
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "total",
			fields: fields{cpuID: "cpu"},
			want:   "Total CPU Usage",
		},
		{
			name:   "core",
			fields: fields{cpuID: "cpu2"},
			want:   "Core 2 CPU Usage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cpuUsageSensor{
				cpuID:           tt.fields.cpuID,
				usageAttributes: tt.fields.usageAttributes,
				Sensor:          tt.fields.Sensor,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("cpuUsageSensor.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cpuUsageSensor_ID(t *testing.T) {
	type fields struct {
		cpuID           string
		usageAttributes map[string]any
		Sensor          linux.Sensor
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "total",
			fields: fields{cpuID: "cpu"},
			want:   "total_cpu_usage",
		},
		{
			name:   "core",
			fields: fields{cpuID: "cpu2"},
			want:   "core_2_cpu_usage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &cpuUsageSensor{
				cpuID:           tt.fields.cpuID,
				usageAttributes: tt.fields.usageAttributes,
				Sensor:          tt.fields.Sensor,
			}
			if got := s.ID(); got != tt.want {
				t.Errorf("cpuUsageSensor.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUsageWorker(t *testing.T) {
	skipCI(t)

	clktck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	require.NoError(t, err)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	defer cancelFunc()
	ctx = logging.ToContext(ctx, slog.Default())

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
			name: "valid",
			args: args{in0: ctx},
			want: &linux.SensorWorker{
				Value: &usageWorker{
					clktck: clktck,
				},
				WorkerID: usageWorkerID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUsageWorker(tt.args.in0, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUsageWorker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewUsageWorker() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_usageWorker_newUsageSensor(t *testing.T) {
	skipCI(t)

	clktck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	require.NoError(t, err)

	validValues := []string{"cpu", "100", "0", "0", "0", "0", "0", "0", "0", "0"}
	validSensor := &cpuUsageSensor{
		cpuID: "cpu",
		Sensor: linux.Sensor{
			IconString:      "mdi:chip",
			UnitsString:     "%",
			SensorSrc:       linux.DataSrcProcfs,
			StateClassValue: types.StateClassMeasurement,
			SensorTypeValue: linux.SensorCPUPc,
			IsDiagnostic:    false,
		},
	}
	validSensor.generateValues(clktck, validValues[1:])

	type fields struct {
		clktck int64
	}
	type args struct {
		details    []string
		diagnostic bool
	}
	tests := []struct {
		want   *cpuUsageSensor
		name   string
		args   args
		fields fields
	}{
		{
			name:   "valid values",
			args:   args{details: validValues, diagnostic: false},
			fields: fields{clktck: clktck},
			want:   validSensor,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &usageWorker{
				clktck: tt.fields.clktck,
			}
			if got := w.newUsageSensor(tt.args.details, tt.args.diagnostic); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("usageWorker.newUsageSensor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_usageWorker_newCountSensor(t *testing.T) {
	type fields struct {
		logger *slog.Logger
		clktck int64
	}
	type args struct {
		icon       string
		details    string
		sensorType linux.SensorTypeValue
	}
	tests := []struct {
		fields fields
		want   *linux.Sensor
		name   string
		args   args
	}{
		{
			name:   "valid values",
			args:   args{sensorType: linux.SensorCPUCtxSwitch, icon: "mdi:counter", details: "400"},
			fields: fields{logger: slog.Default()},
			want: &linux.Sensor{
				Value:           400,
				IconString:      "mdi:counter",
				SensorSrc:       linux.DataSrcProcfs,
				StateClassValue: types.StateClassTotalIncreasing,
				SensorTypeValue: linux.SensorCPUCtxSwitch,
				IsDiagnostic:    true,
				LastReset:       "0001-01-01T00:00:00Z",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &usageWorker{
				clktck: tt.fields.clktck,
				logger: tt.fields.logger,
			}
			if got := w.newCountSensor(tt.args.sensorType, tt.args.icon, tt.args.details); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("usageWorker.newProcCntSensor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_usageWorker_getStats(t *testing.T) {
	skipCI(t)

	clktck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	require.NoError(t, err)

	type fields struct {
		logger *slog.Logger
		clktck int64
	}
	tests := []struct {
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name:   "valid",
			fields: fields{clktck: clktck},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &usageWorker{
				clktck: tt.fields.clktck,
				logger: tt.fields.logger,
			}
			got, err := w.getStats()
			if (err != nil) != tt.wantErr {
				t.Errorf("getStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Require a total cpu usage sensor.
				require.True(t, slices.ContainsFunc(got, func(d sensor.Details) bool {
					return d.Name() == "Total CPU Usage"
				}))
				// Require at least 1 cpu core usage sensor.
				require.True(t, slices.ContainsFunc(got, func(d sensor.Details) bool {
					return d.Name() == "Core 1 CPU Usage"
				}))
				// Require a context switches sensor
				require.True(t, slices.ContainsFunc(got, func(d sensor.Details) bool {
					return d.Name() == "Total CPU Context Switches"
				}))
				// Require a processes total sensor
				require.True(t, slices.ContainsFunc(got, func(d sensor.Details) bool {
					return d.Name() == "Total Processes Created"
				}))
				// Require a procs running sensor
				require.True(t, slices.ContainsFunc(got, func(d sensor.Details) bool {
					return d.Name() == "Processes Running"
				}))
				// Require a processes blocked sensor
				require.True(t, slices.ContainsFunc(got, func(d sensor.Details) bool {
					return d.Name() == "Processes Blocked"
				}))
			}
		})
	}
}

func Benchmark_getStats(b *testing.B) {
	clktck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	require.NoError(b, err)

	w := &usageWorker{clktck: clktck}

	b.Run(fmt.Sprintf("run %d", b.N), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w.getStats() //nolint:errcheck
		}
	})
}
