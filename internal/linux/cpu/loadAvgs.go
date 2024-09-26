// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package cpu

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	loadAvgIcon = "mdi:chip"
	loadAvgUnit = "load"

	loadAvgUpdateInterval = time.Minute
	loadAvgUpdateJitter   = 5 * time.Second

	loadAvgsTotal = 3

	loadAvgsWorkerID = "load_averages_sensors"
)

var ErrParseLoadAvgs = errors.New("could not parse load averages")

type loadAvgsWorker struct {
	path     string
	loadAvgs []sensor.Entity
}

func (w *loadAvgsWorker) UpdateDelta(_ time.Duration) {}

func (w *loadAvgsWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	sensors := make([]sensor.Entity, loadAvgsTotal)

	loadAvgData, err := os.ReadFile(w.path)
	if err != nil {
		return nil, fmt.Errorf("fetch load averages: %w", err)
	}

	loadAvgs, err := parseLoadAvgs(loadAvgData)
	if err != nil {
		return nil, fmt.Errorf("parse load averages: %w", err)
	}

	for idx := range loadAvgs {
		w.loadAvgs[idx].State = loadAvgs[idx]
		sensors[idx] = w.loadAvgs[idx]
	}

	return sensors, nil
}

func newLoadAvgSensors() []sensor.Entity {
	sensors := make([]sensor.Entity, loadAvgsTotal)

	for idx, loadType := range []string{"CPU load average (1 min)", "CPU load average (5 min)", "CPU load average (15 min)"} {
		loadAvgSensor := sensor.Entity{
			Name:       loadType,
			Units:      loadAvgUnit,
			StateClass: types.StateClassMeasurement,
			EntityState: &sensor.EntityState{
				ID:   strcase.ToSnake(loadType),
				Icon: loadAvgIcon,
				Attributes: map[string]any{
					"data_source": linux.DataSrcProcfs,
				},
			},
		}

		sensors[idx] = loadAvgSensor
	}

	return sensors
}

func parseLoadAvgs(data []byte) ([]string, error) {
	loadAvgsData := bytes.Split(data, []byte(" "))

	if len(loadAvgsData) != 5 { //nolint:mnd
		return nil, ErrParseLoadAvgs
	}

	loadAvgs := make([]string, loadAvgsTotal)

	for idx := range loadAvgs {
		loadAvgs[idx] = string(loadAvgsData[idx][:])
	}

	return loadAvgs, nil
}

func NewLoadAvgWorker(_ context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingWorker(loadAvgsWorkerID, loadAvgUpdateInterval, loadAvgUpdateJitter)
	worker.PollingType = &loadAvgsWorker{loadAvgs: newLoadAvgSensors(), path: filepath.Join(linux.ProcFSRoot, "loadavg")}

	return worker, nil
}
