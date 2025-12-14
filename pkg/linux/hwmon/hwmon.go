// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package hwmon contains methods for accessing hwmon sensor values from the Linux kernel.
package hwmon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	slogctx "github.com/veqryn/slog-context"
	"golang.org/x/sys/unix"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// HWMonPath is the detault path prefix where the hwmon userspace API exists in
// SysFS. Generally, this will never change, but is exposed as a variable to
// ease with writing tests.
var HWMonPath = "/sys/class/hwmon"

//go:generate go tool stringer -type=MonitorType -output hwmon_MonitorType_generated.go
const (
	// Unknown hwmon sensor.
	Unknown MonitorType = iota
	// Temp hwmon sensor.
	Temp
	// Fan hwmon sensor.
	Fan
	// Voltage hwmon sensor.
	Voltage
	// PWM hwmon sensor.
	PWM
	// Current hwmon sensor.
	Current
	// Power hwmon sensor.
	Power
	// Energy hwmon sensor.
	Energy
	// Humidity hwmon sensor.
	Humidity
	// Frequency hwmon sensor.
	Frequency
	// Alarm hwmon sensor.
	Alarm
	// Intrusion hwmon sensor.
	Intrusion
)

// MonitorType represents the type of sensor. For example, a temp sensor, a fan
// sensor, etc.
type MonitorType int

// Chip represents a sensor chip exposed by the Linux kernel hardware monitoring
// API. These are retrieved from the directories in the sysfs /sys/devices tree
// under /sys/class/hwmon/hwmon*.
type Chip struct {
	chipName    string
	chipID      string
	deviceModel string
	Path        string
	Sensors     []*Sensor
}

// Chip returns a formatted string for identifying the chip to which this sensor
// belongs.
func (c *Chip) String() string {
	if c.deviceModel != "" {
		return c.deviceModel
	}

	if c.chipName != "" {
		return c.chipName
	}

	return c.chipID
}

// getSensors retrieves all the sensors for the chip from hwmon sysfs. It
// returns a slice of the sensors. If it cannot read the hwmon sysfs path, a
// non-nil error is returned with details. It will not return an error if there
// was an error retrieving an individual sensor for the chip.
func (c *Chip) getSensors() ([]*Sensor, error) {
	allSensorFiles, err := getSensorFiles(c.Path)
	if err != nil {
		return []*Sensor{}, fmt.Errorf("could not gather sensor files: %w", err)
	}

	// generate a map of allSensors from the files.
	allSensors := make(map[string]*Sensor)

	for file := range allSensorFiles {
		trackerID := file.sensorID + "_" + file.sensorType.String()

		if _, ok := allSensors[trackerID]; !ok {
			allSensors[trackerID] = newSensor(&file, c)
		}
		// Update based on the file contents
		if err := allSensors[trackerID].updateInfo(&file); err != nil {
			slog.Debug("Could not update sensor.",
				slog.String("sensor", allSensors[trackerID].Name()),
				slog.Any("error", err))
		}
	}

	sensors := make([]*Sensor, 0, len(allSensors))

	for _, sensor := range allSensors {
		if sensor.value == nil {
			slog.Debug("Ignoring sensor with nil value.", slog.String("sensor", sensor.Name()))
			continue
		}

		sensors = append(sensors, sensor)
	}

	return sensors, nil
}

func getSensorFiles(hwMonPath string) (chan sensorFile, error) {
	fileList, err := os.ReadDir(hwMonPath)
	if err != nil {
		return nil, fmt.Errorf("could not read files at path %s: %w", hwMonPath, err)
	}

	fileCh := make(chan sensorFile)

	go func() {
		var wg sync.WaitGroup

		for _, file := range fileList {
			wg.Add(1)

			go func() {
				defer wg.Done()
				// ignore directories
				if file.IsDir() {
					return
				}
				// ignore files that can't be parsed as a sensor
				id, attr, ok := strings.Cut(file.Name(), "_")
				if !ok {
					return
				}
				// adjust id for alarms.
				if strings.Contains(attr, "alarm") {
					id += "_alarm"
				}
				// get and store the contents of the sensor file.
				contents, err := getFileContents(filepath.Join(hwMonPath, file.Name()))
				if err != nil {
					return
				}
				// return as a sensorFile.
				fileCh <- sensorFile{
					path:       hwMonPath,
					name:       file.Name(),
					sensorID:   id,
					attribute:  attr,
					sensorType: parseSensorType(id),
					contents:   contents,
				}
			}()
		}

		wg.Wait()
		close(fileCh)
	}()

	return fileCh, nil
}

// newChip creates a new chip from the given hwmon sysfs path.
func newChip(ctx context.Context, path string) (*Chip, error) {
	chipName, err := getFileContents(filepath.Join(path, "name"))
	if err != nil {
		return nil, err
	}

	chip := &Chip{
		chipName: chipName,
		chipID:   filepath.Base(path),
		Path:     path,
	}

	fh, err := os.Stat(filepath.Join(path, "device", "model"))
	if err == nil && fh.Mode().IsRegular() {
		chip.deviceModel, err = getFileContents(filepath.Join(path, "device", "model"))
		if err == nil {
			slogctx.FromCtx(ctx).Debug("Could not retrieve a device model for chip.",
				slog.String("chip", chip.chipName),
				slog.Any("error", err))
		}
	}

	sensors, err := chip.getSensors()
	chip.Sensors = sensors

	return chip, err
}

