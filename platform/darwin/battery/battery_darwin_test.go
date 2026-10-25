package battery

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
)

func TestBatteryInfoPercentageAndState(t *testing.T) {
	tests := []struct {
		name    string
		info    batteryInfo
		wantPct float64
		wantSt  string
	}{
		{
			name:    "charging",
			info:    batteryInfo{CurrentCapacity: 50, MaxCapacity: 100, ExternalConnected: true, IsCharging: true},
			wantPct: 50,
			wantSt:  "Charging",
		},
		{
			name:    "discharging",
			info:    batteryInfo{CurrentCapacity: 73, MaxCapacity: 100, ExternalConnected: false},
			wantPct: 73,
			wantSt:  "Discharging",
		},
		{
			name:    "fully charged",
			info:    batteryInfo{CurrentCapacity: 100, MaxCapacity: 100, ExternalConnected: true, FullyCharged: true},
			wantPct: 100,
			wantSt:  "Fully Charged",
		},
		{
			name:    "plugged in but not charging (e.g. optimized charging hold)",
			info:    batteryInfo{CurrentCapacity: 80, MaxCapacity: 100, ExternalConnected: true},
			wantPct: 80,
			wantSt:  "Not Charging",
		},
		{
			name:    "rounds to nearest percent",
			info:    batteryInfo{CurrentCapacity: 2, MaxCapacity: 3, ExternalConnected: true, IsCharging: true},
			wantPct: 67,
			wantSt:  "Charging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.wantPct, tt.info.percentage(), 0.001)
			assert.Equal(t, tt.wantSt, tt.info.state())
		})
	}
}

func TestExecuteReal(t *testing.T) {
	worker := &batteryWorker{
		WorkerMetadata: models.SetWorkerMetadata(batteryWorkerID, batteryWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{
			OutCh: make(chan models.Entity, 4),
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

	// Mac desktops have no battery; this test is only meaningful on a
	// machine that has one, which GitHub's macOS runners and most dev
	// machines running this agent do.
	if len(sensors) == 0 {
		t.Skip("no battery present on this machine")
	}

	require.Len(t, sensors, 2)
	assert.Equal(t, "battery_level", sensors[0].UniqueID)
	assert.Equal(t, "battery_state", sensors[1].UniqueID)
}
