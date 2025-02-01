// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/hwmon"
)

const (
	hwMonInterval = time.Minute
	hwMonJitter   = 5 * time.Second

	hwmonWorkerID      = "hwmon_sensors"
	hwmonPreferencesID = sensorsPrefPrefix + "hardware_sensors"
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

func newHWSensor(details *hwmon.Sensor) sensor.Entity {
	var (
		icon        string
		deviceClass types.DeviceClass
		stateClass  types.StateClass
	)

	switch details.MonitorType {
	case hwmon.Alarm, hwmon.Intrusion:
		if v, ok := details.Value().(bool); ok && v {
			icon = "mdi:alarm-light"
		} else {
			icon = "mdi:alarm-light-off"
		}

		if details.MonitorType == hwmon.Alarm {
			deviceClass = types.BinarySensorDeviceClassProblem
		} else {
			deviceClass = types.BinarySensorDeviceClassTamper
		}
	default:
		icon, deviceClass = parseSensorType(details.MonitorType.String())
		stateClass = types.StateClassMeasurement
	}

	hwMonSensor := sensor.NewSensor(
		sensor.WithName(details.Name()),
		sensor.WithID(details.ID()),
		sensor.WithDeviceClass(deviceClass),
		sensor.AsDiagnostic(),
		sensor.WithUnits(details.Units()),
		sensor.WithState(
			sensor.WithIcon(icon),
			sensor.WithValue(details.Value()),
			sensor.WithAttributes(hwmonSensorAttributes(details)),
		),
	)

	if stateClass != types.StateClassNone {
		hwMonSensor = sensor.WithStateClass(stateClass)(hwMonSensor)
	}

	if details.MonitorType == hwmon.Alarm || details.MonitorType == hwmon.Intrusion {
		hwMonSensor.EntityType = types.BinarySensor
	}

	return hwMonSensor
}

type hwMonWorker struct {
	prefs *HWMonPrefs
}

func (w *hwMonWorker) UpdateDelta(_ time.Duration) {}

func (w *hwMonWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	hwmonSensors, err := hwmon.GetAllSensors()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve hardware sensors: %w", err)
	}

	sensors := make([]sensor.Entity, 0, len(hwmonSensors))

	for _, s := range hwmonSensors {
		sensors = append(sensors, newHWSensor(s))
	}

	return sensors, nil
}

func (w *hwMonWorker) PreferencesID() string {
	return hwmonPreferencesID
}

func (w *hwMonWorker) DefaultPreferences() HWMonPrefs {
	return HWMonPrefs{
		UpdateInterval: hwMonInterval.String(),
	}
}

func NewHWMonWorker(ctx context.Context) (*linux.PollingSensorWorker, error) {
	var err error

	hwMonWorker := &hwMonWorker{}

	hwMonWorker.prefs, err = preferences.LoadWorker(hwMonWorker)
	if err != nil {
		return nil, fmt.Errorf("could not load preferences: %w", err)
	}

	//nolint:nilnil
	if hwMonWorker.prefs.IsDisabled() {
		return nil, nil
	}

	pollInterval, err := time.ParseDuration(hwMonWorker.prefs.UpdateInterval)
	if err != nil {
		logging.FromContext(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", hwmonWorkerID),
			slog.String("given_interval", hwMonWorker.prefs.UpdateInterval),
			slog.String("default_interval", hwMonInterval.String()))

		pollInterval = hwMonInterval
	}

	worker := linux.NewPollingSensorWorker(hwmonWorkerID, pollInterval, hwMonJitter)
	worker.PollingSensorType = hwMonWorker

	return worker, nil
}

func parseSensorType(t string) (icon string, deviceclass types.DeviceClass) {
	switch t {
	case "Temp":
		return "mdi:thermometer", types.SensorDeviceClassTemperature
	case "Fan":
		return "mdi:turbine", 0
	case "Power":
		return "mdi:flash", types.SensorDeviceClassPower
	case "Voltage":
		return "mdi:lightning-bolt", types.SensorDeviceClassVoltage
	case "Energy":
		return "mdi:lightning-bolt", types.SensorDeviceClassEnergyStorage
	case "Current":
		return "mdi:current-ac", types.SensorDeviceClassCurrent
	case "Frequency", "PWM":
		return "mdi:sawtooth-wave", types.SensorDeviceClassFrequency
	case "Humidity":
		return "mdi:water-percent", types.SensorDeviceClassHumidity
	default:
		return "mdi:chip", 0
	}
}
