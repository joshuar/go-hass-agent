// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go tool golang.org/x/tools/cmd/stringer -type=ioSensor -output ioSensors_generated.go -linecomment
package disk

import (
	"context"
	"errors"
	"maps"

	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
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

var (
	ErrNewDiskStatSensor = errors.New("could not create disk stat sensor")
	ErrNewDiskRateSensor = errors.New("could not create disk rate sensor")
)

type ioSensor int

type ioRate struct {
	linux.RateValue[uint64]
	rateType ioSensor
}

func newDiskStatSensor(ctx context.Context, device *device, sensorType ioSensor, value uint64, attributes models.Attributes) models.Entity {
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
		diagnosticOption = sensor.WithCategory(models.EntityCategoryDiagnostic)
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

func newDiskRateSensor(ctx context.Context, device *device, sensorType ioSensor, value uint64) models.Entity {
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
		diagnosticOption = sensor.WithCategory(models.EntityCategoryDiagnostic)
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