// GetAllChips will return a slice of Chips containing their sensors. If there
// are any errors in parsing chip or sensor values, it will return a non-nill
// composite error as well.
func GetAllChips(ctx context.Context) ([]*Chip, error) {
	// Get all the hwmon chips.
	files, err := os.ReadDir(HWMonPath)
	if err != nil {
		return nil, fmt.Errorf("could not read hwmon data at path %s: %w", HWMonPath, err)
	}

	chips := make([]*Chip, 0, len(files))
	chipCh := make(chan *Chip)

	// Spawn a goroutine for each chip to get its details and retrieve its sensors.
	go func() {
		defer close(chipCh)

		var wg sync.WaitGroup

		for _, file := range files {
			wg.Go(func() {
				if chip, err := newChip(ctx, filepath.Join(HWMonPath, file.Name())); err != nil {
					slogctx.FromCtx(ctx).Debug("Could not process hwmon path.",
						slog.String("path", file.Name()),
						slog.Any("error", err))
				} else {
					chipCh <- chip
				}
			})
		}

		wg.Wait()
	}()

	// Collect all valid chips.
	for chip := range chipCh {
		chips = append(chips, chip)
	}

	return chips, nil
}

// Sensor represents a single sensor exposed by a sensor chip. A Sensor may have
// a label, which is a formatted name of the sensor, otherwise it will just have
// a name. The Sensor will also have a value. It may also have zero or more
// Attributes, which are additional measurements like max/min/avg of the value.
type Sensor struct {
	*Chip
	value any
	label string
	id    string
	units string
	// Attributes is a slice of additional attributes, such as max, min, crit
	// values for the sensor.
	Attributes  []Attribute
	scaleFactor float64
	MonitorType MonitorType
}

// Value returns the current value of the sensor.
func (s *Sensor) Value() any {
	return s.value
}

// Name returns a formatted string as the name for the sensor. It will be
// derived from the chip name plus either any label, else name of the sensor
// itself.
func (s *Sensor) Name() string {
	var name strings.Builder

	capitaliser := cases.Title(language.English)

	if s.deviceModel != "" {
		name.WriteString(s.deviceModel)
	} else {
		name.WriteString("Hardware Sensor")

		if s.chipName != "" {
			name.WriteString(" ")
			name.WriteString(capitaliser.String(strings.ReplaceAll(s.chipName, "_", " ")))
		}
	}

	name.WriteString(" ")

	if s.MonitorType == Alarm || s.MonitorType == Intrusion {
		if !strings.Contains(s.id, "_") {
			name.WriteString(capitaliser.String(s.id))
			name.WriteString(" ")
		}

		name.WriteString(capitaliser.String(s.label))
	} else {
		if s.label != "" {
			name.WriteString(capitaliser.String(s.label))
		} else {
			name.WriteString(capitaliser.String(s.id))
		}
	}

	return name.String()
}

// ID returns a formatted string that can be used as a unique identifier for
// this sensor. This will be some combination of the chip and sensor details, as
// appropriate.
func (s *Sensor) ID() string {
	var id strings.Builder

	id.WriteString(s.chipID)
	id.WriteString("_")
	id.WriteString(s.chipName)
	id.WriteString("_")
	id.WriteString(s.id)

	return strcase.ToSnake(id.String())
}

// Units returns the units for the value of this sensor.
func (s *Sensor) Units() string {
	return s.units
}

// String will format the sensor name and value as a pretty string.
func (s *Sensor) String() string {
	var sensorStr strings.Builder

	fmt.Fprintf(&sensorStr,
		"%s: %v %s [%s] (id: %s, path: %s, chip: %s)",
		s.Name(), s.Value(), s.Units(), s.MonitorType, s.ID(), s.Path, s.Chip.String())

	for idx, a := range s.Attributes {
		if idx == 0 {
			fmt.Fprintf(&sensorStr, " (")
		}

		sensorStr.WriteString(a.String())

		if idx < len(s.Attributes)-1 {
			fmt.Fprintf(&sensorStr, ", ")
		}

		if idx == len(s.Attributes)-1 {
			fmt.Fprintf(&sensorStr, ")")
		}
	}

	return sensorStr.String()
}

