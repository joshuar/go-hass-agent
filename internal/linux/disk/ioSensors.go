// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go run golang.org/x/tools/cmd/stringer -type=ioSensor -output ioSensors_generated.go -linecomment
package disk

import (
	"strings"
	"time"
	"unicode"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
)

const (
	diskReads        ioSensor = iota // Disk Reads
	diskWrites                       // Disk Writes
	diskReadRate                     // Disk Read Rate
	diskWriteRate                    // Disk Write Rate
	diskIOInProgress                 // Disk IOs In Progress

	diskRateUnits  = "kB/s"
	diskCountUnits = "requests"
	diskIOsUnits   = "ops"

	ioReadsIcon  = "mdi:file-upload"
	ioWritesIcon = "mdi:file-download"
	ioOpsIcon    = "mdi:content-save"
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
		s.Value = stats[TotalReads]
	case diskWrites:
		s.Value = stats[TotalWrites]
	case diskIOInProgress:
		s.Value = stats[ActiveIOs]
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
		s.Attributes["total_sectors"] = stats[TotalSectorsRead]
		s.Attributes["total_milliseconds"] = stats[TotalTimeReading]
	case diskWrites:
		s.Attributes["total_sectors"] = stats[TotalSectorsWritten]
		s.Attributes["total_milliseconds"] = stats[TotalTimeWriting]
	}
}

func newDiskIOSensor(device *device, sensorType ioSensor, boottime time.Time) *diskIOSensor {
	r := []rune(device.id)
	name := string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...)) + " " + sensorType.String()
	id := strings.ToLower(device.id + "_" + strings.ReplaceAll(sensorType.String(), " ", "_"))

	// Base diskIOSensor fields.
	ioSensor := &diskIOSensor{
		Entity: &sensor.Entity{
			Name: name,
			State: &sensor.State{
				ID: id,
				Attributes: map[string]any{
					"data_source": linux.DataSrcSysfs,
				},
			},
		},
		sensorType: sensorType,
	}

	// Add attributes from device if available.
	if device.model != "" {
		ioSensor.Attributes["device_model"] = device.model
	}

	if device.sysFSPath != "" {
		ioSensor.Attributes["sysfs_path"] = device.sysFSPath
	}

	if device.id != "total" {
		ioSensor.Category = types.CategoryDiagnostic
	}

	// Fill out additional fields based on sensor type.
	switch sensorType {
	case diskIOInProgress:
		ioSensor.Icon = ioOpsIcon
		ioSensor.StateClass = types.StateClassMeasurement
		ioSensor.Units = diskIOsUnits
	case diskReads, diskWrites:
		if sensorType == diskReads {
			ioSensor.Icon = ioReadsIcon
		} else {
			ioSensor.Icon = ioWritesIcon
		}

		ioSensor.Units = diskCountUnits
		ioSensor.Attributes["native_unit_of_measurement"] = diskCountUnits
		ioSensor.StateClass = types.StateClassTotalIncreasing
		ioSensor.Attributes["last_reset"] = boottime.Format(time.RFC3339)
	case diskReadRate, diskWriteRate:
		if sensorType == diskReadRate {
			ioSensor.Icon = ioReadsIcon
		} else {
			ioSensor.Icon = ioWritesIcon
		}

		ioSensor.Units = diskRateUnits
		ioSensor.Attributes["native_unit_of_measurement"] = diskRateUnits
		ioSensor.DeviceClass = types.SensorDeviceClassDataRate
		ioSensor.StateClass = types.StateClassMeasurement
	}

	return ioSensor
}
