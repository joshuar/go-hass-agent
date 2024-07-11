// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:exhaustruct
//revive:disable:unused-receiver
package system

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/pkg/linux/hwmon"
)

const (
	hwMonInterval = time.Minute
	hwMonJitter   = 5 * time.Second

	hwmonWorkerID = "hwmon_sensors"
)

type hwSensor struct {
	ExtraAttrs map[string]float64
	hwType     string
	name       string
	id         string
	path       string
	linux.Sensor
}

//nolint:errcheck
func (s *hwSensor) asBool(details *hwmon.Sensor) {
	s.Value, _ = details.Value()
	if v, ok := s.Value.(bool); ok && v {
		s.IconString = "mdi:alarm-light"
	} else {
		s.IconString = "mdi:alarm-light-off"
	}

	s.IsBinary = true
}

//nolint:errcheck
func (s *hwSensor) asFloat(details *hwmon.Sensor) {
	s.Value, _ = details.Value()
	s.UnitsString = details.Units()
	i, d := parseSensorType(details.SensorType.String())
	s.IconString = i
	s.DeviceClassValue = d
	s.StateClassValue = types.StateClassMeasurement

	for _, a := range details.Attributes {
		s.ExtraAttrs[a.Name] = a.Value
	}
}

func (s *hwSensor) Name() string {
	return s.name
}

func (s *hwSensor) ID() string {
	return s.id
}

func (s *hwSensor) Attributes() map[string]any {
	attributes := make(map[string]any)
	if s.ExtraAttrs != nil {
		attributes["extra_attributes"] = s.ExtraAttrs
	}

	if s.UnitsString != "" {
		attributes["native_unit_of_measurement"] = s.UnitsString
	}

	attributes["data_source"] = linux.DataSrcSysfs
	attributes["sensor_type"] = s.hwType
	attributes["sysfs_path"] = s.path

	return attributes
}

//nolint:exhaustruct
func newHWSensor(details *hwmon.Sensor) *hwSensor {
	newSensor := &hwSensor{
		name:       details.Name(),
		id:         details.ID(),
		hwType:     details.SensorType.String(),
		path:       details.SysFSPath,
		ExtraAttrs: make(map[string]float64),
	}
	newSensor.IsDiagnostic = true

	switch newSensor.hwType {
	case hwmon.Alarm.String(), hwmon.Intrusion.String():
		newSensor.asBool(details)
	default:
		newSensor.asFloat(details)
	}

	return newSensor
}

type hwMonWorker struct {
	logger *slog.Logger
}

func (w *hwMonWorker) Interval() time.Duration { return hwMonInterval }

func (w *hwMonWorker) Jitter() time.Duration { return hwMonJitter }

func (w *hwMonWorker) Sensors(_ context.Context, _ time.Duration) ([]sensor.Details, error) {
	hwmonSensors, err := hwmon.GetAllSensors()
	sensors := make([]sensor.Details, 0, len(hwmonSensors))

	if err != nil && len(hwmonSensors) > 0 {
		w.logger.Warn("Errors fetching some chip/sensor values from hwmon API.", "error", err.Error())
	}

	if err != nil && len(hwmonSensors) == 0 {
		return nil, fmt.Errorf("could not retrieve hwmon sensor details: %w", err)
	}

	for _, s := range hwmonSensors {
		sensors = append(sensors, newHWSensor(s))
	}

	return sensors, nil
}

func NewHWMonWorker(ctx context.Context, _ *dbusx.DBusAPI) (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			Value: &hwMonWorker{
				logger: logging.FromContext(ctx).With(slog.String("worker", hwmonWorkerID)),
			},
			WorkerID: hwmonWorkerID,
		},
		nil
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
