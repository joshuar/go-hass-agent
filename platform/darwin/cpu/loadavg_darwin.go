package cpu

import (
	"context"
	"fmt"
	"time"

	"github.com/reugn/go-quartz/quartz"
	"github.com/shirou/gopsutil/v4/load"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/darwin"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	loadAvgIcon = "mdi:chip"
	loadAvgUnit = "load"

	loadAvgUpdateInterval = time.Minute
	loadAvgUpdateJitter   = 5 * time.Second

	loadAvgWorkerID   = "load_averages"
	loadAvgWorkerDesc = "CPU load averages"

	loadAvgWorkerPreferencesID = prefPrefix + "load_averages"
)

var (
	_ quartz.Job                  = (*loadAvgWorker)(nil)
	_ workers.PollingEntityWorker = (*loadAvgWorker)(nil)
)

// loadAvgWorker creates sensors for the 1, 5 and 15 minute load averages.
type loadAvgWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *LoadAvgPrefs
}

// NewLoadAvgWorker creates a new worker to generate entities for CPU load averages.
func NewLoadAvgWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &loadAvgWorker{
		WorkerMetadata:          models.SetWorkerMetadata(loadAvgWorkerID, loadAvgWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &LoadAvgPrefs{
		UpdateInterval: loadAvgUpdateInterval.String(),
	}

	var err error

	worker.prefs, err = workers.LoadWorkerPreferences(loadAvgWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = loadAvgUpdateInterval
	}

	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, loadAvgUpdateJitter)

	return worker, nil
}

func (w *loadAvgWorker) Execute(ctx context.Context) error {
	avg, err := load.AvgWithContext(ctx)
	if err != nil {
		return fmt.Errorf("could not get load averages: %w", err)
	}

	loadAvgs := []struct {
		name, id string
		value    float64
	}{
		{"CPU load average (1 min)", "cpu_load_average_1_min", avg.Load1},
		{"CPU load average (5 min)", "cpu_load_average_5_min", avg.Load5},
		{"CPU load average (15 min)", "cpu_load_average_15_min", avg.Load15},
	}

	for _, l := range loadAvgs {
		w.OutCh <- models.NewSensor(ctx,
			models.WithName(l.name),
			models.WithID(l.id),
			models.WithUnits(loadAvgUnit),
			models.WithStateClass(models.StateMeasurement),
			models.WithIcon(loadAvgIcon),
			models.WithState(l.value),
			models.WithDataSourceAttribute(darwin.DataSrcSysctl),
		)
	}

	return nil
}

func (w *loadAvgWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start load averages worker: %w", err)
	}

	return w.OutCh, nil
}

func (w *loadAvgWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
