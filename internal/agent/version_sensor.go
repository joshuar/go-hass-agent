// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

const (
	versionWorkerID = "agent_version_sensor"
)

type versionWorker struct {
	version
}

func (w *versionWorker) ID() string { return versionWorkerID }

func (w *versionWorker) Stop() error { return nil }

func (w *versionWorker) Start(_ context.Context) (<-chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	go func() {
		defer close(sensorCh)
		sensorCh <- w
	}()

	return sensorCh, nil
}

func (w *versionWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return []sensor.Details{&w.version}, nil
}

func newVersionWorker(value string) *versionWorker {
	return &versionWorker{version: version(value)}
}
