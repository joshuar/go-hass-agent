// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/hwmon"
)

type hwSensor struct {
	ExtraAttrs map[string]float64
	hwType     string
	name       string
	id         string
	path       string
	linux.Sensor
}

func (s *hwSensor) asBool(h *hwmon.Sensor) {
	s.Value = h.Value()
	if v, ok := s.Value.(bool); ok && v {
		s.IconString = "mdi:alarm-light"
	} else {
		s.IconString = "mdi:alarm-light-off"
	}
	s.IsBinary = true
}

func (s *hwSensor) asFloat(h *hwmon.Sensor) {
	s.Value = h.Value()
	s.UnitsString = h.Units()
	i, d := parseSensorType(h.SensorType.String())
	s.IconString = i
	s.DeviceClassValue = d
	s.StateClassValue = types.StateClassMeasurement
	for _, a := range h.Attributes {
		s.ExtraAttrs[a.Name] = a.Value
	}
}

func (s *hwSensor) Name() string {
	return s.name
}

func (s *hwSensor) ID() string {
	return s.id
}

func (s *hwSensor) Attributes() any {
	return struct {
		Attributes map[string]float64 `json:"Extra Attributes,omitempty"`
		NativeUnit string             `json:"native_unit_of_measurement,omitempty"`
		DataSource string             `json:"Data Source"`
		SensorType string             `json:"Sensor Type"`
		HWMonPath  string             `json:"SysFS Path"`
	}{
		NativeUnit: s.UnitsString,
		DataSource: linux.DataSrcSysfs,
		SensorType: s.hwType,
		Attributes: s.ExtraAttrs,
		HWMonPath:  s.path,
	}
}

func newHWSensor(s *hwmon.Sensor) *hwSensor {
	hw := &hwSensor{
		name:       s.Name(),
		id:         s.ID(),
		hwType:     s.SensorType.String(),
		path:       s.SysFSPath,
		ExtraAttrs: make(map[string]float64),
	}
	hw.IsDiagnostic = true
	switch hw.hwType {
	case hwmon.Alarm.String(), hwmon.Intrusion.String():
		hw.asBool(s)
	default:
		hw.asFloat(s)
	}
	return hw
}

type hwMonWorker struct{}

func (w *hwMonWorker) Interval() time.Duration { return time.Minute }

func (w *hwMonWorker) Jitter() time.Duration { return 5 * time.Second }

func (w *hwMonWorker) Sensors(_ context.Context, _ time.Duration) ([]sensor.Details, error) {
	var sensors []sensor.Details
	hwmonSensors, err := hwmon.GetAllSensors()
	if err != nil && len(hwmonSensors) > 0 {
		log.Warn().Err(err).Msg("Errors fetching some chip/sensor values from hwmon API.")
	}
	if err != nil && len(hwmonSensors) == 0 {
		return nil, fmt.Errorf("could not retrieve hwmon sensor details: %w", err)
	}
	for _, s := range hwmonSensors {
		sensors = append(sensors, newHWSensor(s))
	}
	return sensors, nil
}

func NewHWMonWorker() (*linux.SensorWorker, error) {
	return &linux.SensorWorker{
			WorkerName: "HWMon Sensors",
			WorkerDesc: "Sensors from the hwmon kernel interface (chip temps, fan speeds, etc.)",
			Value:      &hwMonWorker{},
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
