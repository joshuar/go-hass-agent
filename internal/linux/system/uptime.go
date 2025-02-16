// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package system

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	uptimePollInterval = 15 * time.Minute
	uptimePollJitter   = time.Minute

	uptimeWorkerID = "uptime_sensor"
)

var ErrInitUptimeWorker = errors.New("could not init uptime worker")

type uptimeWorker struct{}

func (w *uptimeWorker) UpdateDelta(_ time.Duration) {}

func (w *uptimeWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	entity, err := sensor.NewSensor(ctx,
		sensor.WithName("Uptime"),
		sensor.WithID("uptime"),
		sensor.AsDiagnostic(),
		sensor.WithDeviceClass(class.SensorClassDuration),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithUnits("h"),
		sensor.WithIcon("mdi:restart"),
		sensor.WithState(w.getUptime(ctx)/60/60),
		sensor.WithDataSourceAttribute(linux.ProcFSRoot),
		sensor.WithAttribute("native_unit_of_measurement", "h"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not generate uptime sensor: %w", err)
	}

	return []models.Entity{entity}, nil
}

func (w *uptimeWorker) PreferencesID() string {
	return infoWorkerPreferencesID
}

func (w *uptimeWorker) DefaultPreferences() UptimePrefs {
	return UptimePrefs{
		UpdateInterval: uptimePollInterval.String(),
	}
}

// getUptime retrieve the uptime of the device running Go Hass Agent, in
// seconds. If the uptime cannot be retrieved, it will return 0.
func (w *uptimeWorker) getUptime(ctx context.Context) float64 {
	data, err := os.Open(linux.UptimeFile)
	if err != nil {
		logging.FromContext(ctx).Debug("Unable to retrieve uptime.", slog.Any("error", err))

		return 0
	}

	defer data.Close()

	line := bufio.NewScanner(data)
	line.Split(bufio.ScanWords)

	if !line.Scan() {
		logging.FromContext(ctx).Debug("Could not parse uptime.")

		return 0
	}

	uptimeValue, err := strconv.ParseFloat(line.Text(), 64)
	if err != nil {
		logging.FromContext(ctx).Debug("Could not parse uptime.")

		return 0
	}

	return uptimeValue
}

func NewUptimeTimeWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	uptimeWorker := &uptimeWorker{}

	prefs, err := preferences.LoadWorker(uptimeWorker)
	if err != nil {
		return nil, errors.Join(ErrInitUptimeWorker, err)
	}

	//nolint:nilnil
	if prefs.IsDisabled() {
		return nil, nil
	}

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", uptimeWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", uptimePollInterval.String()))

		pollInterval = uptimePollInterval
	}

	worker := linux.NewPollingSensorWorker(uptimeWorkerID, pollInterval, uptimePollJitter)
	worker.PollingSensorType = uptimeWorker

	return worker, nil
}
