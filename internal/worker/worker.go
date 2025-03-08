// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package worker

import (
	"context"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/models"
)

type Worker interface {
	// ID returns an ID for the worker.
	ID() models.ID
}

// EntityWorker is a worker that produces entities.
type EntityWorker interface {
	Worker
	// Start will run the worker. When the worker needs to be stopped, the
	// passed-in context should be canceled and the worker cleans itself up. If
	// the worker cannot be started, a non-nill error is returned.
	Start(ctx context.Context) (<-chan models.Entity, error)
}

// MQTTWorker is a worker that manages some MQTT functionality.
type MQTTWorker interface {
	Worker
	// Start will run the worker. When the worker needs to be stopped, the
	// passed-in context should be canceled and the worker cleans itself up. If
	// the worker cannot be started, a non-nill error is returned.
	Start(ctx context.Context) ([]*models.MQTTConfig, []*models.MQTTSubscription, <-chan models.MQTTMsg, error)
}

// mergeCh merges a list of channels of any type into a single channel of that
// type (channel fan-in).
func mergeCh[T any](ctx context.Context, inCh ...<-chan T) chan T {
	var wg sync.WaitGroup

	outCh := make(chan T)

	// Start an output goroutine for each input channel in sensorCh.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(ch <-chan T) { //nolint:varnamelen
		defer wg.Done()

		if ch == nil {
			return
		}

		for n := range ch {
			select {
			case outCh <- n:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(inCh))

	for _, c := range inCh {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
