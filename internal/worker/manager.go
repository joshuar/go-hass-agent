// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/models"
)

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

	for _, worker := range workers {
		workerCtx, cancelFunc := context.WithCancel(ctx)
		workerCh, err := worker.Start(workerCtx)
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
	}

	return mergeCh(ctx, outCh...)
}

func (m *Manager) StartMQTTWorkers(ctx context.Context, workers ...MQTTWorker) ([]*models.MQTTConfig, []*models.MQTTSubscription, <-chan models.MQTTMsg) {
	m.mu.Lock()
	defer m.mu.Unlock()

	configs := make([]*models.MQTTConfig, 0, len(workers))
	subscriptions := make([]*models.MQTTSubscription, 0, len(workers))
	msgCh := make([]<-chan models.MQTTMsg, 0, len(workers))

	for _, worker := range workers {
		workerCtx, cancelFunc := context.WithCancel(ctx)
		workerConfigs, workerSubscriptions, workerMsgs, err := worker.Start(workerCtx)
		if err != nil {
			m.logger.Warn("Could not start worker.",
				slog.String("id", worker.ID()),
				slog.Any("errors", err))
		} else {
			m.workers[worker.ID()] = cancelFunc
			configs = append(configs, workerConfigs...)
			subscriptions = append(subscriptions, workerSubscriptions...)
			msgCh = append(msgCh, workerMsgs)
			m.logger.Debug("Started worker.",
				slog.String("id", worker.ID()))
		}
	}

	return configs, subscriptions, mergeCh(ctx, msgCh...)
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

func NewWorkerManager(ctx context.Context) *Manager {
	return &Manager{
		workers: make(map[models.ID]context.CancelFunc),
		logger:  logging.FromContext(ctx).WithGroup("worker"),
	}
}
