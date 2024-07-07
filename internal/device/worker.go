// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"errors"
	"fmt"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

var ErrUnknownWorker = errors.New("unknown worker")

type WorkerControl struct {
	externalIP        *externalIPWorker
	externalIPControl context.CancelFunc
	version           *versionWorker
	versionControl    context.CancelFunc
}

//nolint:mnd
func (w *WorkerControl) ActiveWorkers() []string {
	activeWorkers := make([]string, 0, 2)

	if w.externalIPControl != nil {
		activeWorkers = append(activeWorkers, w.externalIP.Name())
	}

	if w.versionControl != nil {
		activeWorkers = append(activeWorkers, w.version.Name())
	}

	return activeWorkers
}

//nolint:mnd
func (w *WorkerControl) InactiveWorkers() []string {
	inactiveWorkers := make([]string, 0, 2)

	if w.externalIPControl == nil {
		inactiveWorkers = append(inactiveWorkers, w.externalIP.Name())
	}

	if w.versionControl == nil {
		inactiveWorkers = append(inactiveWorkers, w.version.Name())
	}

	return inactiveWorkers
}

func (w *WorkerControl) Start(ctx context.Context, name string) (<-chan sensor.Details, error) {
	workerCtx, workerCancelFunc := context.WithCancel(ctx)

	switch name {
	case w.externalIP.Name():
		workerCh, err := w.externalIP.Updates(workerCtx)
		if err != nil {
			return nil, fmt.Errorf("could not start worker: %w", err)
		}

		w.externalIPControl = workerCancelFunc

		return workerCh, nil
	case w.version.Name():
		workerCh, err := w.version.Updates(workerCtx)
		if err != nil {
			return nil, fmt.Errorf("could not start worker: %w", err)
		}

		w.externalIPControl = workerCancelFunc

		return workerCh, nil
	}

	return nil, ErrUnknownWorker
}

func (w *WorkerControl) Stop(name string) error {
	switch name {
	case w.externalIP.Name():
		w.externalIPControl()
	case w.version.Name():
		w.versionControl()
	}

	return nil
}

func (w *WorkerControl) StartAll(ctx context.Context) (<-chan sensor.Details, error) {
	var allerr error

	ipWorkerCtx, ipCancelFunc := context.WithCancel(ctx)

	ipUpdates, err := w.externalIP.Updates(ipWorkerCtx)
	if err != nil {
		allerr = errors.Join(allerr, err)
	} else {
		w.externalIPControl = ipCancelFunc
	}

	verWorkerCtx, verCancelFunc := context.WithCancel(ctx)

	verUpdates, err := w.version.Updates(verWorkerCtx)
	if err != nil {
		allerr = errors.Join(allerr, err)
	} else {
		w.versionControl = verCancelFunc
	}

	return sensor.MergeSensorCh(ctx, ipUpdates, verUpdates), allerr
}

func (w *WorkerControl) StopAll() error {
	w.externalIPControl()
	w.versionControl()

	return nil
}

//nolint:exhaustruct
func CreateSensorWorkers() *WorkerControl {
	return &WorkerControl{
		externalIP: newExternalIPUpdaterWorker(),
		version:    newVersionWorker(),
	}
}
