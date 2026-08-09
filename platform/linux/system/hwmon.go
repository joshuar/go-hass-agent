// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/pkg/linux/hwmon"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	hwMonInterval = time.Minute
	hwMonJitter   = 5 * time.Second
)

var (
	_ quartz.Job                  = (*hwMonWorker)(nil)
	_ workers.PollingEntityWorker = (*hwMonWorker)(nil)
)

func hwmonSensorAttributes(details *hwmon.Sensor) map[string]any {
	attributes := make(map[string]any)

	attributes["sensor_type"] = details.MonitorType.String()
	attributes["sysfs_path"] = details.Path
	attributes["data_source"] = linux.DataSrcSysFS

	if details.Units() != "" {
		attributes["native_unit_of_measurement"] = details.Units()
	}

	return attributes
}

func newHWSensor(ctx context.Context, details *hwmon.Sensor) models.Entity {
	var (
		icon             string
		deviceClass      models.SensorDeviceClass
		stateClass       models.SensorStateClass
		sensorTypeOption models.SensorOption
	)

	switch details.MonitorType {
	case hwmon.Alarm, hwmon.Intrusion:
		if v, ok := details.Value().(bool); ok && v {
			icon = "mdi:alarm-light"
		} else {
			icon = "mdi:alarm-light-off"
		}

		if details.MonitorType == hwmon.Alarm {
			deviceClass = models.BinaryClassProblem
		} else {
			deviceClass = models.BinaryClassTamper
		}
	default:
		icon, deviceClass = parseSensorType(details.MonitorType.String())
		stateClass = models.StateMeasurement
	}

	if details.MonitorType == hwmon.Alarm || details.MonitorType == hwmon.Intrusion {
		sensorTypeOption = models.AsTypeBinarySensor()
	} else {
		sensorTypeOption = models.AsTypeSensor()
	}

	return models.NewSensor(ctx,
		models.WithName(details.Name()),
		models.WithID(details.ID()),
		models.WithDeviceClass(deviceClass),
		models.AsDiagnostic(),
		sensorTypeOption,
		models.WithUnits(details.Units()),
		models.WithIcon(icon),
		models.WithState(details.Value()),
		models.WithAttributes(hwmonSensorAttributes(details)),
		models.WithStateClass(stateClass),
	)
}

type hwMonWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *HWMonPrefs
}

func NewHWMonWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &hwMonWorker{
		WorkerMetadata:          models.SetWorkerMetadata("hwmon", "Hardware sensor monitoring"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &HWMonPrefs{
		UpdateInterval: hwMonInterval.String(),
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(sensorsPrefPrefix+"hardware_sensors", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = hwMonInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, hwMonJitter)

	return worker, nil
}

func (w *hwMonWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk IO worker: %w", err)
	}
	return w.OutCh, nil
}

func (w *hwMonWorker) Execute(ctx context.Context) error {
	ctx = slogctx.With(ctx, "worker", w.ID())

	hwmonSensors, err := hwmon.GetAllSensors(ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve hardware sensors: %w", err)
	}

	for hwMonSensor := range slices.Values(hwmonSensors) {
		w.OutCh <- newHWSensor(ctx, hwMonSensor)
	}

	return nil
}

func (w *hwMonWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func parseSensorType(t string) (string, models.SensorDeviceClass) {
	switch t {
	case "Temp":
		return "mdi:thermometer", models.SensorClassTemperature
	case "Fan":
		return "mdi:turbine", 0
	case "Power":
		return "mdi:flash", models.SensorClassPower
	case "Voltage":
		return "mdi:lightning-bolt", models.SensorClassVoltage
	case "Energy":
		return "mdi:lightning-bolt", models.SensorClassEnergyStorage
	case "Current":
		return "mdi:current-ac", models.SensorClassCurrent
	case "Frequency", "PWM":
		return "mdi:sawtooth-wave", models.SensorClassFrequency
	case "Humidity":
		return "mdi:water-percent", models.SensorClassHumidity
	default:
		return "mdi:chip", 0
	}
}
