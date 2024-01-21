// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system

import (
	"context"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/linux/hwmon"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type hwSensor struct {
	ExtraAttrs map[string]float64
	hwType     string
	name       string
	linux.Sensor
}

func (s *hwSensor) Name() string {
	c := cases.Title(language.AmericanEnglish)
	return c.String(strcase.ToDelimited(s.name+"_"+s.hwType, ' '))
}

func (s *hwSensor) ID() string {
	return strcase.ToSnake(s.hwType + "_" + s.name)
}

func (s *hwSensor) Attributes() any {
	return struct {
		Attributes map[string]float64 `json:"Extra Attributes,omitempty"`
		NativeUnit string             `json:"native_unit_of_measurement"`
		DataSource string             `json:"Data Source"`
		SensorType string             `json:"Sensor Type"`
	}{
		NativeUnit: s.UnitsString,
		DataSource: linux.DataSrcSysfs,
		SensorType: s.hwType,
		Attributes: s.ExtraAttrs,
	}
}

func newHWSensor(s *hwmon.Sensor) *hwSensor {
	hw := &hwSensor{
		name:       s.Name(),
		hwType:     s.SensorType.String(),
		ExtraAttrs: make(map[string]float64),
	}
	hw.Value = s.Value()
	hw.UnitsString = s.Units()
	i, d := parseSensorType(s.SensorType.String())
	hw.IconString = i
	hw.DeviceClassValue = d
	hw.StateClassValue = sensor.StateMeasurement
	hw.IsDiagnostic = true
	for _, a := range s.Attributes {
		hw.ExtraAttrs[a.Name] = a.Value
	}
	return hw
}

func HWSensorUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)
	update := func(_ time.Duration) {
		allSensors := hwmon.GetAllSensors()
		for _, s := range allSensors {
			sensor := newHWSensor(&s)
			sensorCh <- sensor
		}
	}

	go helpers.PollSensors(ctx, update, time.Minute, time.Second*5)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped temp sensors.")
	}()
	return sensorCh
}

func parseSensorType(t string) (icon string, deviceclass sensor.SensorDeviceClass) {
	switch t {
	case "Temp":
		return "mdi:thermometer", sensor.SensorTemperature
	case "Fan":
		return "mdi:turbine", 0
	case "Power":
		return "mdi:zap", sensor.SensorPower
	case "Voltage":
		return "mdi:lightning-bolt", sensor.Voltage
	case "Energy":
		return "mdi:lightning-bolt", sensor.Energy
	case "Current":
		return "mdi:current-ac", sensor.Current
	case "Frequency", "PWM":
		return "mdi:sawtooth-wave", sensor.Frequency
	case "Humidity":
		return "mdi:water-percent", sensor.Humidity
	default:
		return "mdi:chip", 0
	}
}
