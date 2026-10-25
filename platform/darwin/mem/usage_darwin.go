package mem

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/reugn/go-quartz/quartz"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/darwin"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	usageUpdateInterval = time.Minute
	usageUpdateJitter   = 5 * time.Second

	usageWorkerID   = "memory_usage"
	usageWorkerDesc = "Memory usage stats"

	memorySensorIcon = "mdi:memory"

	bytesUnits   = "B"
	percentUnits = "%"
)

var (
	_ quartz.Job                  = (*usageWorker)(nil)
	_ workers.PollingEntityWorker = (*usageWorker)(nil)
)

// usageWorker creates sensors for memory usage.
type usageWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *WorkerPrefs
}

// NewUsageWorker creates a new worker to generate entities for memory usage.
func NewUsageWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &usageWorker{
		WorkerMetadata:          models.SetWorkerMetadata(usageWorkerID, usageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &WorkerPrefs{
		UpdateInterval: usageUpdateInterval.String(),
	}

	var err error

	worker.prefs, err = workers.LoadWorkerPreferences(memUsagePreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = usageUpdateInterval
	}

	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, usageUpdateJitter)

	return worker, nil
}

func (w *usageWorker) Execute(ctx context.Context) error {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("could not get memory usage: %w", err)
	}

	w.OutCh <- newByteSensor(ctx, "Memory Total", "memory_total", vm.Total)
	w.OutCh <- newByteSensor(ctx, "Memory Used", "memory_used", vm.Used)
	w.OutCh <- newByteSensor(ctx, "Memory Available", "memory_available", vm.Available)
	w.OutCh <- newPercentSensor(ctx, "Memory Used %", "memory_used_pc", vm.UsedPercent)

	return nil
}

func newByteSensor(ctx context.Context, name, id string, value uint64) models.Entity {
	return models.NewSensor(ctx,
		models.WithName(name),
		models.WithID(id),
		models.WithUnits(bytesUnits),
		models.WithDeviceClass(models.SensorClassDataSize),
		models.WithStateClass(models.StateMeasurement),
		models.WithIcon(memorySensorIcon),
		models.WithState(value),
		models.WithDataSourceAttribute(darwin.DataSrcMach),
		models.WithAttribute("native_unit_of_measurement", bytesUnits),
	)
}

func newPercentSensor(ctx context.Context, name, id string, value float64) models.Entity {
	return models.NewSensor(ctx,
		models.WithName(name),
		models.WithID(id),
		models.WithUnits(percentUnits),
		models.WithStateClass(models.StateMeasurement),
		models.WithIcon(memorySensorIcon),
		models.WithState(math.Round(value/0.05)*0.05),
		models.WithDataSourceAttribute(darwin.DataSrcMach),
		models.WithAttribute("native_unit_of_measurement", percentUnits),
	)
}

func (w *usageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start memory usage worker: %w", err)
	}

	return w.OutCh, nil
}

func (w *usageWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
