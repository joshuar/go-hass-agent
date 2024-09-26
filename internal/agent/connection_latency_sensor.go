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
	connectionLatencyWorkerID = "connection_latency"
	connectionLatencyTimeout  = 5 * time.Second

	connectionLatencyPollInterval = time.Minute
	connectionLatencyJitterAmount = 5 * time.Second

	connectionLatencyUnits = "ms"
)

type serverPrefs interface {
	Server() string
	Token() string
}

func newConnectionLatencySensor(info resty.TraceInfo) sensor.Entity {
	connectionLatency := sensor.Entity{
		Name:        "Connection Latency",
		Units:       connectionLatencyUnits,
		DeviceClass: types.DeviceClassDuration,
		StateClass:  types.StateClassMeasurement,
		Category:    types.CategoryDiagnostic,
		EntityState: &sensor.EntityState{
			ID:         "connection_latency",
			Icon:       "mdi:connection",
			EntityType: types.Sensor,
			State:      info.TotalTime.Milliseconds(),
			Attributes: map[string]any{
				"DNS Lookup Time":            info.DNSLookup.Milliseconds(),
				"Connection Time":            info.ConnTime.Milliseconds(),
				"TCP Connection Time":        info.TCPConnTime.Milliseconds(),
				"TLS Handshake Time":         info.TLSHandshake.Milliseconds(),
				"Server Time":                info.ServerTime.Milliseconds(),
				"Response Time":              info.ResponseTime.Milliseconds(),
				"native_unit_of_measurement": connectionLatencyUnits,
			},
		},
	}

	return connectionLatency
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

func (w *connectionLatencyWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	resp, err := w.client.R().
		SetContext(ctx).
		Get("/")
	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}

	// Save the latency info as a connectionLatency sensor.

	return []sensor.Entity{newConnectionLatencySensor(resp.Request.TraceInfo())}, nil
}

func (w *connectionLatencyWorker) Start(ctx context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)
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
