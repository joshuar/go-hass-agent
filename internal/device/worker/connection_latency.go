// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
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
	_ quartz.Job                  = (*ConnectionLatency)(nil)
	_ workers.PollingEntityWorker = (*ConnectionLatency)(nil)
)

var ErrConnLatency = errors.New("connection latency worker error")

type ConnectionLatency struct {
	client   *resty.Client
	endpoint string
	prefs    *preferences.CommonWorkerPrefs
	*workers.PollingEntityWorkerData
	*models.WorkerMetadata
}

func (w *ConnectionLatency) PreferencesID() string {
	return preferences.SensorsPrefPrefix + "agent" + preferences.PathDelim + "connection_latency"
}

func (w *ConnectionLatency) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *ConnectionLatency) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *ConnectionLatency) Execute(ctx context.Context) error {
	resp, err := w.client.R().
		SetContext(ctx).SetBody(api.Request{Type: api.GetConfig}).
		Post(w.endpoint)

	// Handle errors and bad responses.
	switch {
	case err != nil:
		return fmt.Errorf("%w: %w", ErrConnLatency, err)
	case resp.Error():
		return fmt.Errorf("%w: received error response %s", ErrConnLatency, resp.Status())
	}

	if resp.Request != nil {
		entity, err := newConnectionLatencySensor(ctx, resp.Request.TraceInfo())
		if err != nil {
			return err
		}
		// Save the latency info as a connectionLatency models.
		w.OutCh <- entity
	}

	return nil
}

func (w *ConnectionLatency) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

func newConnectionLatencySensor(ctx context.Context, info resty.TraceInfo) (models.Entity, error) {
	entity, err := sensor.NewSensor(ctx,
		sensor.WithName("Connection Latency"),
		sensor.WithID("connection_latency"),
		sensor.WithUnits(connectionLatencyUnits),
		sensor.WithDeviceClass(class.SensorClassDuration),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:connection"),
		sensor.WithState(info.TotalTime.Milliseconds()),
		sensor.WithAttribute("DNS Lookup Time", info.DNSLookup.Milliseconds()),
		sensor.WithAttribute("Connection Time", info.ConnTime.Milliseconds()),
		sensor.WithAttribute("TCP Connection Time", info.TCPConnTime.Milliseconds()),
		sensor.WithAttribute("TLS Handshake Time", info.TLSHandshake.Milliseconds()),
		sensor.WithAttribute("Server Time", info.ServerTime.Milliseconds()),
		sensor.WithAttribute("Response Time", info.ResponseTime.Milliseconds()),
		sensor.WithAttribute("native_unit_of_measurement", connectionLatencyUnits),
	)
	if err != nil {
		return entity, fmt.Errorf("could not create connection latency sensor: %w", err)
	}

	return entity, nil
}

func NewConnectionLatencyWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &ConnectionLatency{
		WorkerMetadata:          models.SetWorkerMetadata(connectionLatencyWorkerID, connectionLatencyWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		client: resty.New().
			SetTimeout(connectionLatencyTimeout).
			EnableTrace(),
		endpoint: preferences.RestAPIURL(),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return worker, errors.Join(ErrConnLatency, err)
	}
	worker.prefs = prefs

	worker.Trigger = scheduler.NewPollTriggerWithJitter(connectionLatencyPollInterval, connectionLatencyJitterAmount)

	return worker, nil
}
