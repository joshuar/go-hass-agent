package cpu

import (
	"context"
	"testing"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
)

func TestUsageWorkerExecute(t *testing.T) {
	worker := &usageWorker{
		WorkerMetadata: models.SetWorkerMetadata(usageWorkerID, usageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{
			OutCh: make(chan models.Entity, 64),
		},
	}

	require.NoError(t, worker.Execute(context.Background()))
	close(worker.OutCh)

	var sensors []models.Sensor
	for entity := range worker.OutCh {
		sensor, err := entity.AsSensor()
		require.NoError(t, err)
		sensors = append(sensors, sensor)
	}

	require.NotEmpty(t, sensors, "expected at least one core plus the total sensor")

	last := sensors[len(sensors)-1]
	assert.Equal(t, "total_cpu_usage", last.UniqueID)

	pct, ok := last.State.(float64)
	require.True(t, ok, "state should be a float64 percentage")
	assert.GreaterOrEqual(t, pct, 0.0)
	assert.LessOrEqual(t, pct, 100.0)

	// One sensor per core plus the total.
	cores, err := cpu.CountsWithContext(context.Background(), true)
	require.NoError(t, err)
	assert.Len(t, sensors, cores+1)
}
