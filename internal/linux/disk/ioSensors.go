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
	sensor.Entity
	sensorType ioSensor
	prevValue  uint64
}

//nolint:mnd
func (s *diskIOSensor) update(stats map[stat]uint64, delta time.Duration) {
	var curr uint64

	switch s.sensorType {
	case diskReads:
		s.UpdateValue(stats[TotalReads])
	case diskWrites:
		s.UpdateValue(stats[TotalWrites])
	case diskIOInProgress:
		s.UpdateValue(stats[ActiveIOs])
	case diskReadRate:
		curr = stats[TotalSectorsRead]
	case diskWriteRate:
		curr = stats[TotalSectorsWritten]
	}

	// For rate sensors, calculate the current value based on previous value and
	// time interval since last measurement.
	if s.sensorType == diskReadRate || s.sensorType == diskWriteRate {
		if uint64(delta.Seconds()) > 0 {
			s.UpdateValue((curr - s.prevValue) / uint64(delta.Seconds()) / 2)
		} else {
			s.UpdateValue(0)
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
		s.UpdateAttribute("total_sectors", stats[TotalSectorsRead])
		s.UpdateAttribute("total_milliseconds", stats[TotalTimeReading])
	case diskWrites:
		s.UpdateAttribute("total_sectors", stats[TotalSectorsWritten])
		s.UpdateAttribute("total_milliseconds", stats[TotalTimeWriting])
	}
}

func newDiskIOSensor(device *device, sensorType ioSensor, boottime time.Time) *diskIOSensor {
	r := []rune(device.id)
	name := string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...)) + " " + sensorType.String()
	id := strings.ToLower(device.id + "_" + strings.ReplaceAll(sensorType.String(), " ", "_"))

	var (
		icon, units string
		stateClass  types.StateClass
	)

	attributes := map[string]any{
		"data_source": linux.DataSrcSysfs,
	}

	// Add attributes from device if available.
	if device.model != "" {
		attributes["device_model"] = device.model
	}

	if device.sysFSPath != "" {
		attributes["sysfs_path"] = device.sysFSPath
	}

	// Fill out additional fields based on sensor type.
	switch sensorType {
	case diskIOInProgress:
		icon = ioOpsIcon
		stateClass = types.StateClassMeasurement
		units = diskIOsUnits
	case diskReads, diskWrites:
		if sensorType == diskReads {
			icon = ioReadsIcon
		} else {
			icon = ioWritesIcon
		}

		units = diskCountUnits
		stateClass = types.StateClassTotalIncreasing
		attributes["native_unit_of_measurement"] = diskCountUnits
		attributes["last_reset"] = boottime.Format(time.RFC3339)
	case diskReadRate, diskWriteRate:
		if sensorType == diskReadRate {
			icon = ioReadsIcon
		} else {
			icon = ioWritesIcon
		}

		units = diskRateUnits
		stateClass = types.StateClassMeasurement
		attributes["native_unit_of_measurement"] = diskRateUnits
	}

	ioSensor := &diskIOSensor{
		Entity: sensor.NewSensor(
			sensor.WithName(name),
			sensor.WithID(id),
			sensor.WithUnits(units),
			sensor.WithStateClass(stateClass),
			sensor.WithState(
				sensor.WithIcon(icon),
				sensor.WithAttributes(attributes),
			),
		),
		sensorType: sensorType,
	}

	if device.id != "total" {
		ioSensor.Entity = sensor.AsDiagnostic()(ioSensor.Entity)
	}

	if sensorType == diskReadRate || sensorType == diskWriteRate {
		ioSensor.Entity = sensor.WithDeviceClass(types.SensorDeviceClassDataRate)(ioSensor.Entity)
	}

	return ioSensor
}
