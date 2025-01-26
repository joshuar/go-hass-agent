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

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

const (
	connectionLatencyWorkerID = "connection_latency"
	connectionLatencyTimeout  = 5 * time.Second

	connectionLatencyPollInterval = time.Minute
	connectionLatencyJitterAmount = 5 * time.Second

	connectionLatencyUnits = "ms"
)

var ErrEmptyResponse = errors.New("empty response")

func newConnectionLatencySensor(info resty.TraceInfo) sensor.Entity {
	return sensor.NewSensor(
		sensor.WithName("Connection Latency"),
		sensor.WithID("connection_latency"),
		sensor.WithUnits(connectionLatencyUnits),
		sensor.WithDeviceClass(types.SensorDeviceClassDuration),
		sensor.WithStateClass(types.StateClassMeasurement),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon("mdi:connection"),
			sensor.WithValue(info.TotalTime.Milliseconds()),
			sensor.WithAttribute("DNS Lookup Time", info.DNSLookup.Milliseconds()),
			sensor.WithAttribute("Connection Time", info.ConnTime.Milliseconds()),
			sensor.WithAttribute("TCP Connection Time", info.TCPConnTime.Milliseconds()),
			sensor.WithAttribute("TLS Handshake Time", info.TLSHandshake.Milliseconds()),
			sensor.WithAttribute("Server Time", info.ServerTime.Milliseconds()),
			sensor.WithAttribute("Response Time", info.ResponseTime.Milliseconds()),
			sensor.WithAttribute("native_unit_of_measurement", connectionLatencyUnits),
		),
	)
}

type ConnectionLatencySensorWorker struct {
	client   *resty.Client
	doneCh   chan struct{}
	endpoint string
	prefs    *preferences.CommonWorkerPrefs
}

func (w *ConnectionLatencySensorWorker) PreferencesID() string {
	return "connection_latency_sensor"
}

func (w *ConnectionLatencySensorWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *ConnectionLatencySensorWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
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

func NewConnectionLatencySensorWorker(_ context.Context) *ConnectionLatencySensorWorker {
	worker := &ConnectionLatencySensorWorker{
		client: resty.New().
			SetTimeout(connectionLatencyTimeout).
			EnableTrace(),
		endpoint: preferences.RestAPIURL(),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil
	}

	worker.prefs = prefs

	return worker
}
