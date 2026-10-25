package cpu

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/reugn/go-quartz/quartz"
	"github.com/shirou/gopsutil/v4/cpu"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/darwin"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	usageUpdateInterval = 10 * time.Second
	usageUpdateJitter   = 500 * time.Millisecond
	usageSampleWindow   = 200 * time.Millisecond

	usageWorkerID   = "cpu_usage"
	usageWorkerDesc = "CPU usage metrics"

	cpuIcon = "mdi:chip"

	usageWorkerPreferencesID = prefPrefix + "usage"
)

var (
	_ quartz.Job                  = (*usageWorker)(nil)
	_ workers.PollingEntityWorker = (*usageWorker)(nil)
)

// usageWorker creates sensors for total and per-core CPU usage.
type usageWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *UsagePrefs
}

// NewUsageWorker creates a new worker to generate entities for CPU usage.
func NewUsageWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &usageWorker{
		WorkerMetadata:          models.SetWorkerMetadata(usageWorkerID, usageWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &UsagePrefs{
		UpdateInterval: usageUpdateInterval.String(),
	}

	var err error

	worker.prefs, err = workers.LoadWorkerPreferences(usageWorkerPreferencesID, defaultPrefs)
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

// Execute samples per-core CPU usage over usageSampleWindow and reports a
// sensor for each core plus a "Total CPU Usage" sensor, which is the mean
// of the per-core values (rather than a second sample) to avoid doubling
// the blocking sample time.
func (w *usageWorker) Execute(ctx context.Context) error {
	percpu, err := cpu.PercentWithContext(ctx, usageSampleWindow, true)
	if err != nil {
		return fmt.Errorf("could not get CPU usage: %w", err)
	}

	var total float64

	for core, pct := range percpu {
		total += pct
		w.OutCh <- newUsageSensor(ctx,
			"Core "+strconv.Itoa(core)+" CPU Usage",
			"core_"+strconv.Itoa(core)+"_cpu_usage",
			pct,
			models.EntityCategoryDiagnostic,
		)
	}

	if len(percpu) > 0 {
		total /= float64(len(percpu))
	}

	w.OutCh <- newUsageSensor(ctx, "Total CPU Usage", "total_cpu_usage", total, "")

	return nil
}

func newUsageSensor(ctx context.Context, name, id string, value float64, category models.EntityCategory) models.Entity {
	return models.NewSensor(ctx,
		models.WithName(name),
		models.WithID(id),
		models.WithUnits("%"),
		models.WithStateClass(models.StateMeasurement),
		models.WithIcon(cpuIcon),
		models.WithState(value),
		models.WithCategory(category),
		models.WithDataSourceAttribute(darwin.DataSrcMach),
	)
}

func (w *usageWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start CPU usage worker: %w", err)
	}

	return w.OutCh, nil
}

func (w *usageWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
