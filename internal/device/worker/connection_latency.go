// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	connectionLatencyWorkerID = "connection_latency"
	connectionLatencyTimeout  = 5 * time.Second

	connectionLatencyPollInterval = time.Minute
	connectionLatencyJitterAmount = 5 * time.Second

	connectionLatencyUnits = "ms"
)

var (
	ErrInitConnLatencyWorker = errors.New("could not init connection latency worker")
	ErrEmptyResponse         = errors.New("empty response")
)

type ConnectionLatency struct {
	client   *resty.Client
	doneCh   chan struct{}
	endpoint string
	prefs    *preferences.CommonWorkerPrefs
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

// ID returns the unique string to represent this worker and its sensors.
func (w *ConnectionLatency) ID() string { return connectionLatencyWorkerID }

func (w *ConnectionLatency) sensors(ctx context.Context) ([]models.Entity, error) {
	resp, err := w.client.R().
		SetContext(ctx).SetBody(api.Request{Type: api.GetConfig}).
		Post(w.endpoint)

	// Handle errors and bad responses.
	switch {
	case err != nil:
		return nil, fmt.Errorf("unable to connect: %w", err)
	case resp.Error():
		return nil, fmt.Errorf("received error response %s", resp.Status())
	}

	if resp.Request != nil {
		entity, err := newConnectionLatencySensor(ctx, resp.Request.TraceInfo())
		if err != nil {
			return nil, err
		}
		// Save the latency info as a connectionLatency models.
		return []models.Entity{entity}, nil
	}

	return nil, ErrEmptyResponse
}

func (w *ConnectionLatency) Start(ctx context.Context) (<-chan models.Entity, error) {
	sensorCh := make(chan models.Entity)
	w.doneCh = make(chan struct{})

	// Create a new context for the updates scope.
	workerCtx, cancelFunc := context.WithCancel(ctx)

	updater := func(_ time.Duration) {
		sensors, err := w.sensors(workerCtx)
		if err != nil {
			logging.FromContext(workerCtx).
				With(slog.String("worker", connectionLatencyWorkerID)).
				Debug("Could not generated latency sensors.", slog.Any("error", err))
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}

	go func() {
		helpers.PollSensors(ctx, updater, connectionLatencyPollInterval, connectionLatencyJitterAmount)
	}()

	go func() {
		<-workerCtx.Done()
		cancelFunc()
		close(sensorCh)
	}()

	return sensorCh, nil
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

func NewConnectionLatencyWorker(_ context.Context) (*ConnectionLatency, error) {
	worker := &ConnectionLatency{
		client: resty.New().
			SetTimeout(connectionLatencyTimeout).
			EnableTrace(),
		endpoint: preferences.RestAPIURL(),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return worker, errors.Join(ErrInitConnLatencyWorker, err)
	}

	worker.prefs = prefs

	return worker, nil
}
