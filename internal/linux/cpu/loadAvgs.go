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
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
)

const (
	loadAvgIcon = "mdi:chip"
	loadAvgUnit = "load"

	loadAvgUpdateInterval = time.Minute
	loadAvgUpdateJitter   = 5 * time.Second

	loadAvgsTotal = 3

	loadAvgsWorkerID      = "cpu_loadavg_sensors"
	loadAvgsPreferencesID = prefPrefix + "load_averages"
)

var (
	ErrInitLoadAvgsWorker = errors.New("could not init load averages worker")
	ErrParseLoadAvgs      = errors.New("could not parse load averages")
)

type loadAvgsWorker struct {
	path string
}

func (w *loadAvgsWorker) UpdateDelta(_ time.Duration) {}

func (w *loadAvgsWorker) Sensors(ctx context.Context) ([]models.Entity, error) {
	sensors := make([]models.Entity, 0, loadAvgsTotal)

	loadAvgData, err := os.ReadFile(w.path)
	if err != nil {
		return nil, fmt.Errorf("fetch load averages: %w", err)
	}

	loadAvgs, err := parseLoadAvgs(loadAvgData)
	if err != nil {
		return nil, fmt.Errorf("parse load averages: %w", err)
	}

	for name, value := range loadAvgs {
		entity, err := newLoadAvgSensor(ctx, name, value)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not generate load average sensor.",
				slog.Any("error", err))
			continue
		}
		sensors = append(sensors, entity)
	}

	return sensors, nil
}

func (w *loadAvgsWorker) PreferencesID() string {
	return loadAvgsPreferencesID
}

func (w *loadAvgsWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func newLoadAvgSensor(ctx context.Context, name, value string) (models.Entity, error) {
	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(strcase.ToSnake(name)),
		sensor.WithUnits(loadAvgUnit),
		sensor.WithStateClass(class.StateMeasurement),
		sensor.WithIcon(loadAvgIcon),
		sensor.WithState(value),
		sensor.WithDataSourceAttribute(linux.DataSrcProcfs),
	)
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

func NewLoadAvgWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	loadAvgsWorker := &loadAvgsWorker{path: filepath.Join(linux.ProcFSRoot, "loadavg")}

	prefs, err := preferences.LoadWorker(loadAvgsWorker)
	if err != nil {
		return nil, errors.Join(ErrInitLoadAvgsWorker, err)
	}

	//nolint:nilnil
	if prefs.Disabled {
		return nil, nil
	}

	worker := linux.NewPollingSensorWorker(loadAvgsWorkerID, loadAvgUpdateInterval, loadAvgUpdateJitter)
	worker.PollingSensorType = loadAvgsWorker

	return worker, nil
}
