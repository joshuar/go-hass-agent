// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

// newTestUsageWorker builds a usageWorker reading a fixture instead of /proc,
// bypassing NewUsageWorker so no clktck/boottime context values are needed.
func newTestUsageWorker(path string) *usageWorker {
	return &usageWorker{
		WorkerMetadata:          models.SetWorkerMetadata("cpu_usage", "CPU usage metrics"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		path:                    path,
		rateSensors:             make(map[string]*linux.RateValue[uint64]),
		cpuSensors:              make(map[string]float64),
		clktck:                  100,
		boottime:                time.Unix(0, 0),
	}
}

func Test_usageWorker_getUsageStats(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name: "valid",
			path: "testdata/stat",
		},
		{
			name:    "unavailable",
			path:    "/nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker := newTestUsageWorker(tt.path)

			got, err := worker.getUsageStats(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			// The fixture has a total, four cores, two rates and two counts.
			assert.Len(t, got, 9)
		})
	}
}

// Test_usageWorker_getUsageStats_concurrent runs the worker the way the
// scheduler can: with a poll starting before the previous one has finished.
// That happens in practice across a suspend, which freezes an execution
// mid-flight for the resume to start the next one into.
//
// The worker keeps its previous readings in the cpuSensors and rateSensors
// maps so it can turn counters into deltas. Both are read and written on every
// execution, so without synchronisation two overlapping executions touch them
// at once and the Go runtime aborts the whole process with "fatal error:
// concurrent map read and map write" -- not a panic a caller could recover.
func Test_usageWorker_getUsageStats_concurrent(t *testing.T) {
	const (
		executions = 8
		iterations = 50
	)

	worker := newTestUsageWorker("testdata/stat")
	ctx := t.Context()

	var wg sync.WaitGroup

	start := make(chan struct{})

	for range executions {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start

			for range iterations {
				if _, err := worker.getUsageStats(ctx); err != nil {
					assert.NoError(t, err)
					return
				}
			}
		}()
	}

	close(start)
	wg.Wait()
}
