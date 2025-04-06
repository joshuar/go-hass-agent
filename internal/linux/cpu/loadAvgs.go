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
	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	loadAvgIcon = "mdi:chip"
	loadAvgUnit = "load"

	loadAvgUpdateInterval = time.Minute
	loadAvgUpdateJitter   = 5 * time.Second

	loadAvgsTotal = 3

	loadAvgsWorkerID      = "cpu_loadavg_sensors"
	loadAvgsWorkerDesc    = "Load averages"
	loadAvgsPreferencesID = prefPrefix + "load_averages"
)

var (
	_ quartz.Job                  = (*loadAvgsWorker)(nil)
	_ workers.PollingEntityWorker = (*loadAvgsWorker)(nil)
)

var (
	ErrInitLoadAvgsWorker = errors.New("could not init load averages worker")
	ErrParseLoadAvgs      = errors.New("could not parse load averages")
)

type loadAvgsWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData
	prefs *preferences.CommonWorkerPrefs
	path  string
}

func (w *loadAvgsWorker) Execute(ctx context.Context) error {
	var warnings error

	loadAvgData, err := os.ReadFile(w.path)
	if err != nil {
		return fmt.Errorf("fetch load averages: %w", err)
	}

	loadAvgs, err := parseLoadAvgs(loadAvgData)
	if err != nil {
		return fmt.Errorf("parse load averages: %w", err)
	}

	for name, value := range loadAvgs {
		entity, err := newLoadAvgSensor(ctx, name, value)
		if err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not generate %s sensor: %w", name, err))
			continue
		}
		w.OutCh <- entity
	}

	return warnings
}

func (w *loadAvgsWorker) PreferencesID() string {
	return loadAvgsPreferencesID
}

func (w *loadAvgsWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (w *loadAvgsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *loadAvgsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk usage worker: %w", err)
	}
	return w.OutCh, nil
}

func newLoadAvgSensor(ctx context.Context, name, value string) (models.Entity, error) {
	entity, err := sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithUnits(loadAvgUnit),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithIcon(loadAvgIcon),
		sensor.WithState(value),
		sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
	)
	if err != nil {
		return entity, fmt.Errorf("could not generate %s sensor: %w", name, err)
	}

	return entity, nil
}

func parseLoadAvgs(data []byte) (map[string]string, error) {
	loadAvgsData := bytes.Split(data, []byte(" "))

	if len(loadAvgsData) != 5 { //nolint:mnd
		return nil, ErrParseLoadAvgs
	}

	return map[string]string{
		"CPU load average (1 min)":  string(loadAvgsData[0][:]),
		"CPU load average (5 min)":  string(loadAvgsData[1][:]),
		"CPU load average (15 min)": string(loadAvgsData[2][:]),
	}, nil
}

func NewLoadAvgWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &loadAvgsWorker{
		WorkerMetadata:          models.SetWorkerMetadata(loadAvgsWorkerID, loadAvgsWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		path:                    filepath.Join(linux.ProcFSRoot, "loadavg"),
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitLoadAvgsWorker, err)
	}
	worker.prefs = prefs

	worker.Trigger = scheduler.NewPollTriggerWithJitter(loadAvgUpdateInterval, loadAvgUpdateJitter)

	return worker, nil
}