// updateInfo will add any additional info from the given sensorFile to the
// sensor. This function is called in a loop processing files for a chip from
// the hwmon sysfs, and will gradually build all the details of the sensor as
// relevant.
func (s *Sensor) updateInfo(file *sensorFile) error {
	switch {
	case file.attribute == "label":
		s.label = file.contents
	case strings.Contains(file.attribute, "alarm"):
		id, _, _ := strings.Cut(file.sensorID, "_")
		parts := strings.Split(file.attribute, "_")

		if len(parts) == 2 { // 2 parts, limit alarm
			s.label = strings.Join([]string{id, parts[0], file.sensorType.String()}, " ")
			s.id = strings.Join([]string{id, parts[0], "alarm"}, "_")
		} else { // channel alarm
			s.label = strings.Join([]string{id, file.sensorType.String()}, " ")
		}

		if value, err := strconv.ParseBool(file.contents); err != nil {
			return fmt.Errorf("could not parse as bool: %w", err)
		} else {
			s.value = value
		}
	case strings.Contains(file.attribute, "intrusion"):
		id, _, _ := strings.Cut(file.sensorID, "_")

		s.label = strings.Join([]string{id, file.sensorType.String()}, " ")

		value, err := strconv.ParseBool(file.contents)
		if err != nil {
			return fmt.Errorf("could not parse as bool: %w", err)
		}
		s.value = value

	default: // Either the sensor value or an attribute of the sensor.
		value, err := strconv.ParseFloat(file.contents, 64)
		if err != nil {
			return fmt.Errorf("could not parse as float: %w", err)
		}

		if file.attribute == "input" {
			s.value = value / s.scaleFactor
		} else {
			s.Attributes = append(s.Attributes, Attribute{Name: file.attribute, Value: value / s.scaleFactor})
		}
	}

	return nil
}

// newSensor creates a new sensor representation from the given sensorFile.
func newSensor(file *sensorFile, chip *Chip) *Sensor {
	return &Sensor{
		Chip:        chip,
		id:          file.sensorID,
		MonitorType: file.sensorType,
		scaleFactor: getScaleFactor(file.sensorType),
		units:       getUnits(file.sensorType),
	}
}

// GetAllSensors returns a slice of Sensor objects, representing all detected
// chip sensors found on the host. If there were any errors in fetching chips or
// chip sensors, it will also return a non-nill composite error.
func GetAllSensors(ctx context.Context) ([]*Sensor, error) {
	chips, err := GetAllChips(ctx)
	sensors := make([]*Sensor, 0, len(chips))
	for chip := range slices.Values(chips) {
		sensors = append(sensors, chip.Sensors...)
	}
	return sensors, err
}

// Attribute represents an attribute of a sensor, like its max, min or average
// value.
type Attribute struct {
	Name  string
	Value float64
}

// String will format the attribute name and value as a pretty string.
func (a *Attribute) String() string {
	return fmt.Sprintf("%s: %.3f", a.Name, a.Value)
}

type sensorFile struct {
	path       string
	name       string
	sensorID   string
	attribute  string
	contents   string
	sensorType MonitorType
}

func getScaleFactor(sensorType MonitorType) float64 {
	switch sensorType {
	case Intrusion, Alarm, Fan, PWM, Current, Humidity:
		return 1
	case Temp, Voltage, Power, Energy, Frequency:
		return 1000
	default:
		return 1
	}
}

func getUnits(sensorType MonitorType) string {
	switch sensorType {
	case Temp:
		return "°C"
	case Fan:
		return "rpm"
	case Voltage:
		return "V"
	case PWM, Frequency:
		return "Hz"
	case Current:
		return "A"
	case Power:
		return "W"
	case Energy:
		return "J"
	case Humidity:
		return "%"
	default:
		return ""
	}
}

func parseSensorType(id string) MonitorType {
	switch {
	case strings.Contains(id, "intrusion"):
		return Intrusion
	case strings.Contains(id, "alarm"):
		return Alarm
	case strings.Contains(id, "temp"):
		return Temp
	case strings.Contains(id, "fan"):
		return Fan
	case strings.Contains(id, "in"):
		return Voltage
	case strings.Contains(id, "pwm"):
		return PWM
	case strings.Contains(id, "curr"):
		return Current
	case strings.Contains(id, "power"):
		return Power
	case strings.Contains(id, "energy"):
		return Energy
	case strings.Contains(id, "humidity"):
		return Humidity
	case strings.Contains(id, "freq"):
		return Frequency
	default:
		return Unknown
	}
}

// Adapted from:
// https://github.com/prometheus/node_exporter/blob/master/collector/hwmon_linux.go
func getFileContents(file string) (string, error) {
	handle, err := os.Open(file) // #nosec: G304
	if err != nil {
		return "", fmt.Errorf("could not open file: %w", err)
	}
	defer handle.Close() //nolint:errcheck

	// On some machines, hwmon drivers are broken and return EAGAIN.  This causes
	// Go's os.ReadFile implementation to poll forever.
	//
	// Since we either want to read data or bail immediately, do the simplest
	// possible read using system call directly.
	bufPtr := hwmonBufPool.Get().(*[]byte)
	data := *bufPtr
	defer hwmonBufPool.Put(bufPtr)

	n, err := unix.Read(int(handle.Fd()), data) // #nosec G115 // I do not believe this is a problem.
	if err != nil {
		return "", fmt.Errorf("could not read contents of file: %w", err)
	}

	return strings.TrimSpace(string(data[:n])), nil
}

var hwmonBufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 128)
		return &buf
	},
}
