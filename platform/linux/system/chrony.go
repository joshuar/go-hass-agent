// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package system

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	chronyPollInterval = 5 * time.Minute
	chronyPollJitter   = 10 * time.Second

	chronyPreferencesID = sensorsPrefPrefix + "chrony"

	sensorStat = "System time"
)

var (
	_ quartz.Job                  = (*chronyWorker)(nil)
	_ workers.PollingEntityWorker = (*chronyWorker)(nil)
)

type chronyWorker struct {
	*workers.PollingEntityWorkerData
	*models.WorkerMetadata

	chronycPath string
	prefs       *ChronyPrefs
}

// NewChronyWorker creates a worker to track sensors from chronyd.
func NewChronyWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &chronyWorker{
		WorkerMetadata:          models.SetWorkerMetadata("chrony", "Chrony stats"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	var err error

	worker.chronycPath, err = exec.LookPath("chronyc")
	if err != nil {
		return worker, fmt.Errorf("find chrony executable: %w", err)
	}

	defaultPrefs := &ChronyPrefs{
		UpdateInterval: chronyPollInterval.String(),
	}
	worker.prefs, err = workers.LoadWorkerPreferences(chronyPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = chronyPollInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, chronyPollJitter)

	return worker, nil
}

func (w *chronyWorker) Execute(ctx context.Context) error {
	// Get chrony tracking stats via chronyc.
	stats, err := w.getChronyTrackingStats()
	if err != nil {
		return fmt.Errorf("could not retrieve chrony stats: %w", err)
	}
	// Generate a sensor.
	w.OutCh <- newChronyOffsetSensor(ctx, stats)
	return nil
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
func newChronyOffsetSensor(ctx context.Context, stats map[string]string) models.Entity {
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
