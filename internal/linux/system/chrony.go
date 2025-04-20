// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
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

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	chronyPollInterval = 5 * time.Minute
	chronyPollJitter   = 10 * time.Second

	chronyWorkerID      = "chrony_sensors"
	chronyWorkerDesc    = "Chrony stats"
	chronyPreferencesID = sensorsPrefPrefix + "chrony"

	sensorStat = "System time"
)

var (
	_ quartz.Job                  = (*chronyWorker)(nil)
	_ workers.PollingEntityWorker = (*chronyWorker)(nil)
)

var ErrInitChronyWorker = errors.New("could not init chrony worker")

type chronyWorker struct {
	chronycPath string
	prefs       *ChronyPrefs
	*workers.PollingEntityWorkerData
	*models.WorkerMetadata
}

func (w *chronyWorker) Execute(ctx context.Context) error {
	// Get chrony tracking stats via chronyc.
	stats, err := w.getChronyTrackingStats()
	if err != nil {
		return fmt.Errorf("could not retrieve chrony stats: %w", err)
	}
	// Generate a sensor.
	entity, err := newChronyOffsetSensor(ctx, stats)
	if err != nil {
		return fmt.Errorf("could not generate chrony sensor: %w", err)
	}
	w.OutCh <- entity
	return nil
}

func (w *chronyWorker) PreferencesID() string {
	return chronyPreferencesID
}

func (w *chronyWorker) DefaultPreferences() ChronyPrefs {
	return ChronyPrefs{
		UpdateInterval: chronyPollInterval.String(),
	}
}

func (w *chronyWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *chronyWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk IO worker: %w", err)
	}
	return w.OutCh, nil
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
func NewChronyWorker(ctx context.Context) (workers.EntityWorker, error) {
	path, err := exec.LookPath("chronyc")
	if err != nil {
		return nil, errors.Join(ErrInitChronyWorker,
			fmt.Errorf("chronyc is not available: %w", err))
	}

	worker := &chronyWorker{
		WorkerMetadata:          models.SetWorkerMetadata(chronyWorkerID, chronyWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		chronycPath:             path,
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitChronyWorker, err)
	}
	worker.prefs = prefs

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", chronyWorkerID),
			slog.String("given_interval", worker.prefs.UpdateInterval),
			slog.String("default_interval", chronyPollInterval.String()))

		pollInterval = chronyPollInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, chronyPollJitter)

	return worker, nil
}
