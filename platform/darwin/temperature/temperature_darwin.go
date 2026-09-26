package temperature

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/reugn/go-quartz/quartz"
	"github.com/shirou/gopsutil/v4/sensors"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/darwin"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	temperatureUpdateInterval = time.Minute
	temperatureUpdateJitter   = 5 * time.Second

	temperatureWorkerID   = "temperature_sensors"
	temperatureWorkerDesc = "Hardware temperature sensors"

	temperatureUnits = "°C"
	temperatureIcon  = "mdi:thermometer"
)

var (
	_ quartz.Job                  = (*temperatureWorker)(nil)
	_ workers.PollingEntityWorker = (*temperatureWorker)(nil)
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// temperatureWorker creates a sensor for each temperature sensor macOS
// exposes: IOHID event sensors on Apple Silicon, SMC keys on Intel.
type temperatureWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *WorkerPrefs
	read  func(ctx context.Context) ([]sensors.TemperatureStat, error)
}

// NewTemperatureWorker creates a new polling entity worker for reporting hardware temperatures.
func NewTemperatureWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &temperatureWorker{
		WorkerMetadata:          models.SetWorkerMetadata(temperatureWorkerID, temperatureWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		read:                    sensors.TemperaturesWithContext,
	}

	defaultPrefs := &WorkerPrefs{
		UpdateInterval: temperatureUpdateInterval.String(),
	}

	var err error

	worker.prefs, err = workers.LoadWorkerPreferences(temperatureWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = temperatureUpdateInterval
	}

	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, temperatureUpdateJitter)

	return worker, nil
}

// Execute reads all temperature sensors and reports each valid one.
func (w *temperatureWorker) Execute(ctx context.Context) error {
	stats, err := w.read(ctx)
	// gopsutil may return partial results alongside a warning-type error.
	if err != nil && len(stats) == 0 {
		return fmt.Errorf("could not read temperatures: %w", err)
	}

	slices.SortFunc(stats, func(a, b sensors.TemperatureStat) int {
		return strings.Compare(a.SensorKey, b.SensorKey)
	})

	seen := make(map[string]bool, len(stats))

	for _, stat := range stats {
		if seen[stat.SensorKey] || !validTemperature(stat.Temperature) {
			continue
		}

		seen[stat.SensorKey] = true
		w.OutCh <- newTemperatureSensor(ctx, stat)
	}

	return nil
}

// validTemperature filters out readings that aren't real measurements: unused
// Apple Silicon sensors report placeholder values such as -21.5, and
// unreadable Intel SMC keys come back as 0.
func validTemperature(t float64) bool {
	return !math.IsNaN(t) && t > 0 && t < 150
}

func newTemperatureSensor(ctx context.Context, stat sensors.TemperatureStat) models.Entity {
	id := "temperature_" + strings.Trim(nonAlphanumeric.ReplaceAllString(strings.ToLower(stat.SensorKey), "_"), "_")

	return models.NewSensor(ctx,
		models.WithName(stat.SensorKey),
		models.WithID(id),
		models.AsTypeSensor(),
		models.WithDeviceClass(models.SensorClassTemperature),
		models.WithStateClass(models.StateMeasurement),
		models.AsDiagnostic(),
		models.WithUnits(temperatureUnits),
		models.WithIcon(temperatureIcon),
		models.WithState(math.Round(stat.Temperature*10)/10),
		models.WithAttributes(map[string]any{
			"data_source":                darwin.DataSrcIOKit,
			"native_unit_of_measurement": temperatureUnits,
		}),
	)
}

func (w *temperatureWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("schedule worker: %w", err)
	}

	return w.OutCh, nil
}

func (w *temperatureWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
