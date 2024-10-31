// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package agentsensor

import (
	"context"
	"errors"
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

var ErrEmptyResponse = errors.New("empty response")

type serverPrefs interface {
	RestAPIURL() string
}

func newConnectionLatencySensor(info resty.TraceInfo) sensor.Entity {
	connectionLatency := sensor.Entity{
		Name:        "Connection Latency",
		Units:       connectionLatencyUnits,
		DeviceClass: types.SensorDeviceClassDuration,
		StateClass:  types.StateClassMeasurement,
		Category:    types.CategoryDiagnostic,
		State: &sensor.State{
			ID:         "connection_latency",
			Icon:       "mdi:connection",
			EntityType: types.Sensor,
			Value:      info.TotalTime.Milliseconds(),
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

type ConnectionLatencySensorWorker struct {
	client   *resty.Client
	doneCh   chan struct{}
	endpoint string
}

// TODO: implement ability to disable.
func (w *ConnectionLatencySensorWorker) Disabled() bool {
	return false
}

// ID returns the unique string to represent this worker and its sensors.
func (w *ConnectionLatencySensorWorker) ID() string { return connectionLatencyWorkerID }

// Stop will stop any processing of sensors controlled by this worker.
func (w *ConnectionLatencySensorWorker) Stop() error {
	close(w.doneCh)

	return nil
}

func (w *ConnectionLatencySensorWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	resp, err := w.client.R().
		SetContext(ctx).
		Get(w.endpoint)

	// Handle errors and bad responses.
	switch {
	case err != nil:
		return nil, fmt.Errorf("unable to connect: %w", err)
	case resp.Error():
		return nil, fmt.Errorf("received error response %s", resp.Status())
	}

	if resp.Request != nil {
		// Save the latency info as a connectionLatency sensor.
		return []sensor.Entity{newConnectionLatencySensor(resp.Request.TraceInfo())}, nil
	}

	return nil, ErrEmptyResponse
}

func (w *ConnectionLatencySensorWorker) Start(ctx context.Context) (<-chan sensor.Entity, error) {
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

func NewConnectionLatencySensorWorker(prefs serverPrefs) *ConnectionLatencySensorWorker {
	return &ConnectionLatencySensorWorker{
		client: resty.New().
			SetTimeout(connectionLatencyTimeout).
			EnableTrace(),
		endpoint: prefs.RestAPIURL(),
	}
}
