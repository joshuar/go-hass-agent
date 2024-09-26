// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package system

import (
	"context"
	"fmt"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/hwmon"
)

const (
	hwMonInterval = time.Minute
	hwMonJitter   = 5 * time.Second

	hwmonWorkerID = "hwmon"
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
	newSensor := sensor.Entity{
		Name:     details.Name(),
		Category: types.CategoryDiagnostic,
		Units:    details.Units(),
		EntityState: &sensor.EntityState{
			ID:         details.ID(),
			State:      details.Value(),
			Attributes: hwmonSensorAttributes(details),
		},
	}

	switch details.MonitorType {
	case hwmon.Alarm, hwmon.Intrusion:
		if v, ok := details.Value().(bool); ok && v {
			newSensor.Icon = "mdi:alarm-light"
		}

		newSensor.Icon = "mdi:alarm-light-off"
		newSensor.EntityType = types.BinarySensor
	default:
		icon, deviceClass := parseSensorType(details.MonitorType.String())
		newSensor.Icon = icon
		newSensor.DeviceClass = deviceClass
		newSensor.StateClass = types.StateClassMeasurement
	}

	return newSensor
}

type hwMonWorker struct{}

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

func NewHWMonWorker(_ context.Context) (*linux.PollingSensorWorker, error) {
	worker := linux.NewPollingWorker(hwmonWorkerID, hwMonInterval, hwMonJitter)
	worker.PollingType = &hwMonWorker{}

	return worker, nil
}

func parseSensorType(t string) (icon string, deviceclass types.DeviceClass) {
	switch t {
	case "Temp":
		return "mdi:thermometer", types.DeviceClassTemperature
	case "Fan":
		return "mdi:turbine", 0
	case "Power":
		return "mdi:flash", types.DeviceClassPower
	case "Voltage":
		return "mdi:lightning-bolt", types.DeviceClassVoltage
	case "Energy":
		return "mdi:lightning-bolt", types.DeviceClassEnergyStorage
	case "Current":
		return "mdi:current-ac", types.DeviceClassCurrent
	case "Frequency", "PWM":
		return "mdi:sawtooth-wave", types.DeviceClassFrequency
	case "Humidity":
		return "mdi:water-percent", types.DeviceClassHumidity
	default:
		return "mdi:chip", 0
	}
}
