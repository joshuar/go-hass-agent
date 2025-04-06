// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/reugn/go-quartz/quartz"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/hwmon"
)

const (
	hwMonInterval = time.Minute
	hwMonJitter   = 5 * time.Second

	hwmonWorkerID      = "hwmon_sensors"
	hwmonWorkerDesc    = "hwmon sensor details"
	hwmonPreferencesID = sensorsPrefPrefix + "hardware_sensors"
)

var (
	_ quartz.Job                  = (*hwMonWorker)(nil)
	_ workers.PollingEntityWorker = (*hwMonWorker)(nil)
)

var (
	ErrNewHWMonSensor  = errors.New("could not create hardware monitor sensor")
	ErrInitHWMonWorker = errors.New("could not init hardware sensors worker")
)

func hwmonSensorAttributes(details *hwmon.Sensor) map[string]any {
	attributes := make(map[string]any)

	attributes["sensor_type"] = details.MonitorType.String()
	attributes["sysfs_path"] = details.Path
	attributes["data_source"] = linux.DataSrcSysfs

	if details.Units() != "" {
		attributes["native_unit_of_measurement"] = details.Units()
	}

	return attributes
}

func newHWSensor(ctx context.Context, details *hwmon.Sensor) (*models.Entity, error) {
	var (
		icon             string
		deviceClass      class.SensorDeviceClass
		stateClass       class.SensorStateClass
		sensorTypeOption sensor.Option
	)

	switch details.MonitorType {
	case hwmon.Alarm, hwmon.Intrusion:
		if v, ok := details.Value().(bool); ok && v {
			icon = "mdi:alarm-light"
		} else {
			icon = "mdi:alarm-light-off"
		}

		if details.MonitorType == hwmon.Alarm {
			deviceClass = class.BinaryClassProblem
		} else {
			deviceClass = class.BinaryClassTamper
		}
	default:
		icon, deviceClass = parseSensorType(details.MonitorType.String())
		stateClass = class.StateMeasurement
	}

	if details.MonitorType == hwmon.Alarm || details.MonitorType == hwmon.Intrusion {
		sensorTypeOption = sensor.AsTypeBinarySensor()
	} else {
		sensorTypeOption = sensor.AsTypeSensor()
	}

	hwMonSensor, err := sensor.NewSensor(ctx,
		sensor.WithName(details.Name()),
		sensor.WithID(details.ID()),
		sensor.WithDeviceClass(deviceClass),
		sensor.AsDiagnostic(),
		sensorTypeOption,
		sensor.WithUnits(details.Units()),
		sensor.WithIcon(icon),
		sensor.WithState(details.Value()),
		sensor.WithAttributes(hwmonSensorAttributes(details)),
		sensor.WithStateClass(stateClass),
	)
	if err != nil {
		return nil, errors.Join(ErrNewHWMonSensor, err)
	}

	return &hwMonSensor, nil
}

type hwMonWorker struct {
	prefs *HWMonPrefs
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData
}

func (w *hwMonWorker) Execute(ctx context.Context) error {
	var warnings error

	hwmonSensors, err := hwmon.GetAllSensors()
	if err != nil {
		return fmt.Errorf("could not retrieve hardware sensors: %w", err)
	}

	for hwMonSensor := range slices.Values(hwmonSensors) {
		if entity, err := newHWSensor(ctx, hwMonSensor); err != nil {
			warnings = errors.Join(warnings, fmt.Errorf("could not generate hwmon sensor: %w", err))
		} else {
			w.OutCh <- *entity
		}
	}

	return nil
}

func (w *hwMonWorker) PreferencesID() string {
	return hwmonPreferencesID
}

func (w *hwMonWorker) DefaultPreferences() HWMonPrefs {
	return HWMonPrefs{
		UpdateInterval: hwMonInterval.String(),
	}
}

func (w *hwMonWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *hwMonWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk IO worker: %w", err)
	}
	return w.OutCh, nil
}

func NewHWMonWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &hwMonWorker{
		WorkerMetadata:          models.SetWorkerMetadata(hwmonWorkerID, hwmonWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitHWMonWorker, err)
	}
	worker.prefs = prefs

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", hwmonWorkerID),
			slog.String("given_interval", worker.prefs.UpdateInterval),
			slog.String("default_interval", hwMonInterval.String()))

		pollInterval = hwMonInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, hwMonJitter)

	return worker, nil
}

func parseSensorType(t string) (icon string, deviceclass class.SensorDeviceClass) {
	switch t {
	case "Temp":
		return "mdi:thermometer", class.SensorClassTemperature
	case "Fan":
		return "mdi:turbine", 0
	case "Power":
		return "mdi:flash", class.SensorClassPower
	case "Voltage":
		return "mdi:lightning-bolt", class.SensorClassVoltage
	case "Energy":
		return "mdi:lightning-bolt", class.SensorClassEnergyStorage
	case "Current":
		return "mdi:current-ac", class.SensorClassCurrent
	case "Frequency", "PWM":
		return "mdi:sawtooth-wave", class.SensorClassFrequency
	case "Humidity":
		return "mdi:water-percent", class.SensorClassHumidity
	default:
		return "mdi:chip", 0
	}
}
