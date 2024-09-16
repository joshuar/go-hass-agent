// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	connectionLatencyWorkerID = "connection_latency_worker"
	connectionLatencyTimeout  = 5 * time.Second

	connectionLatencyPollInterval = time.Minute
	connectionLatencyJitterAmount = 5 * time.Second

	connectionLatencyUnits = "ms"
)

type serverPrefs interface {
	Server() string
	Token() string
}

type connectionLatency resty.TraceInfo

func (l *connectionLatency) Name() string { return "Connection Latency" }

func (l *connectionLatency) ID() string { return "connection_latency" }

func (l *connectionLatency) Icon() string { return "mdi:connection" }

func (l *connectionLatency) SensorType() types.SensorClass { return types.Sensor }

func (l *connectionLatency) DeviceClass() types.DeviceClass { return types.DeviceClassDuration }

func (l *connectionLatency) StateClass() types.StateClass { return types.StateClassMeasurement }

func (l *connectionLatency) State() any { return l.TotalTime.Milliseconds() }

func (l *connectionLatency) Units() string { return connectionLatencyUnits }

func (l *connectionLatency) Category() string { return types.CategoryDiagnostic }

func (l *connectionLatency) Attributes() map[string]any {
	return map[string]any{
		"DNS Lookup Time":     l.DNSLookup.Milliseconds(),
		"Connection Time":     l.ConnTime.Milliseconds(),
		"TCP Connection Time": l.TCPConnTime.Milliseconds(),
		"TLS Handshake Time":  l.TLSHandshake.Milliseconds(),
		"Server Time":         l.ServerTime.Milliseconds(),
		"Response Time":       l.ResponseTime.Milliseconds(),
	}
}

type connectionLatencyWorker struct {
	client *resty.Client
	doneCh chan struct{}
}

// ID returns the unique string to represent this worker and its sensors.
func (w *connectionLatencyWorker) ID() string { return connectionLatencyWorkerID }

// Stop will stop any processing of sensors controlled by this worker.
func (w *connectionLatencyWorker) Stop() error {
	close(w.doneCh)

	return nil
}

func (w *connectionLatencyWorker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	resp, err := w.client.R().
		SetContext(ctx).
		Get("/")
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}

	// Save the latency info as a connectionLatency sensor.
	latency := connectionLatency(resp.Request.TraceInfo())

	return []sensor.Details{&latency}, nil
}

func (w *connectionLatencyWorker) Start(ctx context.Context) (<-chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)
	w.doneCh = make(chan struct{})

	// Create a new context for the updates scope.
	workerCtx, cancelFunc := context.WithCancel(ctx)

	updater := func(_ time.Duration) {
		sensors, err := w.Sensors(workerCtx)
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

func newConnectionLatencyWorker(prefs serverPrefs) *connectionLatencyWorker {
	return &connectionLatencyWorker{
		client: resty.New().
			SetTimeout(connectionLatencyTimeout).
			SetBaseURL(prefs.Server() + "/api").
			SetAuthToken(prefs.Token()).
			EnableTrace(),
	}
}
