// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package cpu

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
)

const (
	cpuFreqUpdateInterval = 30 * time.Second
	cpuFreqUpdateJitter   = time.Second

	cpuFreqWorkerID      = "cpu_freq_sensors"
	cpuFreqPreferencesID = prefPrefix + "frequencies"
)

var ErrInitFreqWorker = errors.New("could not start CPU frequencies worker")

type freqWorker struct{}

func (w *freqWorker) UpdateDelta(_ time.Duration) {}

func (w *freqWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	sensors := make([]models.Entity, totalCPUs)

	for i := range totalCPUs {
		entity, err := newCPUFreqSensor(ctx, "cpu"+strconv.Itoa(i))
		if err != nil {
			logging.FromContext(ctx).Warn("Could not create CPU frequency sensor: %w", err)
		}
		sensors[i] = entity
	}

	return sensors, nil
}

func (w *freqWorker) PreferencesID() string {
	return cpuFreqPreferencesID
}

func (w *freqWorker) DefaultPreferences() FreqWorkerPrefs {
	return FreqWorkerPrefs{
		UpdateInterval: cpuFreqUpdateInterval.String(),
	}
}

func NewFreqWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	freqWorker := &freqWorker{}

	prefs, err := preferences.LoadWorker(freqWorker)
	if err != nil {
		return nil, errors.Join(ErrInitFreqWorker, err)
	}

	//nolint:nilnil
	if prefs.Disabled {
		return nil, nil
	}

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", cpuFreqWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", cpuFreqUpdateInterval.String()))

		pollInterval = cpuFreqUpdateInterval
	}

	pollWorker := linux.NewPollingSensorWorker(cpuFreqWorkerID, pollInterval, cpuFreqUpdateJitter)
	pollWorker.PollingSensorType = freqWorker

	return pollWorker, nil
}
