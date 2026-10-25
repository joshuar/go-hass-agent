package mem

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
)

func TestUsageWorkerExecute(t *testing.T) {
	worker := &usageWorker{
		WorkerMetadata: models.SetWorkerMetadata(usageWorkerID, usageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{
			OutCh: make(chan models.Entity, 8),
		},
	}

	require.NoError(t, worker.Execute(context.Background()))
	close(worker.OutCh)

	sensors := make(map[string]models.Sensor)
	for entity := range worker.OutCh {
		sensor, err := entity.AsSensor()
		require.NoError(t, err)
		sensors[sensor.UniqueID] = sensor
	}

	require.Len(t, sensors, 4)

	total, ok := sensors["memory_total"].State.(float64)
	require.True(t, ok)
	assert.Positive(t, total)

	used, ok := sensors["memory_used"].State.(float64)
	require.True(t, ok)
	assert.Positive(t, used)
	assert.LessOrEqual(t, used, total)

	usedPc, ok := sensors["memory_used_pc"].State.(float64)
	require.True(t, ok)
	assert.GreaterOrEqual(t, usedPc, 0.0)
	assert.LessOrEqual(t, usedPc, 100.0)
}
