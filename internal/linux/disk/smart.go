// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package disk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/anatol/smart.go"
	"github.com/jaypipes/ghw"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"kernel.org/pub/linux/libs/security/libcap/cap"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/scheduler"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

var (
	_ quartz.Job                  = (*smartWorker)(nil)
	_ workers.PollingEntityWorker = (*smartWorker)(nil)
)

var ErrSmartWorker = errors.New("smart worker encountered an error")

// ignoredSmartDiskPatterns contains prefix patterns of disk devices that don't support SMART and should not be
// processed.
var ignoredSmartDiskPatterns = []string{"sr", "dm", "loop", "zram"}

// smartWorkerRequiredChecks are the groups and capabilities required for monitoring SMART attributes.
//
// Permissions required for accessing SMART data:
//
// - cap_sys_rawio,cap_sys_admin,cap_mknod,cap_dac_override=+ep capabilities set.
var smartWorkerRequiredChecks = &linux.Checks{
	Capabilities: []cap.Value{cap.SYS_RAWIO, cap.SYS_ADMIN, cap.MKNOD, cap.DAC_OVERRIDE},
}

const (
	smartWorkerUpdateInterval = time.Minute
	smartWorkerUpdateJitter   = 15 * time.Second
	smartWorkerID             = "smart_status_sensors"
	smartWorkerDesc           = "Disk SMART Status"
)

// smartWorker creates sensors for per disk SMART status.
type smartWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData
	prefs *WorkerPrefs
}

func NewSmartWorker(ctx context.Context) (workers.EntityWorker, error) {
	passed, err := smartWorkerRequiredChecks.Passed()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSmartWorker, err)
	}
	if !passed {
		return nil, fmt.Errorf("%w: required process permissions are missing", ErrSmartWorker)
	}

	smartWorker := &smartWorker{
		WorkerMetadata:          models.SetWorkerMetadata(ioWorkerID, ioWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	prefs, err := preferences.LoadWorker(smartWorker)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSmartWorker, err)
	}
	smartWorker.prefs = prefs

	pollInterval, err := time.ParseDuration(prefs.UpdateInterval)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Invalid polling interval, using default",
			slog.String("worker", smartWorkerID),
			slog.String("given_interval", prefs.UpdateInterval),
			slog.String("default_interval", smartWorkerUpdateInterval.String()))

		pollInterval = smartWorkerUpdateInterval
	}
	smartWorker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, smartWorkerUpdateJitter)

	return smartWorker, nil
}

//nolint:funlen
func (w *smartWorker) Execute(ctx context.Context) error {
	block, err := ghw.Block()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSmartWorker, err)
	}

	for _, disk := range block.Disks {
		if slices.ContainsFunc(ignoredSmartDiskPatterns, func(pattern string) bool {
			return strings.HasPrefix(disk.Name, pattern)
		}) {
			continue
		}
		dev, err := smart.Open("/dev/" + disk.Name)
		if err != nil {
			slogctx.FromCtx(ctx).Debug("Could not read SMART data from device.",
				slog.String("device", disk.Name),
				slog.Any("error", err),
			)
			continue
		}
		defer dev.Close() //nolint:errcheck

		var smartData smartData

		details := &diskDetails{
			Disk:   disk.Name,
			Model:  disk.Model,
			Serial: disk.SerialNumber,
		}

		switch smartDevice := dev.(type) {
		case *smart.SataDevice:
			data, err := smartDevice.ReadSMARTData()
			if err != nil {
				slogctx.FromCtx(ctx).Debug("Failed to read SATA disk SMART data.",
					slog.String("device", details.Disk),
					slog.Any("error", err),
				)
				continue
			}
			ataSmart := &ataSmartDetails{
				diskDetails:  details,
				AtaSmartPage: data,
			}
			smartData = ataSmart
		case *smart.ScsiDevice:
			// Inspect the SCSI disk.
			inq, err := smartDevice.Inquiry()
			if err != nil {
				slogctx.FromCtx(ctx).Debug("Failed to read SCSI disk SMART data.",
					slog.String("device", details.Disk),
					slog.Any("error", err),
				)
				continue
			}
			// If it indicates ATA, treat it as a SATA disk.
			var ataSmart *smart.SataDevice
			if string(inq.VendorIdent[:]) == "ATA     " {
				ataSmart, err = smart.OpenSata("/dev/" + disk.Name)
				if err != nil {
					slogctx.FromCtx(ctx).Debug("Failed to read SCSI disk as SATA device.",
						slog.String("device", details.Disk),
						slog.Any("error", err),
					)
					continue
				}
			}
			data, err := ataSmart.ReadSMARTData()
			if err != nil {
				slogctx.FromCtx(ctx).Debug("Failed to read SATA disk SMART data.",
					slog.String("device", details.Disk),
					slog.Any("error", err),
				)
				continue
			}
			scsiSmart := &ataSmartDetails{
				diskDetails:  details,
				AtaSmartPage: data,
			}
			smartData = scsiSmart
		case *smart.NVMeDevice:
			data, err := smartDevice.ReadSMART()
			if err != nil {
				slogctx.FromCtx(ctx).Debug("Failed to read NVMe disk SMART data.",
					slog.String("device", details.Disk),
					slog.Any("error", err),
				)
				continue
			}
			nvmeSmart := &nvmeSmartDetails{
				diskDetails:  details,
				NvmeSMARTLog: data,
			}
			smartData = nvmeSmart
		}
		if smartData != nil {
			w.OutCh <- newSmartSensor(ctx, smartData)
		}
	}
	return nil
}

