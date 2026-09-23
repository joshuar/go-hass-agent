package disk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
)

func TestSmartWorkerExecute(t *testing.T) {
	worker := &smartWorker{
		WorkerMetadata: models.SetWorkerMetadata(smartWorkerID, smartWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{
			OutCh: make(chan models.Entity, 16),
		},
	}

	err := worker.Execute(context.Background())
	require.NoError(t, err)
	close(worker.OutCh)

	var sensors []models.Sensor

	for entity := range worker.OutCh {
		sensor, err := entity.AsSensor()
		require.NoError(t, err)
		sensors = append(sensors, sensor)
	}

	require.NotEmpty(t, sensors, "expected at least one SMART sensor")
	assert.Contains(t, sensors[0].UniqueID, "_smart_status")
}
