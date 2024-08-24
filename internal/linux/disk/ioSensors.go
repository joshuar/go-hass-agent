// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package disk

import (
	"time"
	"unicode"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	diskRateUnits  = "kB/s"
	diskCountUnits = "requests"
)

type diskIOSensor struct {
	device     *device
	attributes map[string]any
	linux.Sensor
	prevValue uint64
}

func (s *diskIOSensor) Name() string {
	r := []rune(s.device.id)

	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...)) + " " + s.SensorTypeValue.String()
}

func (s *diskIOSensor) ID() string {
	return s.device.id + "_" + strcase.ToSnake(s.SensorTypeValue.String())
}

func (s *diskIOSensor) Attributes() map[string]any {
	return s.attributes
}

//nolint:exhaustive
func (s *diskIOSensor) Icon() string {
	switch s.SensorTypeValue {
	case linux.SensorDiskReads, linux.SensorDiskReadRate:
		return "mdi:file-upload"
	case linux.SensorDiskWrites, linux.SensorDiskWriteRate:
		return "mdi:file-download"
	}

	return "mdi:file"
}

//nolint:exhaustive,mnd
func (s *diskIOSensor) update(stats map[stat]uint64, delta time.Duration) {
	var curr uint64

	switch s.SensorTypeValue {
	case linux.SensorDiskReads:
		s.Value = stats[TotalReads]
	case linux.SensorDiskWrites:
		s.Value = stats[TotalWrites]
	case linux.SensorDiskReadRate:
		curr = stats[TotalSectorsRead]
	case linux.SensorDiskWriteRate:
		curr = stats[TotalSectorsWritten]
	}

	// For rate sensors, calculate the current value based on previous value and
	// time interval since last measurement.
	if s.SensorTypeValue == linux.SensorDiskReadRate || s.SensorTypeValue == linux.SensorDiskWriteRate {
		if uint64(delta.Seconds()) > 0 {
			s.Value = (curr - s.prevValue) / uint64(delta.Seconds()) / 2
		}

		s.prevValue = curr
	}

	// Update attributes with new stats.
	s.updateAttributes(stats)
}

//nolint:exhaustive
func (s *diskIOSensor) updateAttributes(stats map[stat]uint64) {
	switch s.SensorTypeValue {
	case linux.SensorDiskReads:
		s.attributes["total_sectors"] = stats[TotalSectorsRead]
		s.attributes["total_milliseconds"] = stats[TotalTimeReading]
	case linux.SensorDiskWrites:
		s.attributes["total_sectors"] = stats[TotalSectorsWritten]
		s.attributes["total_milliseconds"] = stats[TotalTimeWriting]
	}
}

func newDiskIOSensor(device *device, sensorType linux.SensorTypeValue) *diskIOSensor {
	newSensor := &diskIOSensor{
		device: device,
		Sensor: linux.Sensor{
			StateClassValue: types.StateClassTotalIncreasing,
			SensorTypeValue: sensorType,
			UnitsString:     diskCountUnits,
		},
		attributes: make(map[string]any),
	}

	newSensor.attributes["data_source"] = linux.DataSrcSysfs
	newSensor.attributes["native_unit_of_measurement"] = diskCountUnits

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

func newDiskIORateSensor(device *device, sensorType linux.SensorTypeValue) *diskIOSensor {
	newSensor := &diskIOSensor{
		device: device,
		Sensor: linux.Sensor{
			DeviceClassValue: types.DeviceClassDataRate,
			StateClassValue:  types.StateClassMeasurement,
			UnitsString:      diskRateUnits,
			SensorTypeValue:  sensorType,
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
