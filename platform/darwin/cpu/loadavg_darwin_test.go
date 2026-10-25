package cpu

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
)

func TestLoadAvgWorkerExecute(t *testing.T) {
	worker := &loadAvgWorker{
		WorkerMetadata: models.SetWorkerMetadata(loadAvgWorkerID, loadAvgWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{
			OutCh: make(chan models.Entity, 8),
		},
	}

	require.NoError(t, worker.Execute(context.Background()))
	close(worker.OutCh)

	var ids []string
	for entity := range worker.OutCh {
		sensor, err := entity.AsSensor()
		require.NoError(t, err)
		ids = append(ids, sensor.UniqueID)

		value, ok := sensor.State.(float64)
		require.True(t, ok)
		assert.GreaterOrEqual(t, value, 0.0)
	}

	assert.ElementsMatch(t, []string{
		"cpu_load_average_1_min",
		"cpu_load_average_5_min",
		"cpu_load_average_15_min",
	}, ids)
}
