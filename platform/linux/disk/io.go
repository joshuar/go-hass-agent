// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go tool stringer -type=ioSensor,stat -output io.gen.go -linecomment
package disk

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/reugn/go-quartz/quartz"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/class"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/platform/linux"
	"github.com/joshuar/go-hass-agent/scheduler"
)

const (
	ioWorkerUpdateInterval = 5 * time.Second
	ioWorkerUpdateJitter   = time.Second

	ioWorkerID   = "disk_rates_sensors"
	ioWorkerDesc = "IO usage stats"

	totalsID = "total"
)

const (
	TotalReads             stat = iota // Total reads completed
	TotalReadsMerged                   // Total reads merged
	TotalSectorsRead                   // Total sectors read
	TotalTimeReading                   // Total milliseconds spent reading
	TotalWrites                        // Total writes completed
	TotalWritesMerged                  // Total writes merged
	TotalSectorsWritten                // Total sectors written
	TotalTimeWriting                   // Total milliseconds spent writing
	ActiveIOs                          // I/Os currently in progress
	ActiveIOTime                       // Milliseconds elapsed spent doing I/Os
	ActiveIOTimeWeighted               // Milliseconds elapsed spent doing I/Os (weighted)
	TotalDiscardsCompleted             // Total discards completed
	TotalDiscardsMerged                // Total discards merged
	TotalSectorsDiscarded              // Total sectors discarded
	TotalTimeDiscarding                // Total milliseconds spent discarding
	TotalFlushRequests                 // Total flush requests completed
	TotalTimeFlushing                  // Total milliseconds spent flushing
)

// stat represents a specific statistic recorded by the kernel for the
// associated disk.
type stat int

const (
	diskReads        ioSensor = iota // Disk Reads
	diskWrites                       // Disk Writes
	diskReadRate                     // Disk Read Rate
	diskWriteRate                    // Disk Write Rate
	diskIOInProgress                 // Disk IOs In Progress
)

// ioSensor represents a type of sensor being collected.
type ioSensor int

const (
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
	ErrParseDevices      = errors.New("could not parse devices")
	ErrInitRatesWorker   = errors.New("could not init rates worker")
)

var (
	deviceMajNo = []string{"8", "252", "253", "259"}
)

var (
	_ quartz.Job                  = (*ioWorker)(nil)
	_ workers.PollingEntityWorker = (*ioWorker)(nil)
)

type ioRate struct {
	linux.RateValue[uint64]

	rateType ioSensor
}

func newDiskStatSensor(
	ctx context.Context,
	device *device,
	sensorType ioSensor,
	value uint64,
	attributes models.Attributes,
) models.Entity {
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

type device struct {
	id        string
	sysFSPath string
	model     string
}

func (d *device) generateIdentifiers(sensorType ioSensor) (string, string) {
	r := []rune(d.id)
	name := string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...)) + " " + sensorType.String()
	id := strings.ToLower(d.id + "_" + strings.ReplaceAll(sensorType.String(), " ", "_"))

	return name, id
}

func (d *device) generateAttributes() models.Attributes {
	attributes := models.Attributes{
		"data_source": linux.DataSrcSysFS,
	}

	// Add attributes from device if available.
	if d.model != "" {
		attributes["device_model"] = d.model
	}

	if d.sysFSPath != "" {
		attributes["sysfs_path"] = d.sysFSPath
	}

	return attributes
}

func getDeviceNames() ([]string, error) {
	data, err := os.Open(filepath.Join(linux.ProcFSRoot, "partitions"))
	if err != nil {
		return nil, fmt.Errorf("getDevices: %w", err)
	}

	defer data.Close()

	var devices []string

	partitions := bufio.NewScanner(data)
	// Skip first two lines (header + blank line).
	for range 2 {
		partitions.Scan()
	}
	// Read remaining lines.
	for partitions.Scan() {
		line := bufio.NewScanner(bytes.NewReader(partitions.Bytes()))
		line.Split(bufio.ScanWords)

		var cols []string

		for line.Scan() {
			cols = append(cols, line.Text())
		}

		if len(cols) == 0 {
			return devices, ErrParseDevices
		}

		if validDeviceNo(cols) {
			devices = append(devices, cols[3])
		}
	}

	return devices, nil
}

