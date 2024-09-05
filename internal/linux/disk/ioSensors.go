// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate stringer -type=ioSensor -output ioSensors_generated.go -linecomment
package disk

import (
	"time"
	"unicode"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	diskReads     ioSensor = iota // Disk Reads
	diskWrites                    // Disk Writes
	diskReadRate                  // Disk Read Rate
	diskWriteRate                 // Disk Write Rate

	diskRateUnits  = "kB/s"
	diskCountUnits = "requests"
)

type ioSensor int

type diskIOSensor struct {
	device     *device
	attributes map[string]any
	linux.Sensor
	sensorType ioSensor
	prevValue  uint64
}

func (s *diskIOSensor) Name() string {
	r := []rune(s.device.id)

	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...)) + " " + s.sensorType.String()
}

func (s *diskIOSensor) ID() string {
	return s.device.id + "_" + strcase.ToSnake(s.sensorType.String())
}

func (s *diskIOSensor) Attributes() map[string]any {
	return s.attributes
}

func (s *diskIOSensor) Icon() string {
	switch s.sensorType {
	case diskReads, diskReadRate:
		return "mdi:file-upload"
	case diskWrites, diskWriteRate:
		return "mdi:file-download"
	}

	return "mdi:file"
}

//nolint:mnd
func (s *diskIOSensor) update(stats map[stat]uint64, delta time.Duration) {
	var curr uint64

	switch s.sensorType {
	case diskReads:
		s.Value = stats[TotalReads]
	case diskWrites:
		s.Value = stats[TotalWrites]
	case diskReadRate:
		curr = stats[TotalSectorsRead]
	case diskWriteRate:
		curr = stats[TotalSectorsWritten]
	}

	// For rate sensors, calculate the current value based on previous value and
	// time interval since last measurement.
	if s.sensorType == diskReadRate || s.sensorType == diskWriteRate {
		if uint64(delta.Seconds()) > 0 {
			s.Value = (curr - s.prevValue) / uint64(delta.Seconds()) / 2
		} else {
			s.Value = 0
		}

		s.prevValue = curr
	}

	// Update attributes with new stats.
	s.updateAttributes(stats)
}

//nolint:exhaustive
func (s *diskIOSensor) updateAttributes(stats map[stat]uint64) {
	switch s.sensorType {
	case diskReads:
		s.attributes["total_sectors"] = stats[TotalSectorsRead]
		s.attributes["total_milliseconds"] = stats[TotalTimeReading]
	case diskWrites:
		s.attributes["total_sectors"] = stats[TotalSectorsWritten]
		s.attributes["total_milliseconds"] = stats[TotalTimeWriting]
	}
}

func newDiskIOSensor(boottime time.Time, device *device, sensorType ioSensor) *diskIOSensor {
	newSensor := &diskIOSensor{
		device:     device,
		sensorType: sensorType,
		attributes: make(map[string]any),
		Sensor: linux.Sensor{
			StateClassValue: types.StateClassTotalIncreasing,
			UnitsString:     diskCountUnits,
			LastReset:       boottime.Format(time.RFC3339),
		},
	}

	newSensor.attributes["data_source"] = linux.DataSrcSysfs
	newSensor.attributes["native_unit_of_measurement"] = diskCountUnits
	newSensor.attributes["last_reset"] = newSensor.LastReset

	if device.model != "" {
		newSensor.attributes["device_model"] = device.model
	}

	if device.sysFSPath != "" {
		newSensor.attributes["sysfs_path"] = device.sysFSPath
	}

	if device.id != "total" {
		newSensor.IsDiagnostic = true
	}

	return newSensor
}

func newDiskIORateSensor(device *device, sensorType ioSensor) *diskIOSensor {
	newSensor := &diskIOSensor{
		device:     device,
		sensorType: sensorType,
		Sensor: linux.Sensor{
			DeviceClassValue: types.DeviceClassDataRate,
			StateClassValue:  types.StateClassMeasurement,
			UnitsString:      diskRateUnits,
		},
		attributes: make(map[string]any),
	}

	newSensor.attributes["data_source"] = linux.DataSrcSysfs
	newSensor.attributes["native_unit_of_measurement"] = diskRateUnits

	if device.model != "" {
		newSensor.attributes["device_model"] = device.model
	}

	if device.sysFSPath != "" {
		newSensor.attributes["sysfs_path"] = device.sysFSPath
	}

	if device.id != "total" {
		newSensor.IsDiagnostic = true
	}

	return newSensor
}
