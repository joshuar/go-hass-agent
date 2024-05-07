// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"context"
	"sync"
	"time"
	"unicode"

	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	diskstats "github.com/joshuar/go-hass-agent/pkg/linux/proc"
)

const (
	diskRateUnits  = "kB/s"
	diskCountUnits = "requests"
)

type diskIOSensor struct {
	stats  map[diskstats.Stat]uint64
	device diskstats.Device
	linux.Sensor
	prev uint64
}

type diskIOSensorAttributes struct {
	DataSource string `json:"Data Source"`
	NativeUnit string `json:"native_unit_of_measurement,omitempty"`
	Model      string `json:"Device Model,omitempty"`
	SysFSPath  string `json:"SysFS Path,omitempty"`
	Sectors    uint64 `json:"Total Sectors,omitempty"`
	Time       uint64 `json:"Total Milliseconds,omitempty"`
}

type sensors struct {
	totalReads  *diskIOSensor
	totalWrites *diskIOSensor
	readRate    *diskIOSensor
	writeRate   *diskIOSensor
}

func (s *diskIOSensor) Name() string {
	r := []rune(s.device.ID)
	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...)) + " " + s.SensorTypeValue.String()
}

func (s *diskIOSensor) ID() string {
	return s.device.ID + "_" + strcase.ToSnake(s.SensorTypeValue.String())
}

func (s *diskIOSensor) Attributes() any {
	// Common attributes for all disk IO sensors
	attrs := &diskIOSensorAttributes{
		DataSource: linux.DataSrcSysfs,
		Model:      s.device.Model,
		SysFSPath:  s.device.SysFSPath,
	}
	switch s.SensorTypeValue {
	case linux.SensorDiskReads:
		attrs.Sectors = s.stats[diskstats.TotalSectorsRead]
		attrs.Time = s.stats[diskstats.TotalTimeReading]
		attrs.NativeUnit = diskCountUnits
		return attrs
	case linux.SensorDiskWrites:
		attrs.Sectors = s.stats[diskstats.TotalSectorsWritten]
		attrs.Time = s.stats[diskstats.TotalTimeWriting]
		attrs.NativeUnit = diskCountUnits
		return attrs
	case linux.SensorDiskReadRate, linux.SensorDiskWriteRate:
		attrs.NativeUnit = diskRateUnits
		return attrs
	}
	return nil
}

func (s *diskIOSensor) Icon() string {
	switch s.SensorTypeValue {
	case linux.SensorDiskReads, linux.SensorDiskReadRate:
		return "mdi:file-upload"
	case linux.SensorDiskWrites, linux.SensorDiskWriteRate:
		return "mdi:file-download"
	}
	return "mdi:file"
}

func (s *diskIOSensor) update(stats map[diskstats.Stat]uint64, delta time.Duration) {
	s.stats = stats
	var curr uint64
	switch s.SensorTypeValue {
	case linux.SensorDiskReads:
		s.Value = s.stats[diskstats.TotalReads]
	case linux.SensorDiskWrites:
		s.Value = s.stats[diskstats.TotalWrites]
	case linux.SensorDiskReadRate:
		curr = s.stats[diskstats.TotalSectorsRead]
	case linux.SensorDiskWriteRate:
		curr = s.stats[diskstats.TotalSectorsWritten]
	}
	if s.SensorTypeValue == linux.SensorDiskReadRate || s.SensorTypeValue == linux.SensorDiskWriteRate {
		if uint64(delta.Seconds()) > 0 {
			log.Trace().Msgf("%s IO rate calc: (%d - %d) / uint64(%d) / 2", s.device, curr, s.prev, uint64(delta.Seconds()))
			s.Value = (curr - s.prev) / uint64(delta.Seconds()) / 2
		}
		s.prev = curr
	}
}

func newDiskIOSensor(device diskstats.Device, sensorType linux.SensorTypeValue) *diskIOSensor {
	s := &diskIOSensor{
		device: device,
		Sensor: linux.Sensor{
			StateClassValue: types.StateClassTotalIncreasing,
			SensorTypeValue: sensorType,
			UnitsString:     diskCountUnits,
		},
	}
	if device.ID != "total" {
		s.IsDiagnostic = true
	}
	return s
}

func newDiskIORateSensor(device diskstats.Device, sensorType linux.SensorTypeValue) *diskIOSensor {
	s := &diskIOSensor{
		device: device,
		Sensor: linux.Sensor{
			DeviceClassValue: types.DeviceClassDataRate,
			StateClassValue:  types.StateClassMeasurement,
			UnitsString:      diskRateUnits,
			SensorTypeValue:  sensorType,
		},
	}
	if device.ID != "total" {
		s.IsDiagnostic = true
	}
	return s
}

func newDevice(dev diskstats.Device) *sensors {
	return &sensors{
		totalReads:  newDiskIOSensor(dev, linux.SensorDiskReads),
		totalWrites: newDiskIOSensor(dev, linux.SensorDiskWrites),
		readRate:    newDiskIORateSensor(dev, linux.SensorDiskReadRate),
		writeRate:   newDiskIORateSensor(dev, linux.SensorDiskWriteRate),
	}
}

func IOUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	newStats, err := diskstats.ReadDiskStatsFromSysFS()
	if err != nil {
		log.Warn().Err(err).Msg("Error reading disk stats from procfs. Will not send disk rate sensors.")
		close(sensorCh)
		return sensorCh
	}
	devices := make(map[diskstats.Device]*sensors)
	var mu sync.Mutex
	for dev := range newStats {
		devices[dev] = newDevice(dev)
	}
	diskIOstats := func(delta time.Duration) {
		newStats, err := diskstats.ReadDiskStatsFromSysFS()
		if err != nil {
			log.Warn().Err(err).Msgf("Error reading disk stats from %s.", linux.DataSrcSysfs)
		}
		for dev, stats := range newStats {
			// add any new devices
			if _, ok := devices[dev]; !ok {
				mu.Lock()
				devices[dev] = newDevice(dev)
				mu.Unlock()
			}
			mu.Lock()
			// update the stats for this device
			devices[dev].totalReads.update(stats, delta)
			devices[dev].totalWrites.update(stats, delta)
			devices[dev].readRate.update(stats, delta)
			devices[dev].writeRate.update(stats, delta)
			mu.Unlock()
			// send the stats to Home Assistant
			go func(d diskstats.Device) {
				mu.Lock()
				defer mu.Unlock()
				sensorCh <- devices[d].totalReads
			}(dev)
			go func(d diskstats.Device) {
				mu.Lock()
				defer mu.Unlock()
				sensorCh <- devices[d].totalWrites
			}(dev)
			go func(d diskstats.Device) {
				mu.Lock()
				defer mu.Unlock()
				sensorCh <- devices[d].readRate
			}(dev)
			go func(d diskstats.Device) {
				mu.Lock()
				defer mu.Unlock()
				sensorCh <- devices[d].writeRate
			}(dev)
		}
	}
	go helpers.PollSensors(ctx, diskIOstats, 5*time.Second, time.Second*1)
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped disk IO sensors.")
	}()
	return sensorCh
}
