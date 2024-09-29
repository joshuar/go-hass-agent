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

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
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
	*sensor.Entity
	sensorType ioSensor
	prevValue  uint64
}

//nolint:mnd
func (s *diskIOSensor) update(stats map[stat]uint64, delta time.Duration) {
	var curr uint64

	switch s.sensorType {
	case diskReads:
		s.State = stats[TotalReads]
	case diskWrites:
		s.State = stats[TotalWrites]
	case diskReadRate:
		curr = stats[TotalSectorsRead]
	case diskWriteRate:
		curr = stats[TotalSectorsWritten]
	}

	// For rate sensors, calculate the current value based on previous value and
	// time interval since last measurement.
	if s.sensorType == diskReadRate || s.sensorType == diskWriteRate {
		if uint64(delta.Seconds()) > 0 {
			s.State = (curr - s.prevValue) / uint64(delta.Seconds()) / 2
		} else {
			s.State = 0
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
		s.Attributes["total_sectors"] = stats[TotalSectorsRead]
		s.Attributes["total_milliseconds"] = stats[TotalTimeReading]
	case diskWrites:
		s.Attributes["total_sectors"] = stats[TotalSectorsWritten]
		s.Attributes["total_milliseconds"] = stats[TotalTimeWriting]
	}
}

func newDiskIOSensor(boottime time.Time, device *device, sensorType ioSensor) *diskIOSensor {
	newSensor := &diskIOSensor{
		Entity: &sensor.Entity{
			Name:       generateName(device.id, sensorType.String()),
			StateClass: types.StateClassTotalIncreasing,
			Units:      diskCountUnits,
			EntityState: &sensor.EntityState{
				ID:   generateID(device.id, sensorType.String()),
				Icon: generateIcon(sensorType),
				Attributes: map[string]any{
					"data_source":                linux.DataSrcSysfs,
					"native_unit_of_measurement": diskRateUnits,
					"last_reset":                 boottime.Format(time.RFC3339),
				},
			},
		},
	}

	if device.model != "" {
		newSensor.Attributes["device_model"] = device.model
	}

	if device.sysFSPath != "" {
		newSensor.Attributes["sysfs_path"] = device.sysFSPath
	}

	if device.id != "total" {
		newSensor.Category = types.CategoryDiagnostic
	}

	return newSensor
}

func newDiskIORateSensor(device *device, sensorType ioSensor) *diskIOSensor {
	newSensor := &diskIOSensor{
		Entity: &sensor.Entity{
			Name:        generateName(device.id, sensorType.String()),
			DeviceClass: types.SensorDeviceClassDataRate,
			StateClass:  types.StateClassMeasurement,
			Units:       diskRateUnits,
			EntityState: &sensor.EntityState{
				ID:   generateID(device.id, sensorType.String()),
				Icon: generateIcon(sensorType),
				Attributes: map[string]any{
					"data_source":                linux.DataSrcSysfs,
					"native_unit_of_measurement": diskRateUnits,
				},
			},
		},
	}

	if device.model != "" {
		newSensor.Attributes["device_model"] = device.model
	}

	if device.sysFSPath != "" {
		newSensor.Attributes["sysfs_path"] = device.sysFSPath
	}

	if device.id != "total" {
		newSensor.Category = types.CategoryDiagnostic
	}

	return newSensor
}

func generateName(id string, sensorType string) string {
	r := []rune(id)

	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...)) + " " + sensorType
}

func generateID(id string, sensorType string) string {
	return id + "_" + strcase.ToSnake(sensorType)
}

func generateIcon(sensorType ioSensor) string {
	switch sensorType {
	case diskReads, diskReadRate:
		return "mdi:file-upload"
	case diskWrites, diskWriteRate:
		return "mdi:file-download"
	}

	return "mdi:file"
}
