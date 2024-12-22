// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
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

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	chronyPollInterval = 5 * time.Minute
	chronyPollJitter   = 10 * time.Second

	chronyWorkerID = "chrony_sensors"

	sensorStat = "System time"
)

type chronyWorker struct {
	chronycPath string
}

//revive:disable:unused-receiver
func (w *chronyWorker) UpdateDelta(_ time.Duration) {}

func (w *chronyWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	// Get chrony tracking stats via chronyc.
	stats, err := w.getChronyTrackingStats()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve chrony stats: %w", err)
	}
	// Generate a sensor.
	return []sensor.Entity{newChronyOffsetSensor(stats)}, nil
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
func newChronyOffsetSensor(stats map[string]string) sensor.Entity {
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

	return sensor.NewSensor(
		sensor.WithName("Chrony System Time Offset"),
		sensor.WithID("chrony_system_time_offset"),
		sensor.AsDiagnostic(),
		sensor.WithUnits("s"),
		sensor.WithStateClass(types.StateClassMeasurement),
		sensor.WithState(
			sensor.WithIcon("mdi:clock"),
			sensor.WithValue(value),
			sensor.WithAttributes(attrs),
		),
	)
}

// NewChronyWorker creates a worker to track sensors from chronyd.
func NewChronyWorker(_ context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingSensorWorker(chronyWorkerID, chronyPollInterval, chronyPollJitter)

	path, err := exec.LookPath("chronyc")
	if err != nil {
		return worker, fmt.Errorf("chronyc is not available: %w", err)
	}

	worker.PollingSensorType = &chronyWorker{chronycPath: path}

	return worker, nil
}