func validDeviceNo(details []string) bool {
	if slices.Contains(deviceMajNo, details[0]) {
		if details[1] == "0" {
			return true
		}
	}

	return false
}

func getDevice(deviceName string) (*device, map[stat]uint64, error) {
	// Create a new device.
	dev := &device{
		id:        deviceName,
		sysFSPath: filepath.Join(linux.SysFSRoot, "block", deviceName),
	}

	// Try to read the model from the appropriate file. Otherwise just leave
	// it empty.
	if model, err := os.ReadFile(dev.sysFSPath + "/device/model"); err == nil {
		dev.model = strings.TrimSpace(string(model))
	}

	data, err := os.ReadFile(dev.sysFSPath + "/stat")
	if err != nil {
		return dev, nil, fmt.Errorf("getDeviceStats: %w", err)
	}

	line := bufio.NewScanner(bytes.NewReader(data))
	line.Split(bufio.ScanWords)

	stats := make(map[stat]uint64)
	statno := stat(0)
	// Parse the rest as stats.
	for line.Scan() {
		readVal, err := strconv.ParseUint(line.Text(), 10, 64)
		if err != nil {
			slog.Warn("Unable to parse device stat.",
				slog.String("device", dev.id),
				slog.String("stat", line.Text()),
				slog.Any("error", err))
		} else {
			stats[statno] = readVal
		}

		statno++
	}

	return dev, stats, nil
}

// ioWorker creates sensors for disk IO counts and rates per device. It
// maintains an internal map of devices being tracked.
type ioWorker struct {
	*models.WorkerMetadata
	*workers.PollingEntityWorkerData

	boottime    time.Time
	rateSensors map[string]map[ioSensor]*ioRate
	mu          sync.Mutex
	prefs       *WorkerPrefs
}

func NewIOWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &ioWorker{
		WorkerMetadata:          models.SetWorkerMetadata(ioWorkerID, ioWorkerDesc),
		PollingEntityWorkerData: &workers.PollingEntityWorkerData{},
	}

	var found bool

	worker.boottime, found = linux.CtxGetBoottime(ctx)
	if !found {
		return worker, errors.Join(ErrInitRatesWorker,
			fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx))
	}

	// Add sensors for a pseudo "total" device which tracks total values from
	// all sensors.
	sensors := make(map[string]map[ioSensor]*ioRate)
	sensors["total"] = map[ioSensor]*ioRate{
		diskReadRate:  {rateType: diskReadRate},
		diskWriteRate: {rateType: diskWriteRate},
	}
	worker.rateSensors = sensors

	defaultPrefs := &WorkerPrefs{
		UpdateInterval: ioWorkerUpdateInterval.String(),
	}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(ioWorkerPreferencesID, defaultPrefs)
	if err != nil {
		return worker, errors.Join(ErrInitRatesWorker, err)
	}

	pollInterval, err := time.ParseDuration(worker.prefs.UpdateInterval)
	if err != nil {
		pollInterval = ioWorkerUpdateInterval
	}
	worker.Trigger = scheduler.NewPollTriggerWithJitter(pollInterval, ioWorkerUpdateJitter)

	return worker, nil
}

