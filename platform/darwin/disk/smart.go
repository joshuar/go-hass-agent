package disk

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/anatol/smart.go"
	"github.com/jaypipes/ghw"
	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	commondisk "github.com/joshuar/go-hass-agent/platform/common/disk"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	smartWorkerUpdateInterval = time.Minute
	smartWorkerUpdateJitter   = 15 * time.Second

	smartWorkerID   = "smart_status"
	smartWorkerDesc = "Report SMART data for disks"
)

var (
	_ quartz.Job                  = (*smartWorker)(nil)
	_ workers.PollingEntityWorker = (*smartWorker)(nil)
)

// smartWorker creates a sensor for the internal NVMe drive's SMART status.
// Only NVMe is supported: anatol/smart.go has no SATA/SCSI backend for
// darwin, which is fine since virtually every Mac still sold uses NVMe
// storage internally.
type smartWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *WorkerPrefs
}

// NewSmartWorker creates a new polling entity worker for monitoring SMART disk status.
func NewSmartWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &smartWorker{
		WorkerMetadata:          models.SetWorkerMetadata(smartWorkerID, smartWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &WorkerPrefs{
		UpdateInterval: smartWorkerUpdateInterval.String(),
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(smartWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = smartWorkerUpdateInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, smartWorkerUpdateJitter)

	return worker, nil
}

// Execute fetches and reports current SMART status for the NVMe disk(s).
//
// macOS exposes the internal SSD as several BSD disk identifiers (the raw
// disk plus its APFS container and any synthesized volumes), all backed by
// the same physical NVMe controller and reporting identical SMART data.
// Disks are deduplicated by serial number so the same drive isn't reported
// more than once.
func (w *smartWorker) Execute(ctx context.Context) error {
	block, err := ghw.Block()
	if err != nil {
		return fmt.Errorf("get block devices: %w", err)
	}

	seen := make(map[string]bool)

	for _, disk := range block.Disks {
		key := disk.SerialNumber
		if key == "" {
			key = disk.Name
		}
		if seen[key] {
			continue
		}
		seen[key] = true

		dev, err := smart.OpenNVMe(disk.Name)
		if err != nil {
			slogctx.FromCtx(ctx).Debug("Could not open NVMe device.",
				slog.String("device", disk.Name),
				slog.Any("error", err),
			)
			continue
		}

		data, err := dev.ReadSMART()
		dev.Close()
		if err != nil {
			slogctx.FromCtx(ctx).Debug("Failed to read NVMe disk SMART data.",
				slog.String("device", disk.Name),
				slog.Any("error", err),
			)
			continue
		}

		smartData := &commondisk.NVMeSmartDetails{
			DiskDetails: &commondisk.DiskDetails{
				Disk:   disk.Name,
				Model:  disk.Model,
				Serial: disk.SerialNumber,
			},
			NvmeSMARTLog: data,
		}

		w.OutCh <- commondisk.NewSmartSensor(ctx, smartData)
	}

	return nil
}

// Start starts the polling entity worker for monitoring disk SMART status.
func (w *smartWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("schedule worker: %w", err)
	}
	return w.OutCh, nil
}

func (w *smartWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
