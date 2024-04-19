// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"context"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/device/helpers"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	diskstats "github.com/joshuar/go-hass-agent/pkg/linux/proc"
)

type diskIOSensor struct {
	stats  map[diskstats.DiskStat]uint64
	device string
	linux.Sensor
	prev uint64
}

type diskIOSensorAttributes struct {
	DataSource string `json:"Data Source"`
	NativeUnit string `json:"native_unit_of_measurement,omitempty"`
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
	return s.device + " " + s.SensorTypeValue.String()
}

func (s *diskIOSensor) ID() string {
	return s.device + "_" + strcase.ToSnake(s.SensorTypeValue.String())
}

func (s *diskIOSensor) Attributes() any {
	switch s.SensorTypeValue {
	case linux.SensorDiskReads:
		return &diskIOSensorAttributes{
			DataSource: linux.DataSrcProcfs,
			Sectors:    s.stats[diskstats.TotalSectorsRead],
			Time:       s.stats[diskstats.TotalTimeReading],
		}
	case linux.SensorDiskWrites:
		return &diskIOSensorAttributes{
			DataSource: linux.DataSrcProcfs,
			Sectors:    s.stats[diskstats.TotalSectorsWritten],
			Time:       s.stats[diskstats.TotalTimeWriting],
		}
	case linux.SensorDiskReadRate:
		return &diskIOSensorAttributes{
			DataSource: linux.DataSrcProcfs,
			NativeUnit: "KB/s",
		}
	case linux.SensorDiskWriteRate:
		return &diskIOSensorAttributes{
			DataSource: linux.DataSrcProcfs,
			NativeUnit: "KB/s",
		}
	}
	return nil
}

func (s *diskIOSensor) Icon() string {
	switch s.SensorTypeValue {
	case linux.SensorDiskReads:
		return "mdi:file-upload"
	case linux.SensorDiskWrites:
		return "mdi:file-download"
	}
	return "mdi:file"
}

func (s *diskIOSensor) update(stats map[diskstats.DiskStat]uint64, delta time.Duration) {
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

func newDiskIOSensor(device string, sensorType linux.SensorTypeValue) *diskIOSensor {
	s := &diskIOSensor{
		device: device,
		Sensor: linux.Sensor{
			DeviceClassValue: types.DeviceClassDataSize,
			StateClassValue:  types.StateClassTotalIncreasing,
			SensorTypeValue:  sensorType,
			IsDiagnostic:     true,
		},
	}
	return s
}

func newDiskIORateSensor(device string, sensorType linux.SensorTypeValue) *diskIOSensor {
	s := &diskIOSensor{
		device: device,
		Sensor: linux.Sensor{
			DeviceClassValue: types.DeviceClassDataRate,
			StateClassValue:  types.StateClassMeasurement,
			UnitsString:      "KB/s",
			SensorTypeValue:  sensorType,
		},
	}
	return s
}

func newDevice(dev string) *sensors {
	return &sensors{
		totalReads:  newDiskIOSensor(dev, linux.SensorDiskReads),
		totalWrites: newDiskIOSensor(dev, linux.SensorDiskWrites),
		readRate:    newDiskIORateSensor(dev, linux.SensorDiskReadRate),
		writeRate:   newDiskIORateSensor(dev, linux.SensorDiskWriteRate),
	}
}

func IOUpdater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	newStats, err := diskstats.ReadDiskstats()
	if err != nil {
		log.Warn().Err(err).Msg("Error reading disk stats from procfs. Will not send disk rate sensors.")
		close(sensorCh)
		return sensorCh
	}
	devices := make(map[string]*sensors)
	for dev := range newStats {
		devices[dev] = newDevice(dev)
	}
	diskIOstats := func(delta time.Duration) {
		newStats, err := diskstats.ReadDiskstats()
		if err != nil {
			log.Warn().Err(err).Msg("Error reading disk stats from procfs.")
		}
		for dev, stats := range newStats {
			if _, ok := devices[dev]; !ok {
				devices[dev] = newDevice(dev)
			}
			devices[dev].totalReads.update(stats, delta)
			devices[dev].totalWrites.update(stats, delta)
			devices[dev].readRate.update(stats, delta)
			devices[dev].writeRate.update(stats, delta)
			go func(d string) {
				sensorCh <- devices[d].totalReads
			}(dev)
			go func(d string) {
				sensorCh <- devices[d].totalWrites
			}(dev)
			go func(d string) {
				sensorCh <- devices[d].readRate
			}(dev)
			go func(d string) {
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
