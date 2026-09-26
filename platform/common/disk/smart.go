// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package disk

import (
	"context"
	"fmt"
	"maps"

	"github.com/anatol/smart.go"

	"github.com/joshuar/go-hass-agent/models"
)

// SmartData is an interface that represents SMART data from any type of disk (nvme, ata, etc.).
type SmartData interface {
	ID() string
	Problem() bool
	Attributes() map[string]any
}

// DiskDetails are the common details about any disk.
type DiskDetails struct {
	Disk   string
	Serial string
	Model  string
}

func (d *DiskDetails) ID() string {
	return d.Disk
}

func (d *DiskDetails) Details() map[string]any {
	return map[string]any{
		"Disk":   d.Disk,
		"Model":  d.Model,
		"Serial": d.Serial,
	}
}

// NVMeSmartDetails are the SMART details for NVMe disks.
type NVMeSmartDetails struct {
	*DiskDetails
	*smart.NvmeSMARTLog
}

// Problem returns a boolean indicating whether the SMART data indicates a problem for the NVMe disk. The heuristic for
// a problem is if the CritWarning attribute has a value greater than zero. For code spelunkers, if you have
// suggestions, please open a GitHub issue with your comments!
func (nvme *NVMeSmartDetails) Problem() bool {
	return nvme.CritWarning != 0
}

func (nvme *NVMeSmartDetails) Attributes() map[string]any {
	nvmeAttrs := map[string]any{
		"Temperature":   fmt.Sprintf("%.2f °C", KelvinToCelsius(nvme.Temperature)),
		"Percent Used":  fmt.Sprintf("%d %%", nvme.PercentUsed),
		"Percent Spare": fmt.Sprintf("%d %%", nvme.AvailSpare),
	}
	attrs := maps.Clone(nvme.Details())
	maps.Copy(attrs, nvmeAttrs)
	return attrs
}

// NewSmartSensor builds the SMART status binary sensor entity for a disk.
func NewSmartSensor(ctx context.Context, data SmartData) models.Entity {
	return models.NewSensor(ctx,
		models.WithName(data.ID()+" SMART Status"),
		models.WithID(data.ID()+"_smart_status"),
		models.AsTypeBinarySensor(),
		models.WithDeviceClass(models.BinaryClassProblem),
		models.AsDiagnostic(),
		models.WithIcon("mdi:harddisk"),
		models.WithState(data.Problem()),
		models.WithAttributes(data.Attributes()),
	)
}

func KelvinToCelsius[T ~int | ~uint16](kelvin T) float32 {
	return float32(kelvin) - 273.15
}
