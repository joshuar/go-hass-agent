// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package disk

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unicode"

	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"

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

// ioWorker creates sensors for disk IO counts and rates per device. It
// maintains an internal map of devices being tracked.
type ioWorker struct {
	devices map[diskstats.Device]*sensors
	mu      sync.Mutex
}

// addDevice adds a new device to the tracker map. If sthe device is already
// being tracked, it will not be added again. The bool return indicates whether
// a device was added (true) or not (false).
func (w *ioWorker) addDevice(dev diskstats.Device) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.devices[dev]; !ok {
		w.devices[dev] = newDevice(dev)
	}
}

// updateDevice will update a tracked device's stats. For rates, it will
// recalculate based on the given time delta.
func (w *ioWorker) updateDevice(dev diskstats.Device, stats map[diskstats.Stat]uint64, delta time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.devices[dev].totalReads.update(stats, delta)
	w.devices[dev].totalWrites.update(stats, delta)
	w.devices[dev].readRate.update(stats, delta)
	w.devices[dev].writeRate.update(stats, delta)
}

// deviceSensors returns the device stats as sensors.
func (w *ioWorker) deviceSensors(dev diskstats.Device) []sensor.Details {
	w.mu.Lock()
	defer w.mu.Unlock()
	return []sensor.Details{
		w.devices[dev].totalReads,
		w.devices[dev].totalWrites,
		w.devices[dev].readRate,
		w.devices[dev].writeRate,
	}
}

func (w *ioWorker) Interval() time.Duration { return 5 * time.Second }

func (w *ioWorker) Jitter() time.Duration { return time.Second }

func (w *ioWorker) Sensors(_ context.Context, d time.Duration) ([]sensor.Details, error) {
	var sensors []sensor.Details
	newStats, err := diskstats.ReadDiskStatsFromSysFS()
	if err != nil {
		return nil, fmt.Errorf("error reading disk stats from %s: %w", linux.DataSrcSysfs, err)
	}
	for dev, stats := range newStats {
		// Add device (if it isn't already tracked).
		w.addDevice(dev)
		// Update the stats.
		w.updateDevice(dev, stats, d)
		// Append its sensors.
		sensors = append(sensors, w.deviceSensors(dev)...)
	}
	return sensors, nil
}

func NewIOWorker() (*linux.SensorWorker, error) {
	worker := &ioWorker{
		devices: make(map[diskstats.Device]*sensors),
	}

	newStats, err := diskstats.ReadDiskStatsFromSysFS()
	if err != nil {
		log.Warn().Err(err).Msg("Error reading disk stats from procfs. Will not send disk rate sensors.")
	}
	for dev := range newStats {
		worker.addDevice(dev)
	}

	return &linux.SensorWorker{
			WorkerName: "Disk IO Sensors",
			WorkerDesc: "Disk IO Counts and Rates.",
			Value:      worker,
		},
		nil
}
