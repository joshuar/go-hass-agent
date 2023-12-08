// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/host"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type tempSensor struct {
	linuxSensor
	idx  int
	id   string
	high float64
	crit float64
}

func (s *tempSensor) Name() string {
	c := cases.Title(language.AmericanEnglish)
	return c.String(strcase.ToDelimited(s.id+"_temp", ' '))
}

func (s *tempSensor) ID() string {
	return "temp_" + s.id
}

func (s *tempSensor) Attributes() interface{} {
	return struct {
		NativeUnit string  `json:"native_unit_of_measurement"`
		DataSource string  `json:"Data Source"`
		HighThresh float64 `json:"High Temperature Threshold"`
		CritThresh float64 `json:"Critical Temperature Threshold"`
	}{
		NativeUnit: s.units,
		DataSource: srcProcfs,
		HighThresh: s.high,
		CritThresh: s.crit,
	}
}

func TempUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor)
	update := func(_ time.Duration) {
		rawTemps, err := host.SensorsTemperaturesWithContext(ctx)
		sensorMap := make(map[string]*tempSensor, len(rawTemps))
		if err != nil {
			log.Warn().Err(err).Msg("Could not fetch temperatures.")
		}
		for _, temp := range rawTemps {
			newSensor := &tempSensor{}
			newSensor.diagnostic = true
			newSensor.deviceClass = sensor.SensorTemperature
			newSensor.stateClass = sensor.StateMeasurement
			newSensor.units = "°C"
			newSensor.value = temp.Temperature
			newSensor.high = temp.High
			newSensor.crit = temp.Critical
			newSensor.sensorType = deviceTemp
			if existingSensor, ok := sensorMap[temp.SensorKey]; ok {
				existingSensor.idx++
				newSensor.id = fmt.Sprintf("%s_%d", temp.SensorKey, existingSensor.idx)
			} else {
				newSensor.id = temp.SensorKey
			}
			sensorMap[newSensor.id] = newSensor
		}
		for _, v := range sensorMap {
			sensorCh <- v
		}
	}

	go helpers.PollSensors(ctx, update, time.Minute, time.Second*5)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
	}()
	return sensorCh
}
