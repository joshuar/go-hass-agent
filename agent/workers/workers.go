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
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/id"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/scheduler"
)

// Worker contains the common methods that define a worker.
type Worker interface {
	// IsDisabled returns a boolean indicating whether the worker has been disabled (i.e, through preferences).
	IsDisabled() bool
}

// CommonWorkerPrefs contains worker preferences that all workers can/should
// implement. For e.g., a toggle to completely disable the worker.
type CommonWorkerPrefs struct {
	Disabled bool `toml:"disabled"`
}

// IsDisabled will return whether the worker is disabled.
func (p *CommonWorkerPrefs) IsDisabled() bool {
	return p.Disabled
}

// LoadWorkerPreferences handles loading preferences from file for the given worker path in the file, into the given worker preferences object.
func LoadWorkerPreferences[T any](path string, preferences T) (T, error) {
	if !config.Exists(path) {
		err := SaveWorkerPreferences(path, preferences)
		if err != nil {
			return preferences, fmt.Errorf("could not save new preferences: %w", err)
		}
	}
	// Load the server config.
	if err := config.Load(path, preferences); err != nil {
		return preferences, fmt.Errorf("unable to load %s preferences: %w", path, err)
	}
	return preferences, nil
}

// SaveWorkerPreferences handles saving the given worker preferences to file at the given path.
func SaveWorkerPreferences[T any](path string, preferences T) error {
	err := config.Save(path, preferences)
	if err != nil {
		return fmt.Errorf("unable to save %s preferences: %w", path, err)
	}
	return nil
}

// EntityWorker is a worker that produces entities.
type EntityWorker interface {
	Worker
	// Start will run the worker. When the worker needs to be stopped, the
	// passed-in context should be canceled and the worker cleans itself up. If
	// the worker cannot be started, a non-nill error is returned.
	Start(ctx context.Context) (<-chan models.Entity, error)
	ID() string
}

// PollingEntityWorker is an entity worker that generates entities via polling for data on a schedule.
type PollingEntityWorker interface {
	EntityWorker
	quartz.Job
	GetTrigger() quartz.Trigger
	ID() string
}

// PollingEntityWorkerData contains the data for handling polling.
type PollingEntityWorkerData struct {
	Trigger      quartz.Trigger
	OutCh        chan models.Entity
	LastFireTime time.Time
}

// GetTrigger returns the poll trigger for the worker.
func (d *PollingEntityWorkerData) GetTrigger() quartz.Trigger {
	return d.Trigger
}

// GetDelta returns a duration indicating the time that has passed since the data was last polled.
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
		return fmt.Errorf("could not schedule polling worker %s: %w", worker.ID(), err)
	}
	// Clean-up on agent close.
	go func() {
		defer close(outCh)
		<-ctx.Done()
	}()
	// Send initial update.
	go func() {
		if err := worker.Execute(ctx); err != nil {
			slogctx.FromCtx(ctx).Warn("Could not send initial polling worker update.",
				slog.String("worker", worker.ID()),
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

// Manager tracks running workers.
type Manager struct {
	mu sync.Mutex

	workerCancelFuncs []context.CancelFunc
}

// NewManager creates a new manager object.
func NewManager() *Manager {
	return &Manager{
		workerCancelFuncs: make([]context.CancelFunc, 0),
	}
}

// StartEntityWorkers starts the given EntityWorkers. Any errors will be logged.
func (m *Manager) StartEntityWorkers(ctx context.Context, workers ...EntityWorker) <-chan models.Entity {
	m.mu.Lock()
	defer m.mu.Unlock()

	outCh := make([]<-chan models.Entity, 0, len(workers))

	for worker := range slices.Values(workers) {
		if worker.IsDisabled() {
			continue
		}
		workerCtx, cancelFunc := context.WithCancel(ctx)
		workerCh, err := worker.Start(workerCtx)
		if workerCh == nil {
			cancelFunc()
			continue
		}
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not start entity worker.",
				slog.String("worker", worker.ID()),
				slog.Any("errors", err))
		} else {
			m.workerCancelFuncs = append(m.workerCancelFuncs, cancelFunc)
			outCh = append(outCh, workerCh)
		}
		go func() {
			defer cancelFunc()
			<-ctx.Done()
		}()
	}

	return MergeCh(ctx, outCh...)
}

// StartMQTTWorkers starts the given MQTTWorkers. Any errors will be logged.
func (m *Manager) StartMQTTWorkers(ctx context.Context, workers ...MQTTWorker) *mqtt.WorkerData {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := &mqtt.WorkerData{}
	msgCh := make([]<-chan models.MQTTMsg, 0, len(workers))

	for worker := range slices.Values(workers) {
		if worker.IsDisabled() {
			continue
		}
		workerCtx, cancelFunc := context.WithCancel(ctx)
		workerCtx = slogctx.NewCtx(workerCtx,
			slogctx.FromCtx(workerCtx).WithGroup("mqtt_worker"))
		workerData, err := worker.Start(workerCtx)
		if err != nil {
			slogctx.FromCtx(ctx).Warn("Could not start mqtt worker.",
				slog.Any("errors", err))
			cancelFunc()
			continue
		}
		m.workerCancelFuncs = append(m.workerCancelFuncs, cancelFunc)
		// Add MQTT worker configs.
		data.Configs = append(data.Configs, workerData.Configs...)
		// Add MQTT worker subscriptions.
		data.Subscriptions = append(data.Subscriptions, workerData.Subscriptions...)
		// Add MQTT worker message channel, if created.
		if workerData.Msgs != nil {
			msgCh = append(msgCh, workerData.Msgs)
		}
		go func() {
			defer cancelFunc()
			<-ctx.Done()
		}()
	}
	// Merge all worker message channels.
	data.Msgs = MergeCh(ctx, msgCh...)

	return data
}

// StopAllWorkers stops all workers.
func (m *Manager) StopAllWorkers() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for workerCancelFunc := range slices.Values(m.workerCancelFuncs) {
		workerCancelFunc()
	}
}

// MergeCh merges a list of channels of any type into a single channel of that
// type (channel fan-in).
func MergeCh[T any](ctx context.Context, inCh ...<-chan T) chan T {
	var wg sync.WaitGroup

	outCh := make(chan T)

	// Start an output goroutine for each input channel in sensorCh.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(ch <-chan T) {
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
