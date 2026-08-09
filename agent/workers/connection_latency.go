// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/hass/api"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	connectionLatencyWorkerID   = "connection_latency"
	connectionLatencyWorkerDesc = "Connection latency for Home Assistant"
	connectionLatencyTimeout    = 5 * time.Second

	connectionLatencyPollInterval = time.Minute
	connectionLatencyJitterAmount = 5 * time.Second

	connectionLatencyUnits = "ms"
)

var (
	_ quartz.Job          = (*ConnectionLatency)(nil)
	_ PollingEntityWorker = (*ConnectionLatency)(nil)
)

var ErrConnLatency = errors.New("connection latency worker error")

type ConnectionLatency struct {
	*PollingEntityWorkerData
	*models.WorkerMetadata

	endpoint string
	prefs    *CommonWorkerPrefs
}

func (w *ConnectionLatency) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *ConnectionLatency) Execute(ctx context.Context) error {
	resp, err := api.NewRequest(
		api.WithBody(api.RequestData{Type: api.GetConfig}),
		api.WithTrace(),
	).Do(ctx, w.endpoint)

	// Handle errors and bad responses.
	switch {
	case err != nil:
		return fmt.Errorf("%w: %w", ErrConnLatency, err)
	case resp.Error():
		return fmt.Errorf("%w: received error response %s", ErrConnLatency, resp.Status())
	}

	if resp.Request != nil {
		info := resp.Request.TraceInfo()
		// Save the latency info as a connectionLatency models.
		w.OutCh <- models.NewSensor(ctx,
			models.WithName("Connection Latency"),
			models.WithID("connection_latency"),
			models.WithUnits(connectionLatencyUnits),
			models.WithDeviceClass(models.SensorClassDuration),
			models.WithStateClass(models.StateMeasurement),
			models.AsDiagnostic(),
			models.WithIcon("mdi:connection"),
			models.WithState(info.TotalTime.Milliseconds()),
			models.WithAttribute("DNS Lookup Time", info.DNSLookup.Milliseconds()),
			models.WithAttribute("Connection Time", info.ConnTime.Milliseconds()),
			models.WithAttribute("TCP Connection Time", info.TCPConnTime.Milliseconds()),
			models.WithAttribute("TLS Handshake Time", info.TLSHandshake.Milliseconds()),
			models.WithAttribute("Server Time", info.ServerTime.Milliseconds()),
			models.WithAttribute("Response Time", info.ResponseTime.Milliseconds()),
			models.WithAttribute("native_unit_of_measurement", connectionLatencyUnits),
		)
	}

	return nil
}

func (w *ConnectionLatency) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

func NewConnectionLatencyWorker(_ context.Context, restAPIURL string) (EntityWorker, error) {
	worker := &ConnectionLatency{
		WorkerMetadata:          models.SetWorkerMetadata(connectionLatencyWorkerID, connectionLatencyWorkerDesc),
		PollingEntityWorkerData: &PollingEntityWorkerData{},
		endpoint:                restAPIURL,
	}

	defaultPrefs := &CommonWorkerPrefs{}
	var err error

	worker.prefs, err = LoadWorkerPreferences("sensors.agent.connection_latency", defaultPrefs)
	if err != nil {
		return worker, errors.Join(ErrConnLatency, err)
	}

	worker.Trigger = scheduler.NewPollTriggerWithJitter(connectionLatencyPollInterval, connectionLatencyJitterAmount)

	return worker, nil
}
