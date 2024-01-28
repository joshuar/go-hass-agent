// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package system

import (
	"context"
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
)

type tempSensor struct {
	high float64
	crit float64
	id   string
	linux.Sensor
}

func (s *tempSensor) Name() string {
	c := cases.Title(language.AmericanEnglish)
	return c.String(strcase.ToDelimited(s.id+"_temp", ' '))
}

func (s *tempSensor) ID() string {
	return "temp_" + s.id
}

func (s *tempSensor) Attributes() any {
	return struct {
		NativeUnit string  `json:"native_unit_of_measurement"`
		DataSource string  `json:"Data Source"`
		HighThresh float64 `json:"High Temperature Threshold"`
		CritThresh float64 `json:"Critical Temperature Threshold"`
	}{
		NativeUnit: s.UnitsString,
		DataSource: linux.DataSrcProcfs,
		HighThresh: s.high,
		CritThresh: s.crit,
	}
}

func newTempSensor(t host.TemperatureStat) *tempSensor {
	s := &tempSensor{}
	s.IsDiagnostic = true
	s.DeviceClassValue = sensor.SensorTemperature
	s.StateClassValue = sensor.StateMeasurement
	s.UnitsString = "°C"
	s.SensorTypeValue = linux.SensorDeviceTemp
	s.Value = t.Temperature
	s.high = t.High
	s.crit = t.Critical
	return s
}

func TempUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)
	update := func(_ time.Duration) {
		rawTemps, err := host.SensorsTemperaturesWithContext(ctx)
		idCounter := make(map[string]int)
		if err != nil {
			log.Warn().Err(err).Msg("Could not fetch some temperatures.")
		}
		for _, temp := range rawTemps {
			s := newTempSensor(temp)
			if _, ok := idCounter[temp.SensorKey]; ok {
				idCounter[s.id]++
				s.id = fmt.Sprintf("%s_%d", temp.SensorKey, idCounter[s.id])
			} else {
				s.id = temp.SensorKey
				idCounter[temp.SensorKey] = 0
			}
			sensorCh <- s
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
