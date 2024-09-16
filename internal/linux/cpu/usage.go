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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	usageUpdateInterval = 10 * time.Second
	usageUpdateJitter   = 500 * time.Millisecond

	usageWorkerID = "cpu_usage_sensors"

	totalCPUString = "cpu"
)

var ErrParseCPUUsage = errors.New("could not parse CPU usage")

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
	path     string
	linux.PollingSensorWorker
	clktck int64
}

func (w *usageWorker) UpdateDelta(_ time.Duration) {}

func (w *usageWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	return w.getStats()
}

func NewUsageWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingWorker(usageWorkerID, usageUpdateInterval, usageUpdateJitter)

	clktck, found := linux.CtxGetClkTck(ctx)
	if !found {
		return worker, fmt.Errorf("%w: no clktck value", linux.ErrInvalidCtx)
	}

	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return worker, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx)
	}

	worker.PollingType = &usageWorker{
		path:     filepath.Join(linux.ProcFSRoot, "stat"),
		boottime: boottime,
		clktck:   clktck,
	}

	return worker, nil
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
	valueInt, _ := strconv.Atoi(valueStr) //nolint:errcheck // if we can't parse it, value will be 0.

	return &linux.Sensor{
		DisplayName:     name,
		UniqueID:        strcase.ToSnake(name),
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
		return nil, fmt.Errorf("fetch cpu usage: %w", err)
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

		if len(cols) == 0 {
			return sensors, ErrParseCPUUsage
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
