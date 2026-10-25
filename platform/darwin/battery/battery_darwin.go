package battery

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"time"

	"github.com/reugn/go-quartz/quartz"
	"howett.net/plist"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/darwin"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	batteryUpdateInterval = 30 * time.Second
	batteryUpdateJitter   = 5 * time.Second

	batteryWorkerID   = "battery_sensors"
	batteryWorkerDesc = "Battery status"

	batteryIcon = "mdi:battery"

	percentUnits = "%"
)

var (
	_ quartz.Job                  = (*batteryWorker)(nil)
	_ workers.PollingEntityWorker = (*batteryWorker)(nil)
)

// batteryInfo is the subset of ioreg's AppleSmartBattery properties this
// worker needs.
type batteryInfo struct {
	Serial            string `plist:"Serial"`
	DeviceName        string `plist:"DeviceName"`
	BuiltIn           bool   `plist:"built-in"`
	CurrentCapacity   int    `plist:"CurrentCapacity"`
	MaxCapacity       int    `plist:"MaxCapacity"`
	CycleCount        int    `plist:"CycleCount"`
	Voltage           int    `plist:"Voltage"`  // millivolts
	Amperage          int    `plist:"Amperage"` // milliamps
	IsCharging        bool   `plist:"IsCharging"`
	FullyCharged      bool   `plist:"FullyCharged"`
	ExternalConnected bool   `plist:"ExternalConnected"`
}

// percentage returns the current charge as a percentage, rounded to the
// nearest whole number. CurrentCapacity/MaxCapacity are already percentages
// on modern macOS versions, but this is deliberately computed as a ratio
// rather than read directly in case that's not true on every version/model.
func (b *batteryInfo) percentage() float64 {
	if b.MaxCapacity == 0 {
		return 0
	}
	return math.Round(float64(b.CurrentCapacity) / float64(b.MaxCapacity) * 100)
}

// state returns a human-readable charging state, mirroring the strings used
// by the Linux battery worker (platform/linux/battery) where macOS exposes
// an equivalent concept. "Not Charging" covers macOS-specific cases such as
// optimized battery charging intentionally holding below 100%.
func (b *batteryInfo) state() string {
	switch {
	case !b.ExternalConnected:
		return "Discharging"
	case b.FullyCharged:
		return "Fully Charged"
	case b.IsCharging:
		return "Charging"
	default:
		return "Not Charging"
	}
}

func (b *batteryInfo) attributes() map[string]any {
	return map[string]any{
		"data_source": darwin.DataSrcIoreg,
		"cycle_count": b.CycleCount,
		"voltage":     float64(b.Voltage) / 1000,
		"amperage":    b.Amperage,
		"serial":      b.Serial,
	}
}

// getBatteryInfo fetches the current state of the internal battery. It
// returns nil, nil if there is no battery (e.g. a Mac desktop).
func getBatteryInfo(_ context.Context) (*batteryInfo, error) {
	out, err := exec.Command("ioreg", "-n", "AppleSmartBattery", "-r", "-a").Output()
	if err != nil {
		return nil, fmt.Errorf("ioreg query failed: %w", err)
	}
	if len(out) == 0 {
		return nil, nil
	}

	var data []batteryInfo
	if _, err := plist.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("ioreg unmarshal failed: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}

	return &data[0], nil
}

func newPercentageSensor(ctx context.Context, b *batteryInfo) models.Entity {
	return models.NewSensor(ctx,
		models.WithName(batteryName(b)+" Battery Level"),
		models.WithID("battery_level"),
		models.WithDeviceClass(models.SensorClassBattery),
		models.WithStateClass(models.StateMeasurement),
		models.WithUnits(percentUnits),
		models.WithIcon(batteryPercentIcon(b.percentage())),
		models.WithState(b.percentage()),
		models.WithAttributes(b.attributes()),
	)
}

func newStateSensor(ctx context.Context, b *batteryInfo) models.Entity {
	return models.NewSensor(ctx,
		models.WithName(batteryName(b)+" Battery State"),
		models.WithID("battery_state"),
		models.AsDiagnostic(),
		models.WithIcon(batteryIcon),
		models.WithState(b.state()),
		models.WithAttributes(b.attributes()),
	)
}

func batteryName(b *batteryInfo) string {
	if b.DeviceName == "" {
		return "Battery"
	}
	return b.DeviceName
}

func batteryPercentIcon(percentage float64) string {
	if percentage >= 95 {
		return batteryIcon
	}
	return fmt.Sprintf("%s-%d", batteryIcon, int(math.Round(percentage/10)*10))
}

// batteryWorker creates sensors for the internal battery's charge level and
// charging state.
type batteryWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs *WorkerPrefs
}

// NewBatteryWorker creates a new polling entity worker for reporting battery status.
func NewBatteryWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &batteryWorker{
		WorkerMetadata:          models.SetWorkerMetadata(batteryWorkerID, batteryWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	defaultPrefs := &WorkerPrefs{
		UpdateInterval: batteryUpdateInterval.String(),
	}

	var err error

	worker.prefs, err = workers.LoadWorkerPreferences(batteryWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = batteryUpdateInterval
	}

	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, batteryUpdateJitter)

	return worker, nil
}

// Execute fetches and reports the current battery status. If the device has
// no battery (e.g. a Mac desktop), it reports nothing and no error.
func (w *batteryWorker) Execute(ctx context.Context) error {
	info, err := getBatteryInfo(ctx)
	if err != nil {
		return fmt.Errorf("could not read battery info: %w", err)
	}
	if info == nil {
		return nil
	}

	w.OutCh <- newPercentageSensor(ctx, info)
	w.OutCh <- newStateSensor(ctx, info)

	return nil
}

func (w *batteryWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("schedule worker: %w", err)
	}

	return w.OutCh, nil
}

func (w *batteryWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}
