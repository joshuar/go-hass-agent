// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go run golang.org/x/tools/cmd/stringer -type=ioSensor -output ioSensors_generated.go -linecomment
package disk

import (
	"context"
	"maps"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
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

type ioRate struct {
	linux.RateValue[uint64]
	rateType ioSensor
}

func newDiskStatSensor(ctx context.Context, device *device, sensorType ioSensor, value uint64, attributes models.Attributes) (models.Entity, error) {
	var (
		icon, units      string
		stateClass       class.SensorStateClass
		diagnosticOption sensor.Option
	)

	name, id := device.generateIdentifiers(sensorType)

	if attributes != nil {
		maps.Copy(attributes, device.generateAttributes())
	} else {
		attributes = device.generateAttributes()
	}

	switch sensorType {
	case diskIOInProgress:
		icon = ioOpsIcon
		stateClass = class.StateMeasurement
		units = diskIOsUnits
	case diskReads, diskWrites:
		if sensorType == diskReads {
			icon = ioReadsIcon
		} else {
			icon = ioWritesIcon
		}

		units = diskCountUnits
		stateClass = class.StateTotal
		attributes["native_unit_of_measurement"] = diskCountUnits
	}

	if device.id != "total" {
		diagnosticOption = sensor.WithCategory(models.Diagnostic)
	} else {
		diagnosticOption = sensor.WithCategory("")
	}

	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.WithUnits(units),
		sensor.WithStateClass(stateClass),
		sensor.WithState(value),
		sensor.WithIcon(icon),
		sensor.WithAttributes(attributes),
		diagnosticOption,
	)
}

func newDiskRateSensor(ctx context.Context, device *device, sensorType ioSensor, value uint64) (models.Entity, error) {
	var (
		diagnosticOption sensor.Option
		icon             string
	)

	name, id := device.generateIdentifiers(sensorType)
	attributes := device.generateAttributes()
	units := diskRateUnits
	stateClass := class.StateMeasurement
	attributes["native_unit_of_measurement"] = diskRateUnits

	switch sensorType {
	case diskReadRate:
		icon = ioReadsIcon
	case diskWriteRate:
		icon = ioWritesIcon
	}

	if device.id != "total" {
		diagnosticOption = sensor.WithCategory(models.Diagnostic)
	} else {
		diagnosticOption = sensor.WithCategory("")
	}

	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.WithUnits(units),
		sensor.WithStateClass(stateClass),
		sensor.WithState(value),
		sensor.WithIcon(icon),
		sensor.WithAttributes(attributes),
		diagnosticOption,
	)
}
