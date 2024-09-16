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

	hwmonWorkerID = "hwmon_sensors"
)

type hwSensor struct {
	*hwmon.Sensor
	icon        func() string
	sensorType  types.SensorClass
	deviceClass types.DeviceClass
	stateClass  types.StateClass
}

func (s *hwSensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	attributes["sensor_type"] = s.MonitorType.String()
	attributes["sysfs_path"] = s.Path
	attributes["data_source"] = linux.DataSrcSysfs

	if s.Units() != "" {
		attributes["native_unit_of_measurement"] = s.Units()
	}

	for _, a := range s.Sensor.Attributes {
		attributes[a.Name] = a.Value
	}

	return attributes
}

func (s *hwSensor) State() any {
	return s.Value()
}

func (s *hwSensor) Icon() string {
	return s.icon()
}

func (s *hwSensor) SensorType() types.SensorClass {
	return s.sensorType
}

func (s *hwSensor) DeviceClass() types.DeviceClass {
	return s.deviceClass
}

func (s *hwSensor) StateClass() types.StateClass {
	return s.stateClass
}

func (s *hwSensor) Category() string {
	return "diagnostic"
}

func newHWSensor(details *hwmon.Sensor) *hwSensor {
	newSensor := &hwSensor{
		Sensor: details,
	}

	switch newSensor.MonitorType {
	case hwmon.Alarm, hwmon.Intrusion:
		newSensor.icon = func() string {
			if v, ok := newSensor.Value().(bool); ok && v {
				return "mdi:alarm-light"
			}

			return "mdi:alarm-light-off"
		}
		newSensor.sensorType = types.BinarySensor
	default:
		icon, deviceClass := parseSensorType(details.MonitorType.String())
		newSensor.icon = func() string { return icon }
		newSensor.deviceClass = deviceClass
		newSensor.stateClass = types.StateClassMeasurement
	}

	return newSensor
}

type hwMonWorker struct{}

func (w *hwMonWorker) UpdateDelta(_ time.Duration) {}

func (w *hwMonWorker) Sensors(_ context.Context) ([]sensor.Details, error) {
	hwmonSensors, err := hwmon.GetAllSensors()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve hardware sensors: %w", err)
	}

	sensors := make([]sensor.Details, 0, len(hwmonSensors))

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
