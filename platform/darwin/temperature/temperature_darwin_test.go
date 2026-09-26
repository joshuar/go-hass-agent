package temperature

import (
	"context"
	"errors"
	"math"
	"os"
	"testing"

	"github.com/shirou/gopsutil/v4/sensors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
)

func newTestWorker(read func(context.Context) ([]sensors.TemperatureStat, error)) *temperatureWorker {
	return &temperatureWorker{
		WorkerMetadata: models.SetWorkerMetadata(temperatureWorkerID, temperatureWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{
			OutCh: make(chan models.Entity, 128),
		},
		read: read,
	}
}

func drain(t *testing.T, w *temperatureWorker) []models.Sensor {
	t.Helper()
	close(w.OutCh)

	var out []models.Sensor

	for entity := range w.OutCh {
		sensor, err := entity.AsSensor()
		require.NoError(t, err)

		out = append(out, sensor)
	}

	return out
}

func TestExecuteFiltersDedupesAndSorts(t *testing.T) {
	worker := newTestWorker(func(context.Context) ([]sensors.TemperatureStat, error) {
		return []sensors.TemperatureStat{
			{SensorKey: "PMU tdie4", Temperature: 43.74},
			{SensorKey: "NAND CH0 temp", Temperature: 38},
			{SensorKey: "PMU tdev4", Temperature: -21.5}, // placeholder
			{SensorKey: "TA0P", Temperature: 0},          // unreadable Intel SMC key
			{SensorKey: "TC0D", Temperature: math.NaN()},
			{SensorKey: "hot", Temperature: 150},
			{SensorKey: "PMU tdie4", Temperature: 50}, // duplicate key
		}, nil
	})

	require.NoError(t, worker.Execute(context.Background()))

	sensorList := drain(t, worker)
	require.Len(t, sensorList, 2)

	assert.Equal(t, "temperature_nand_ch0_temp", sensorList[0].UniqueID)
	assert.InDelta(t, 38.0, sensorList[0].State, 0.001)
	assert.Equal(t, "temperature_pmu_tdie4", sensorList[1].UniqueID)
	assert.InDelta(t, 43.7, sensorList[1].State, 0.001, "rounded to one decimal, first duplicate wins")
}

func TestExecutePartialResultsWithError(t *testing.T) {
	worker := newTestWorker(func(context.Context) ([]sensors.TemperatureStat, error) {
		return []sensors.TemperatureStat{{SensorKey: "a", Temperature: 30}}, errors.New("warning")
	})

	require.NoError(t, worker.Execute(context.Background()))
	assert.Len(t, drain(t, worker), 1)
}

func TestExecuteErrorWithNoResults(t *testing.T) {
	worker := newTestWorker(func(context.Context) ([]sensors.TemperatureStat, error) {
		return nil, errors.New("no sensors")
	})

	require.Error(t, worker.Execute(context.Background()))
}

func TestExecuteReal(t *testing.T) {
	worker := newTestWorker(sensors.TemperaturesWithContext)

	err := worker.Execute(context.Background())
	require.NoError(t, err)

	sensorList := drain(t, worker)

	// Virtualized CI runners don't expose hardware temperature sensors.
	if os.Getenv("CI") != "" && len(sensorList) == 0 {
		t.Skip("no temperature sensors available in this CI environment")
	}

	require.NotEmpty(t, sensorList, "expected at least one temperature sensor")

	for _, s := range sensorList {
		assert.Contains(t, s.UniqueID, "temperature_")
	}
}
