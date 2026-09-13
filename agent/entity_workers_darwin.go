package agent

import (
	"context"
	"log/slog"
	"slices"

	"github.com/joshuar/go-hass-agent/agent/workers"
	slogctx "github.com/veqryn/slog-context"
)

var macosWorkers = []func(ctx context.Context) (workers.EntityWorker, error){}

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
