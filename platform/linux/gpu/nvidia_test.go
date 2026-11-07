// Copyright 2026 opticon454
// SPDX-License-Identifier: MIT

package gpu

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
)

func Test_parseNvidiaSMI(t *testing.T) {
	file, err := os.Open("testdata/nvidia-smi.csv")
	require.NoError(t, err)

	defer file.Close()

	gpus, err := parseNvidiaSMI(file)
	require.NoError(t, err)
	require.Len(t, gpus, 2)

	assert.Equal(t, 0, gpus[0].index)
	assert.Equal(t, "NVIDIA GeForce RTX 3060", gpus[0].name)
	assert.InDelta(t, 19.26, gpus[0].stats["power"], 0.001)
	assert.InDelta(t, 12288, gpus[0].stats["memory_total"], 0.001)
	assert.Len(t, gpus[0].stats, len(nvidiaMetrics))

	// The second card reports [N/A] for its fan, so that sensor is left out.
	assert.Equal(t, 1, gpus[1].index)
	assert.NotContains(t, gpus[1].stats, "fan_speed")
	assert.Len(t, gpus[1].stats, len(nvidiaMetrics)-1)
}

func Test_parseNvidiaSMI_errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "too few columns", input: "0, GPU-1, Card, 10\n"},
		{name: "bad index", input: "x, GPU-1, Card, 1, 2, 3, 4, 5, 6, 7\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseNvidiaSMI(strings.NewReader(tt.input))
			require.ErrorIs(t, err, ErrGetNvidiaStats)
		})
	}

	gpus, err := parseNvidiaSMI(strings.NewReader(""))
	require.NoError(t, err)
	assert.Empty(t, gpus)
}

func Test_newNvidiaSensors(t *testing.T) {
	gpu := nvidiaGPU{index: 1, uuid: "GPU-x", name: "Card", stats: map[string]float64{"power": 65.25, "fan_speed": 40}}

	entities := newNvidiaSensors(context.Background(), gpu)
	require.Len(t, entities, 2)

	sensors := make(map[string]models.Sensor)

	for _, entity := range entities {
		sensor, err := entity.AsSensor()
		require.NoError(t, err)

		sensors[string(sensor.UniqueID)] = sensor
	}

	power, found := sensors["nvidia_gpu_1_power"]
	require.True(t, found)
	assert.Equal(t, "Nvidia GPU 1 Power", string(power.Name))
	assert.Equal(t, "power", power.DeviceClass)
	assert.Equal(t, "W", string(power.UnitOfMeasurement))
	assert.Equal(t, "measurement", power.StateClass)

	// A percentage has no device class.
	fan, found := sensors["nvidia_gpu_1_fan_speed"]
	require.True(t, found)
	assert.Empty(t, fan.DeviceClass)
	assert.Equal(t, "%", string(fan.UnitOfMeasurement))
}

func newTestNvidiaWorker(query func(ctx context.Context, fields string) ([]byte, error)) *nvidiaWorker {
	return &nvidiaWorker{
		WorkerMetadata:          models.SetWorkerMetadata("nvidia_gpu", "Nvidia GPU metrics"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{OutCh: make(chan models.Entity, 32)},
		prefs:                   &NvidiaPrefs{},
		path:                    "/usr/bin/nvidia-smi",
		query:                   query,
	}
}

func Test_nvidiaWorker_Execute(t *testing.T) {
	data, err := os.ReadFile("testdata/nvidia-smi.csv")
	require.NoError(t, err)

	t.Run("sends a sensor per reported value", func(t *testing.T) {
		worker := newTestNvidiaWorker(func(_ context.Context, fields string) ([]byte, error) {
			assert.Equal(t, nvidiaQueryFields, fields)
			return data, nil
		})
		require.NoError(t, worker.Execute(context.Background()))
		assert.Len(t, worker.OutCh, 2*len(nvidiaMetrics)-1)
	})

	t.Run("falls back to the legacy query", func(t *testing.T) {
		var tried []string

		worker := newTestNvidiaWorker(func(_ context.Context, fields string) ([]byte, error) {
			tried = append(tried, fields)
			if fields == nvidiaQueryFields {
				return nil, errors.New("unknown field")
			}
			return data, nil
		})
		require.NoError(t, worker.Execute(context.Background()))
		assert.Equal(t, []string{nvidiaQueryFields, nvidiaQueryFieldsLegacy}, tried)
		assert.NotEmpty(t, worker.OutCh)
	})

	t.Run("returns an error when both queries fail", func(t *testing.T) {
		worker := newTestNvidiaWorker(func(_ context.Context, _ string) ([]byte, error) { return nil, errors.New("no driver") })
		require.Error(t, worker.Execute(context.Background()))
		assert.Empty(t, worker.OutCh)
	})
}

func Test_nvidiaWorker_IsDisabled(t *testing.T) {
	worker := newTestNvidiaWorker(nil)
	assert.False(t, worker.IsDisabled())

	worker.path = ""
	assert.True(t, worker.IsDisabled(), "disabled when nvidia-smi is not installed")

	worker.path = "/usr/bin/nvidia-smi"
	worker.prefs.Disabled = true
	assert.True(t, worker.IsDisabled(), "disabled in the preferences")
}