func (w *ioWorker) Execute(ctx context.Context) error {
	delta := w.GetDelta()
	// Get valid devices.
	deviceNames, err := getDeviceNames()
	if err != nil {
		return fmt.Errorf("could not fetch disk devices: %w", err)
	}

	statsTotals := make(map[stat]uint64)

	// Get the current device info and stats for all valid devices.
	for dev := range slices.Values(deviceNames) {
		dev, stats, err := getDevice(dev)
		if err != nil {
			slogctx.FromCtx(ctx).
				With(slog.String("worker", ioWorkerID)).
				Debug("Unable to read device stats.", slog.Any("error", err))

			continue
		}

		// Add rate sensors for device (if not already added).
		w.addRateSensors(dev)
		for s := range slices.Values(w.generateDeviceRateSensors(ctx, dev, stats, delta)) {
			w.OutCh <- s
		}

		for s := range slices.Values(w.generateDeviceStatSensors(ctx, dev, stats)) {
			w.OutCh <- s
		}

		// Don't include "aggregate" devices in totals.
		if strings.HasPrefix(dev.id, "dm") || strings.HasPrefix(dev.id, "md") {
			continue
		}
		// Add device stats to the totals.
		for stat, value := range stats {
			statsTotals[stat] += value
		}
	}

	// Update total stats.
	for s := range slices.Values(w.generateDeviceRateSensors(ctx, &device{id: totalsID}, statsTotals, delta)) {
		w.OutCh <- s
	}
	for s := range slices.Values(w.generateDeviceStatSensors(ctx, &device{id: totalsID}, statsTotals)) {
		w.OutCh <- s
	}

	return nil
}

func (w *ioWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *ioWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	w.OutCh = make(chan models.Entity)
	if err := workers.SchedulePollingWorker(ctx, w, w.OutCh); err != nil {
		close(w.OutCh)
		return w.OutCh, fmt.Errorf("could not start disk IO worker: %w", err)
	}
	return w.OutCh, nil
}

// addDevice adds a new device to the tracker map. If sthe device is already
// being tracked, it will not be added again. The bool return indicates whether
// a device was added (true) or not (false).
func (w *ioWorker) addRateSensors(dev *device) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, found := w.rateSensors[dev.id]; !found {
		w.rateSensors[dev.id] = map[ioSensor]*ioRate{
			diskReadRate:  {rateType: diskReadRate},
			diskWriteRate: {rateType: diskWriteRate},
		}
	}
}

func (w *ioWorker) generateDeviceRateSensors(
	ctx context.Context,
	device *device,
	stats map[stat]uint64,
	delta time.Duration,
) []models.Entity {
	var sensors []models.Entity

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, found := w.rateSensors[device.id]; found && stats != nil {
		for rateType := range w.rateSensors[device.id] {
			var currValue uint64

			switch rateType {
			case diskReadRate:
				currValue = stats[TotalSectorsRead]
			case diskWriteRate:
				currValue = stats[TotalSectorsWritten]
			}

			rate := w.rateSensors[device.id][rateType].Calculate(currValue, delta)
			sensors = append(sensors, newDiskRateSensor(ctx, device, rateType, rate))
		}
	}

	return sensors
}

func (w *ioWorker) generateDeviceStatSensors(
	ctx context.Context,
	device *device,
	stats map[stat]uint64,
) []models.Entity {
	var sensors []models.Entity

	diskReadsAttributes := models.Attributes{
		"total_sectors_read":         stats[TotalSectorsRead],
		"total_milliseconds_reading": stats[TotalTimeReading],
	}

	diskWriteAttributes := models.Attributes{
		"total_sectors_written":      stats[TotalSectorsWritten],
		"total_milliseconds_writing": stats[TotalTimeWriting],
	}

	// Generate diskReads sensor for device.
	sensors = append(sensors, newDiskStatSensor(ctx, device, diskReads, stats[TotalReads], diskReadsAttributes))
	// Generate diskWrites sensor for device.
	sensors = append(sensors, newDiskStatSensor(ctx, device, diskWrites, stats[TotalWrites], diskWriteAttributes))
	// Generate IOsInProgress sensor for device.
	sensors = append(sensors, newDiskStatSensor(ctx, device, diskIOInProgress, stats[ActiveIOs], nil))

	return sensors
}