func (w *smartWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("%w: %w", ErrSmartWorker, err)
	}
	return w.OutCh, nil
}

func (w *smartWorker) PreferencesID() string {
	return smartWorkerPreferencesID
}

func (w *smartWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{
		UpdateInterval: smartWorkerUpdateInterval.String(),
	}
}

func (w *smartWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

// smartData is an interface that represents SMART data from any type of disk (nvme, ata, etc.).
type smartData interface {
	ID() string
	Problem() bool
	Attributes() map[string]any
}

// diskDetails are the common details about any disk.
type diskDetails struct {
	Disk   string
	Serial string
	Model  string
}

func (disk *diskDetails) ID() string {
	return disk.Disk
}

func (disk *diskDetails) details() map[string]any {
	return map[string]any{
		"Disk":   disk.Disk,
		"Model":  disk.Model,
		"Serial": disk.Serial,
	}
}

// nvmeSmartDetails are the SMART details for nvme disks.
type nvmeSmartDetails struct {
	*diskDetails
	*smart.NvmeSMARTLog
}

// Problem returns a boolean indicating whether the SMART data indicates a problem for the NVMe disk. The heuristic for
// a problem is if the CritWarning attribute has a value greater than zero. For code spelunkers, if you have
// suggestions, please open a GitHub issue with your comments!
func (nvme *nvmeSmartDetails) Problem() bool {
	return nvme.CritWarning != 0
}

func (nvme *nvmeSmartDetails) Attributes() map[string]any {
	nvmeAttrs := map[string]any{
		"Temperature":   fmt.Sprintf("%.2f °C", kelvinToCelsius(nvme.Temperature)),
		"Percent Used":  fmt.Sprintf("%d %%", nvme.PercentUsed),
		"Percent Spare": fmt.Sprintf("%d %%", nvme.AvailSpare),
	}
	attrs := maps.Clone(nvme.details())
	maps.Copy(attrs, nvmeAttrs)
	return attrs
}

// ataSmartDetails are the SMART details for ata disks.
type ataSmartDetails struct {
	*diskDetails
	*smart.AtaSmartPage
}

// Problem returns a boolean indicating whether the SMART data indicates a problem for the ATA disk.
//
// TODO: The heuristic for a problem is any of the values marked as critical at the following link with values greater
// than zero. This may not be ideal. For code spelunkers, if you have suggestions, please open a GitHub issue with your
// comments!
//
// https://en.wikipedia.org/wiki/Self-Monitoring,_Analysis_and_Reporting_Technology#In_ATA
//
//nolint:gocyclo // ¯\_(ツ)_/¯
func (ata *ataSmartDetails) Problem() bool {
	// Read Error Rate value greater than zero.
	if attr, ok := ata.Attrs[1]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	// Reallocated Sectors Count greater than zero.
	if attr, ok := ata.Attrs[5]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	// Spin Retry Count greater than zero.
	if attr, ok := ata.Attrs[10]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	//  End-to-End error / IOEDC  greater than zero.
	if attr, ok := ata.Attrs[184]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	// Reported Uncorrectable Errors greater than zero.
	if attr, ok := ata.Attrs[187]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	// Command Timeout greater than zero.
	if attr, ok := ata.Attrs[188]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	// Reallocation Event Count greater than zero.
	if attr, ok := ata.Attrs[196]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	// Current Pending Sector Count greater than zero.
	if attr, ok := ata.Attrs[197]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	// (Offline) Uncorrectable Sector Count greater than zero.
	if attr, ok := ata.Attrs[198]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	// Soft Read Error Rate or TA Counter Detected greater than zero.
	if attr, ok := ata.Attrs[201]; ok {
		if attr.Current > 0 {
			return true
		}
	}
	return false
}

// Attributes are the SMART attributes for ATA disks. Trying to expose most common/interesting attributes. For code
// spelunkers, if you have suggestions, please open a GitHub issue with your comments!
//
//nolint:gocyclo,funlen // ¯\_(ツ)_/¯
func (ata *ataSmartDetails) Attributes() map[string]any {
	ataAttrs := make(map[string]any)
	// Read Error Rate.
	if attr, ok := ata.Attrs[1]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Throughput Performance.
	if attr, ok := ata.Attrs[2]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Spin-Up Time.
	if attr, ok := ata.Attrs[3]; ok {
		dur, err := attr.ParseAsDuration()
		if err == nil {
			ataAttrs[attr.Name] = dur.String()
		}
	}
	// Start/Stop Count.
	if attr, ok := ata.Attrs[4]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Reallocated Sectors Count.
	if attr, ok := ata.Attrs[5]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Seek Time Performance
	if attr, ok := ata.Attrs[8]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Power-On Hours.
	if attr, ok := ata.Attrs[9]; ok {
		dur, err := attr.ParseAsDuration()
		if err == nil {
			ataAttrs[attr.Name] = dur.String()
		}
	}
	// Spin Retry Count.
	if attr, ok := ata.Attrs[10]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Power Cycle Count.
	if attr, ok := ata.Attrs[12]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Soft Read Error Rate.
	if attr, ok := ata.Attrs[13]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Unexpected Power Loss Count.
	if attr, ok := ata.Attrs[174]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	//  End-to-End error / IOEDC.
	if attr, ok := ata.Attrs[184]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Reported Uncorrectable Errors.
	if attr, ok := ata.Attrs[187]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Command Timeout.
	if attr, ok := ata.Attrs[188]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// High Fly Writes.
	if attr, ok := ata.Attrs[189]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// G-sense Error Rate.
	if attr, ok := ata.Attrs[191]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Unsafe Shutdown Count.
	if attr, ok := ata.Attrs[192]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Temperature.
	if attr, ok := ata.Attrs[194]; ok {
		temp, _, _, _, err := attr.ParseAsTemperature()
		if err == nil {
			ataAttrs[attr.Name] = fmt.Sprintf("%d °C", temp)
		}
	}
	// Reallocation Event Count greater than zero.
	if attr, ok := ata.Attrs[196]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Current Pending Sector Count greater than zero.
	if attr, ok := ata.Attrs[197]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// (Offline) Uncorrectable Sector Count greater than zero.
	if attr, ok := ata.Attrs[198]; ok {
		ataAttrs[attr.Name] = attr.Current
	}
	// Soft Read Error Rate or TA Counter Detected greater than zero.
	if attr, ok := ata.Attrs[201]; ok {
		ataAttrs[attr.Name] = attr.Current
	}

	attrs := maps.Clone(ata.details())
	maps.Copy(attrs, ataAttrs)
	return attrs
}

type scsiSmartDetails struct {
	*diskDetails
	*smart.GenericAttributes
}

func (scsi *scsiSmartDetails) Problem() bool {
	// TODO: work out a health heuristic for SCSI devices. Currently doesn't seem to be any attributes/status exposed by
	// library.
	return false
}

func (scsi *scsiSmartDetails) Attributes() map[string]any {
	scsiattrs := make(map[string]any)
	scsiattrs["Temperature"] = fmt.Sprintf("%d °C", scsi.Temperature)
	scsiattrs["Power On Hours"] = scsi.PowerOnHours
	scsiattrs["Power Cycles"] = scsi.PowerCycles
	scsiattrs["Read Blocks"] = scsi.Read
	scsiattrs["Written Blocks"] = scsi.Written
	attrs := maps.Clone(scsi.details())
	maps.Copy(attrs, scsiattrs)
	return attrs
}

func newSmartSensor(ctx context.Context, data smartData) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName(data.ID()+" SMART Status"),
		sensor.WithID(data.ID()+"_smart_status"),
		sensor.AsTypeBinarySensor(),
		sensor.WithDeviceClass(class.BinaryClassProblem),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:harddisk"),
		sensor.WithState(data.Problem()),
		sensor.WithAttributes(data.Attributes()),
	)
}

func kelvinToCelsius[T ~int | ~uint16](kelvin T) float32 {
	return float32(kelvin) - 273.15
}
