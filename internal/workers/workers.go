// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package workers

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/internal/components/id"
	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/mqtt"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
)

type Worker interface {
	// ID returns an ID for the worker.
	ID() models.ID
	// IsDisabled returns a boolean indicating whether the worker has been disabled (i.e, through preferences).
	IsDisabled() bool
}

// EntityWorker is a worker that produces entities.
type EntityWorker interface {
	Worker
	// Start will run the worker. When the worker needs to be stopped, the
	// passed-in context should be canceled and the worker cleans itself up. If
	// the worker cannot be started, a non-nill error is returned.
	Start(ctx context.Context) (<-chan models.Entity, error)
}

type PollingEntityWorker interface {
	EntityWorker
	quartz.Job
	GetTrigger() quartz.Trigger
}

type PollingEntityWorkerData struct {
	Trigger      quartz.Trigger
	OutCh        chan models.Entity
	LastFireTime time.Time
}

func (d *PollingEntityWorkerData) GetTrigger() quartz.Trigger {
	return d.Trigger
}

func (d *PollingEntityWorkerData) GetDelta() time.Duration {
	delta := time.Since(d.LastFireTime)
	d.LastFireTime = time.Now()
	return delta
}

// SchedulePollingWorker handles submission of a polling entity worker to the quartz job scheduler. If the worker cannot
// be submitted as a job, a non-nil error is returned.
func SchedulePollingWorker(ctx context.Context, worker PollingEntityWorker, outCh chan models.Entity) error {
	// Schedule worker.
	err := scheduler.Manager.ScheduleJob(id.Worker, worker, worker.GetTrigger())
	if err != nil {
		return fmt.Errorf("could not start worker %s: %w", worker.ID(), err)
	}
	// Clean-up on agent close.
	go func() {
		defer close(outCh)
		<-ctx.Done()
	}()
	// Send initial update.
	go func() {
		if err := worker.Execute(ctx); err != nil {
			logging.FromContext(ctx).Warn("Could not send initial worker update.",
				slog.String("worker_id", worker.ID()),
				slog.Any("error", err))
		}
	}()
	return nil
}

// MQTTWorker is a worker that manages some MQTT functionality.
type MQTTWorker interface {
	Worker
	// Start will run the worker. When the worker needs to be stopped, the
	// passed-in context should be canceled and the worker cleans itself up. If
	// the worker cannot be started, a non-nill error is returned.
	Start(ctx context.Context) (*mqtt.WorkerData, error)
}

type Manager struct {
	workers map[models.ID]context.CancelFunc
	logger  *slog.Logger
	mu      sync.Mutex
}

// StartEntityWorkers starts the given EntityWorkers. Any errors will be logged.
func (m *Manager) StartEntityWorkers(ctx context.Context, workers ...EntityWorker) <-chan models.Entity {
	m.mu.Lock()
	defer m.mu.Unlock()

	outCh := make([]<-chan models.Entity, 0, len(workers))

	for worker := range slices.Values(workers) {
		if worker.IsDisabled() {
			m.logger.Warn("Not starting disabled worker.",
				slog.String("id", worker.ID()))
			continue
		}
		workerCtx, cancelFunc := context.WithCancel(ctx)
		workerCh, err := worker.Start(workerCtx)
		if workerCh == nil {
			cancelFunc()
			continue
		}
		if err != nil {
			m.logger.Warn("Could not start worker.",
				slog.String("id", worker.ID()),
				slog.Any("errors", err))
		} else {
			m.workers[worker.ID()] = cancelFunc
			outCh = append(outCh, workerCh)
			m.logger.Debug("Started worker.",
				slog.String("id", worker.ID()))
		}
		go func() {
			defer cancelFunc()
			<-ctx.Done()
		}()
	}

	return MergeCh(ctx, outCh...)
}

func (m *Manager) StartMQTTWorkers(ctx context.Context, workers ...MQTTWorker) *mqtt.WorkerData {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := &mqtt.WorkerData{}
	msgCh := make([]<-chan models.MQTTMsg, 0, len(workers))

	for worker := range slices.Values(workers) {
		if worker.IsDisabled() {
			m.logger.Warn("Not starting disabled worker.",
				slog.String("id", worker.ID()))
			continue
		}
		workerCtx, cancelFunc := context.WithCancel(ctx)
		workerData, err := worker.Start(workerCtx)
		if err != nil {
			m.logger.Warn("Could not start worker.",
				slog.String("id", worker.ID()),
				slog.Any("errors", err))
			cancelFunc()
			continue
		}
		m.workers[worker.ID()] = cancelFunc
		// Add MQTT worker configs.
		data.Configs = append(data.Configs, workerData.Configs...)
		// Add MQTT worker subscriptions.
		data.Subscriptions = append(data.Subscriptions, workerData.Subscriptions...)
		// Add MQTT worker message channel, if created.
		if workerData.Msgs != nil {
			msgCh = append(msgCh, workerData.Msgs)
		}
		m.logger.Debug("Started worker.",
			slog.String("id", worker.ID()))
		go func() {
			defer cancelFunc()
			<-ctx.Done()
		}()
	}
	// Merge all worker message channels.
	data.Msgs = MergeCh(ctx, msgCh...)

	return data
}

// StopWorkers stops the workers with the given IDs. If the worker is
// already stopped or not running, a warning will be logged and the action is a
// no-op.
func (m *Manager) StopAllWorkers() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, workerCancelFunc := range m.workers {
		workerCancelFunc()
		m.logger.Debug("Stopped worker.",
			slog.String("id", id))
	}
}

// StopWorkers stops the workers with the given IDs. If the worker is
// already stopped or not running, a warning will be logged and the action is a
// no-op.
func (m *Manager) StopWorkers(ids ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range ids {
		if workerCancelFunc, found := m.workers[id]; found {
			workerCancelFunc()
			m.logger.Debug("Stopped worker.",
				slog.String("id", id))
		} else {
			m.logger.Warn("Unknown worker or worker not running.",
				slog.String("id", id))
		}
	}
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{
		workers: make(map[models.ID]context.CancelFunc),
		logger:  logging.FromContext(ctx).WithGroup("worker"),
	}
}

// MergeCh merges a list of channels of any type into a single channel of that
// type (channel fan-in).
func MergeCh[T any](ctx context.Context, inCh ...<-chan T) chan T {
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
