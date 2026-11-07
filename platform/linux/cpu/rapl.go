// Copyright 2026 opticon454
// SPDX-License-Identifier: MIT

package cpu

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	raplUpdateInterval = 10 * time.Second
	raplUpdateJitter   = 500 * time.Millisecond
	raplPreferencesID  = prefPrefix + "power"

	raplPowercapDir = "class/powercap"
	raplDomainGlob  = "intel-rapl:*"
	raplIcon        = "mdi:flash"
	raplUnits       = "W"

	microjoulesPerJoule = 1e6
)

var (
	ErrInitRAPLWorker = errors.New("could not init CPU power worker")
	ErrReadRAPL       = errors.New("could not read RAPL energy counter")
)

var (
	_ quartz.Job                  = (*raplWorker)(nil)
	_ workers.PollingEntityWorker = (*raplWorker)(nil)
)

// raplDomain is one RAPL power domain, for example the whole CPU package, its cores or its memory controller.
type raplDomain struct {
	// id is unique to this domain, for example 0 for the first package and 0_1 for its second sub-domain.
	id string
	// name is what the kernel calls the domain, for example package-0, core, uncore or dram.
	name string
	dir  string
	// maxRange is the value at which the energy counter wraps back to zero.
	maxRange uint64
}

// raplReading is a previous energy reading, kept to work out the power drawn since then.
type raplReading struct {
	energy uint64
	at     time.Time
}

// findRAPLDomains lists the RAPL domains in the powercap directory.
func findRAPLDomains(root string) []raplDomain {
	dirs, _ := filepath.Glob(filepath.Join(root, raplDomainGlob)) //nolint:errcheck // only fails on a bad pattern.
	slices.Sort(dirs)

	domains := make([]raplDomain, 0, len(dirs))

	for dir := range slices.Values(dirs) {
		name, err := os.ReadFile(filepath.Join(dir, "name"))
		if err != nil {
			continue
		}

		maxRange, _ := readUint(filepath.Join(dir, "max_energy_range_uj")) //nolint:errcheck // a missing range just disables wrap handling.

		domains = append(domains, raplDomain{
			id:       strings.NewReplacer(":", "_", "-", "_").Replace(strings.TrimPrefix(filepath.Base(dir), "intel-rapl:")),
			name:     strings.TrimSpace(string(name)),
			dir:      dir,
			maxRange: maxRange,
		})
	}

	return domains
}

func readUint(path string) (uint64, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrReadRAPL, err)
	}

	value, err := strconv.ParseUint(strings.TrimSpace(string(contents)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrReadRAPL, err)
	}

	return value, nil
}

// raplPower works out the average power, in W, drawn between two readings of a cumulative energy counter in µJ. A
// counter that went down has wrapped at maxRange. It returns false when no meaningful value can be calculated.
func raplPower(previous, current, maxRange uint64, elapsed time.Duration) (float64, bool) {
	if elapsed <= 0 {
		return 0, false
	}

	delta := current - previous

	if current < previous {
		if maxRange == 0 || maxRange < previous {
			return 0, false
		}

		delta = maxRange - previous + current
	}

	return math.Round(float64(delta)/microjoulesPerJoule/elapsed.Seconds()*100) / 100, true // two decimal places
}

type raplWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	prefs   *RAPLPrefs
	domains []raplDomain
	now     func() time.Time

	mu         sync.Mutex // guards the fields below
	previous   map[string]raplReading
	unreadable bool
}

// NewRAPLWorker creates a worker to generate sensors for the power drawn by the CPU (package, cores, memory controller
// and so on, depending on the CPU), from the RAPL energy counters in /sys/class/powercap. The worker is disabled when
// the machine has no RAPL domains.
func NewRAPLWorker(_ context.Context) (workers.EntityWorker, error) {
	worker := &raplWorker{
		WorkerMetadata:          models.SetWorkerMetadata("cpu_power", "CPU power metrics"),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
		domains:                 findRAPLDomains(filepath.Join(linux.SysFSRoot, raplPowercapDir)),
		now:                     time.Now,
		previous:                make(map[string]raplReading),
	}

	defaultPrefs := &RAPLPrefs{
		UpdateInterval: raplUpdateInterval.String(),
	}

	var err error

	worker.prefs, err = workers.LoadWorkerPreferences(raplPreferencesID, defaultPrefs)
	if err != nil {
		return worker, errors.Join(ErrInitRAPLWorker, err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = raplUpdateInterval
	}

	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, raplUpdateJitter)

	return worker, nil
}

// Execute reads the energy counters and sends a power sensor for each domain. The first run only records the counters,
// as power is the change since the previous reading.
func (w *raplWorker) Execute(ctx context.Context) error {
	sensors := w.readSensors(ctx)

	for sensor := range slices.Values(sensors) {
		w.OutCh <- sensor
	}

	return nil
}

// readSensors holds the lock for the whole read and compare, but not while sending, which can block.
func (w *raplWorker) readSensors(ctx context.Context) []models.Entity {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.unreadable {
		return nil
	}

	sensors := make([]models.Entity, 0, len(w.domains))

	for domain := range slices.Values(w.domains) {
		energy, err := readUint(filepath.Join(domain.dir, "energy_uj"))
		if errors.Is(err, fs.ErrPermission) {
			// Since the fix for CVE-2020-8694 (PLATYPUS) the kernel only lets root read these counters. Say so once.
			slogctx.FromCtx(ctx).Warn("CPU power sensors disabled: the RAPL energy counters are not readable by this user.",
				slog.String("path", domain.dir), slog.Any("error", err))

			w.unreadable = true

			return nil
		}

		if err != nil {
			slogctx.FromCtx(ctx).Debug("Could not read RAPL energy counter.", slog.String("domain", domain.name), slog.Any("error", err))

			continue
		}

		reading := raplReading{energy: energy, at: w.now()}
		previous, found := w.previous[domain.id]
		w.previous[domain.id] = reading

		if !found {
			continue
		}

		if power, ok := raplPower(previous.energy, reading.energy, domain.maxRange, reading.at.Sub(previous.at)); ok {
			sensors = append(sensors, newRAPLSensor(ctx, domain, power))
		}
	}

	return sensors
}

func newRAPLSensor(ctx context.Context, domain raplDomain, power float64) models.Entity {
	return models.NewSensor(ctx,
		models.WithName(fmt.Sprintf("CPU %s Power", domain.name)),
		models.WithID("cpu_rapl_"+domain.id+"_power"),
		models.WithUnits(raplUnits),
		models.WithDeviceClass(models.SensorClassPower),
		models.WithStateClass(models.StateMeasurement),
		models.WithIcon(raplIcon),
		models.WithState(power),
		models.WithDataSourceAttribute(linux.DataSrcSysFS),
		models.WithAttribute("native_unit_of_measurement", raplUnits),
		models.WithAttribute("rapl_domain", domain.name),
	)
}

// IsDisabled is true when disabled in the preferences or when the machine has no RAPL domains.
func (w *raplWorker) IsDisabled() bool {
	return w.prefs.Disabled || len(w.domains) == 0
}

// Start schedules the worker.
func (w *raplWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start CPU power worker: %w", err)
	}

	return w.OutCh, nil
}
