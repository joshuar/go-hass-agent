// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	chronyPollInterval = 5 * time.Minute
	chronyPollJitter   = 10 * time.Second

	chronyWorkerID      = "chrony_sensors"
	chronyPreferencesID = sensorsPrefPrefix + "chrony"

	sensorStat = "System time"
)

var ErrInitChronyWorker = errors.New("could not init chrony worker")

type chronyWorker struct {
	chronycPath string
	prefs       *ChronyPrefs
}

//revive:disable:unused-receiver
func (w *chronyWorker) UpdateDelta(_ time.Duration) {}

func (w *chronyWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	// Get chrony tracking stats via chronyc.
	stats, err := w.getChronyTrackingStats()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve chrony stats: %w", err)
	}
	// Generate a sensor.
	entity, err := newChronyOffsetSensor(ctx, stats)
	if err != nil {
		return nil, fmt.Errorf("could not generate chrony sensor: %w", err)
	}

	return []models.Entity{entity}, nil
}

func (w *chronyWorker) PreferencesID() string {
	return chronyPreferencesID
}

func (w *chronyWorker) DefaultPreferences() ChronyPrefs {
	return ChronyPrefs{
		UpdateInterval: chronyPollInterval.String(),
	}
}

// getChronyTrackingStats executes chronyc, parsing its output and returning the
// individual stats in a map.
func (w *chronyWorker) getChronyTrackingStats() (map[string]string, error) {
	chronycOutput, err := exec.Command(w.chronycPath, "-n", "tracking").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run chronyc: %w", err)
	}

	stats := make(map[string]string)

	lines := bufio.NewScanner(bytes.NewBuffer(chronycOutput))
	for lines.Scan() {
		value := strings.Split(lines.Text(), ":")
		statName := strings.TrimSpace(value[0])
		statValue := strings.TrimSpace(value[1])
		stats[statName] = statValue
	}

	return stats, nil
}

// newChronyOffsetSensor creates a new sensor representing the system clock
// offset from the NTP server time. Attributes contain other stats acquired from
// chrony.
func newChronyOffsetSensor(ctx context.Context, stats map[string]string) (models.Entity, error) {
	var value any

	// Try to parse the value into a float. If failed, use the string value.
	valueArr := strings.Split(stats[sensorStat], " ")
	valueParsed, err := strconv.ParseFloat(valueArr[0], 64)

	if err != nil {
		value = valueArr[0]
	} else {
		value = valueParsed
	}

	// Base sensor attributes.
	attrs := map[string]any{
		"native_unit_of_measurement": "s",
		"data_source":                "chrony",
	}
	// Add other chrony stats as sensor attributes.
	for attr, value := range stats {
		if attr == sensorStat {
			continue
		}

		attrs[attr] = value
	}

	return sensor.NewSensor(ctx,
		sensor.WithName("Chrony System Time Offset"),
		sensor.WithID("chrony_system_time_offset"),
		sensor.AsDiagnostic(),
		sensor.WithUnits("s"),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithIcon("mdi:clock"),
		sensor.WithState(value),
		sensor.WithAttributes(attrs),
	)
}

// NewChronyWorker creates a worker to track sensors from chronyd.
func NewChronyWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	path, err := exec.LookPath("chronyc")
	if err != nil {
		return nil, errors.Join(ErrInitChronyWorker,
			fmt.Errorf("chronyc is not available: %w", err))
	}

	chronyWorker := &chronyWorker{chronycPath: path}

	chronyWorker.prefs, err = preferences.LoadWorker(chronyWorker)
	if err != nil {
		return nil, errors.Join(ErrInitChronyWorker, err)
	}

	//nolint:nilnil
	if chronyWorker.prefs.IsDisabled() {
		return nil, nil
	}

	pollInterval, err := time.ParseDuration(chronyWorker.prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", chronyWorkerID),
			slog.String("given_interval", chronyWorker.prefs.UpdateInterval),
			slog.String("default_interval", chronyPollInterval.String()))

		pollInterval = chronyPollInterval
	}

	worker := linux.NewPollingSensorWorker(chronyWorkerID, pollInterval, chronyPollJitter)
	worker.PollingSensorType = chronyWorker

	return worker, nil
}
