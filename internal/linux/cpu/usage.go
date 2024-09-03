// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cpu

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	usageUpdateInterval = 10 * time.Second
	usageUpdateJitter   = 500 * time.Millisecond

	usageWorkerID = "cpu_usage_sensors"

	totalCPUString = "cpu"
)

//nolint:lll
var times = [...]string{"user_time", "nice_time", "system_time", "idle_time", "iowait_time", "irq_time", "softirq_time", "steal_time", "guest_time", "guest_nice_time"}

type cpuUsageSensor struct {
	cpuID           string
	usageAttributes map[string]any
	linux.Sensor
}

func (s *cpuUsageSensor) generateValues(clktck int64, details []string) {
	var totalTime float64

	// Don't calculate values if the length of the details array doesn't match
	// what we expect.
	if len(details) != len(times) {
		return
	}

	attrs := make(map[string]any, len(times))

	for idx, name := range times {
		value, err := strconv.ParseFloat(details[idx], 64)
		if err != nil {
			continue
		}

		cpuTime := value / float64(clktck)
		attrs[name] = cpuTime
		totalTime += cpuTime
	}

	attrs["total_time"] = totalTime
	s.usageAttributes = attrs

	//nolint:forcetypeassert,mnd // we already parsed the value as a float
	s.Value = attrs["user_time"].(float64) / totalTime * 100
}

func (s *cpuUsageSensor) Name() string {
	switch {
	case s.cpuID == totalCPUString:
		return "Total CPU Usage"
	default:
		return "Core " + strings.TrimPrefix(s.cpuID, "cpu") + " CPU Usage"
	}
}

func (s *cpuUsageSensor) ID() string {
	switch {
	case s.cpuID == totalCPUString:
		return "total_cpu_usage"
	default:
		return "core_" + strings.TrimPrefix(s.cpuID, "cpu") + "_cpu_usage"
	}
}

func (s *cpuUsageSensor) Attributes() map[string]any {
	return s.usageAttributes
}

type usageWorker struct {
	boottime time.Time
	logger   *slog.Logger
	path     string
	clktck   int64
}

func (w *usageWorker) Interval() time.Duration { return usageUpdateInterval }

func (w *usageWorker) Jitter() time.Duration { return usageUpdateJitter }

func (w *usageWorker) Sensors(_ context.Context, _ time.Duration) ([]sensor.Details, error) {
	return w.getStats()
}

func NewUsageWorker(ctx context.Context) (*linux.SensorWorker, error) {
	clktck, found := linux.CtxGetClkTck(ctx)
	if !found {
		return nil, fmt.Errorf("%w: no clktck value", linux.ErrInvalidCtx)
	}

	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return nil, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx)
	}

	return &linux.SensorWorker{
			Value: &usageWorker{
				clktck:   clktck,
				boottime: boottime,
				logger:   logging.FromContext(ctx).WithGroup(usageWorkerID),
				path:     filepath.Join(linux.ProcFSRoot, "stat"),
			},
			WorkerID: usageWorkerID,
		},
		nil
}

func (w *usageWorker) newUsageSensor(details []string, diagnostic bool) *cpuUsageSensor {
	usageSensor := &cpuUsageSensor{
		cpuID: details[0],
		Sensor: linux.Sensor{
			IconString:      "mdi:chip",
			UnitsString:     "%",
			DataSource:      linux.DataSrcProcfs,
			StateClassValue: types.StateClassMeasurement,
			IsDiagnostic:    diagnostic,
		},
	}
	usageSensor.generateValues(w.clktck, details[1:])

	return usageSensor
}

func (w *usageWorker) newCountSensor(name, icon, valueStr string) *linux.Sensor {
	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		w.logger.Debug("Failed to convert count from string to int.",
			slog.String("sensor", name),
			slog.Any("error", err))
	}

	return &linux.Sensor{
		DisplayName:     name,
		Value:           valueInt,
		IconString:      icon,
		DataSource:      linux.DataSrcProcfs,
		StateClassValue: types.StateClassTotalIncreasing,
		IsDiagnostic:    true,
		LastReset:       w.boottime.Format(time.RFC3339),
	}
}

func (w *usageWorker) getStats() ([]sensor.Details, error) {
	var sensors []sensor.Details

	statsFH, err := os.Open(w.path)
	if err != nil {
		return nil, fmt.Errorf("fetch load averages: %w", err)
	}

	defer statsFH.Close()

	statsFile := bufio.NewScanner(statsFH)
	for statsFile.Scan() {
		// Set up word scanner for line.
		line := bufio.NewScanner(bytes.NewReader(statsFile.Bytes()))
		line.Split(bufio.ScanWords)
		// Split line by words
		var cols []string
		for line.Scan() {
			cols = append(cols, line.Text())
		}
		// Create a sensor depending on the line.
		switch {
		case cols[0] == totalCPUString:
			sensors = append(sensors, w.newUsageSensor(cols, false))
		case strings.Contains(cols[0], "cpu"):
			sensors = append(sensors, w.newUsageSensor(cols, true))
			sensors = append(sensors, newCPUFreqSensor(cols[0]))
		case cols[0] == "ctxt":
			sensors = append(sensors, w.newCountSensor("Total CPU Context Switches", "mdi:counter", cols[1]))
		case cols[0] == "processes":
			sensors = append(sensors, w.newCountSensor("Total Processes Created", "mdi:application-cog", cols[1]))
		case cols[0] == "procs_running":
			sensors = append(sensors, w.newCountSensor("Processes Running", "mdi:application-cog", cols[1]))
		case cols[0] == "procs_blocked":
			sensors = append(sensors, w.newCountSensor("Processes Blocked", "mdi:application-cog", cols[1]))
		}
	}

	return sensors, nil
}
