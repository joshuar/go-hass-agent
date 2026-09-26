package agent

import (
	"context"
	"log/slog"
	"slices"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/platform/darwin/disk"
	"github.com/joshuar/go-hass-agent/platform/darwin/temperature"
)

var macosWorkers = []func(ctx context.Context) (workers.EntityWorker, error){
	disk.NewUsageWorker,
	disk.NewSmartWorker,
	temperature.NewTemperatureWorker,
}

func CreateOSEntityWorkers(ctx context.Context) []workers.EntityWorker {
	osWorkers := make([]workers.EntityWorker, 0, len(macosWorkers))

	for workerInit := range slices.Values(macosWorkers) {
		worker, err := workerInit(ctx)

		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not init worker.",
				slog.String("worker", worker.ID()),
				slog.Any("error", err))

			continue
		}

		osWorkers = append(osWorkers, worker)
	}

	return osWorkers
}
