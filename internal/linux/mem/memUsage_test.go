// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mem

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

func Test_newMemSensor(t *testing.T) {
	type args struct {
		stat *memStat
		file string
		id   memStatID
	}
	tests := []struct {
		want *linux.Sensor
		name string
		args args
	}{
		{
			name: "valid sensor",
			args: args{
				id:   memTotal,
				stat: &memStat{value: 32572792 * 1000, units: "B"},
				file: "testing/data/meminfowithswap",
			},
			want: &linux.Sensor{Value: uint64(32572792 * 1000), DisplayName: memTotal.String()},
		},
		{
			name: "missing sensor",
			args: args{
				id:   swapTotal,
				stat: &memStat{value: 0, units: "B"},
				file: "testing/data/meminfowithoutswap",
			},
			want: &linux.Sensor{Value: uint64(0), DisplayName: swapTotal.String()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStatFile = tt.args.file
			got := newMemSensor(tt.args.id, tt.args.stat)
			assert.Equal(t, tt.want.DisplayName, got.DisplayName)
			assert.Equal(t, tt.want.Value, got.Value)
		})
	}
}

func Test_newMemUsedPc(t *testing.T) {
	memStatFile = "testing/data/meminfowithswap"
	withSwap, err := getMemStats()
	require.NoError(t, err)
	memUsed := withSwap[memTotal].value - withSwap[memFree].value - withSwap[memBuffered].value - withSwap[memCached].value
	memUsedPc := math.Round(float64(memUsed)/float64(withSwap[memTotal].value)*100/0.05) * 0.05

	type args struct {
		stats memoryStats
	}
	tests := []struct {
		args args
		want *linux.Sensor
		name string
	}{
		{
			name: "valid sensor",
			args: args{stats: withSwap},
			want: &linux.Sensor{
				Value: memUsedPc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newMemUsedPc(tt.args.stats)
			assert.Equal(t, got.Value, tt.want.Value)
		})
	}
}

func Test_newSwapUsedPc(t *testing.T) {
	memStatFile = "testing/data/meminfowithswap"
	withSwap, err := getMemStats()
	require.NoError(t, err)
	swapUsed := withSwap[swapTotal].value - withSwap[swapFree].value
	swapUsedPc := math.Round(float64(swapUsed)/float64(withSwap[swapTotal].value)*100/0.05) * 0.05
	memStatFile = "testing/data/meminfowithoutswap"
	withoutSwap, err := getMemStats()
	require.NoError(t, err)

	type args struct {
		stats memoryStats
	}
	tests := []struct {
		args args
		want *linux.Sensor
		name string
	}{
		{
			name: "with swap",
			args: args{stats: withSwap},
			want: &linux.Sensor{
				Value: swapUsedPc,
			},
		},
		{
			name: "without swap",
			args: args{stats: withoutSwap},
			want: &linux.Sensor{
				Value: float64(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newSwapUsedPc(tt.args.stats)
			assert.Equal(t, got.Value, tt.want.Value)
		})
	}
}

func Test_usageWorker_Sensors(t *testing.T) {
	memStatFile = "testing/data/meminfowithswap"
	withSwap, err := getMemStats()
	require.NoError(t, err)
	withSwapSensors := make([]sensor.Details, 0, len(memSensors)+len(swapSensors)+2)
	for _, id := range memSensors {
		withSwapSensors = append(withSwapSensors, newMemSensor(id, withSwap[id]))
	}
	withSwapSensors = append(withSwapSensors, newMemUsedPc(withSwap))
	for _, id := range swapSensors {
		withSwapSensors = append(withSwapSensors, newMemSensor(id, withSwap[id]))
	}
	withSwapSensors = append(withSwapSensors, newSwapUsedPc(withSwap))

	memStatFile = "testing/data/meminfowithoutswap"
	withoutSwap, err := getMemStats()
	require.NoError(t, err)
	withoutSwapSensors := make([]sensor.Details, 0, len(memSensors)+len(swapSensors)+2)
	for _, id := range memSensors {
		withoutSwapSensors = append(withoutSwapSensors, newMemSensor(id, withoutSwap[id]))
	}
	withoutSwapSensors = append(withoutSwapSensors, newMemUsedPc(withoutSwap))

	type args struct {
		in0  context.Context
		file string
		in1  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    []sensor.Details
		wantErr bool
	}{
		{
			name: "with swap",
			args: args{in0: linux.NewContext(context.TODO()), file: "testing/data/meminfowithswap"},
			want: withSwapSensors,
		},
		{
			name: "without swap",
			args: args{in0: linux.NewContext(context.TODO()), file: "testing/data/meminfowithoutswap"},
			want: withoutSwapSensors,
		},
		{
			name:    "no stats file",
			args:    args{in0: linux.NewContext(context.TODO()), file: "/nonexistent"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &usageWorker{}
			memStatFile = tt.args.file
			got, err := w.Sensors(tt.args.in0, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("usageWorker.Sensors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i := range tt.want {
				assert.Equal(t, tt.want[i].State(), got[i].State())
				assert.Equal(t, tt.want[i].ID(), got[i].ID())
				assert.Equal(t, tt.want[i].Name(), got[i].Name())
			}
		})
	}
}
