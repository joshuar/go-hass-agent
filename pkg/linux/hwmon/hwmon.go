// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hwmon

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// HWMonPath is the detault path prefix where the hwmon userspace API exists in
// SysFS. Generally, this will never change, but is exposed as a variable to
// ease with writing tests.
var HWMonPath = "/sys/class/hwmon"

//go:generate stringer -type=MonitorType -output hwmon_MonitorType_generated.go
const (
	Unknown MonitorType = iota
	Temp
	Fan
	Voltage
	PWM
	Current
	Power
	Energy
	Humidity
	Frequency
	Alarm
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
	HWMonPath   string
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
	allSensorFiles, err := getSensorFiles(c.HWMonPath)
	if err != nil {
		return []*Sensor{}, fmt.Errorf("could not gather sensor files: %w", err)
	}

	// generate a map of allSensors from the files.
	allSensors := make(map[string]*Sensor)

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	for _, sensorFile := range allSensorFiles {
		wg.Add(1)

		go func() {
			defer wg.Done()

			sensor := newSensor(&sensorFile, c)
			trackerID := sensor.id + "_" + sensor.MonitorType.String()

			mu.Lock()
			defer mu.Unlock()

			if _, ok := allSensors[trackerID]; ok {
				// if this sensor is already tracked, update it from the sensorFile contents.
				if err := allSensors[trackerID].updateInfo(&sensorFile); err != nil {
					slog.Debug("Could not update existing sensor.",
						slog.String("sensor", allSensors[trackerID].Name()),
						slog.Any("error", err))
				}
			} else {
				// else update and add as new.
				if err := sensor.updateInfo(&sensorFile); err != nil {
					slog.Debug("Could not parse sensor.",
						slog.String("sensor", allSensors[trackerID].Name()),
						slog.Any("error", err))
				} else {
					allSensors[trackerID] = sensor
				}
			}
		}()
	}

	wg.Wait()

	sensors := make([]*Sensor, 0, len(allSensors))

	for _, sensor := range allSensors {
		if err := sensor.updateValue(); err != nil {
			slog.Debug("Could not update sensor value.",
				slog.String("sensor", sensor.Name()),
				slog.Any("error", err))
		} else {
			if sensor.value == nil {
				slog.Info("value is nil", slog.String("sensor", sensor.Name()))
			}

			sensors = append(sensors, sensor)
		}
	}

	return sensors, nil
}

// newChip creates a new chip from the given hwmon sysfs path.
func newChip(path string) (*Chip, error) {
	chipName, err := getFileContents(filepath.Join(path, "name"))
	if err != nil {
		return nil, err
	}

	chip := &Chip{
		chipName:  chipName,
		chipID:    filepath.Base(path),
		HWMonPath: path,
	}

	fh, err := os.Stat(filepath.Join(path, "device", "model"))
	if err == nil && fh.Mode().IsRegular() {
		chip.deviceModel, err = getFileContents(filepath.Join(path, "device", "model"))
		if err == nil {
			slog.Debug("Could not retrieve a device model for chip.",
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
func GetAllChips() ([]*Chip, error) {
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
			wg.Add(1)

			go func() {
				defer wg.Done()

				if chip, err := newChip(filepath.Join(HWMonPath, file.Name())); err != nil {
					slog.Debug("Could not process hwmon path.",
						slog.String("path", file.Name()),
						slog.Any("error", err))
				} else {
					chipCh <- chip
				}
			}()
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

	if s.MonitorType == Alarm || s.MonitorType == Intrusion {
		id.WriteString("_")
		id.WriteString(s.MonitorType.String())
	}

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
		s.Name(), s.Value(), s.Units(), s.MonitorType, s.ID(), s.HWMonPath, s.Chip.String())

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

// updateValue will update the value of the sensor to its current value, as read
// from its input file in the hwmon sysfs tree.
func (s *Sensor) updateValue() error {
	var (
		path  string
		value any
		err   error
	)

	switch s.MonitorType {
	case Alarm:
		path = filepath.Join(s.HWMonPath, s.id+"_alarm")

		value, err = getValueAsBool(path)
		if err != nil {
			return fmt.Errorf("unable to read value: %w", err)
		}
	case Intrusion:
		path = filepath.Join(s.HWMonPath, s.id+"_intrusion")

		value, err = getValueAsBool(path)
		if err != nil {
			return fmt.Errorf("unable to read value: %w", err)
		}
	default:
		path = filepath.Join(s.HWMonPath, s.id+"_input")

		var floatValue float64

		floatValue, err = getValueAsFloat(path)
		if err != nil {
			return fmt.Errorf("unable to read value: %w", err)
		}

		value = floatValue / s.scaleFactor
	}

	s.value = value

	return nil
}

// updateInfo will add any additional info from the given sensorFile to the
// sensor. This function is called in a loop processing files for a chip from
// the hwmon sysfs, and will gradually build all the details of the sensor as
// relevant.
func (s *Sensor) updateInfo(file *sensorFile) error {
	path := filepath.Join(file.path, file.filename)

	switch {
	case file.sensorAttr == "input":
	case file.sensorAttr == "label":
		l, err := getValueAsString(path)
		if err != nil {
			return err
		}

		s.label = l
	case strings.Contains(file.sensorAttr, "alarm"):
		if b, _, ok := strings.Cut(file.sensorAttr, "_"); ok {
			s.label = file.sensorType + " " + b + " Alarm"
			s.id += "_" + b
		} else {
			s.label = "Alarm"
		}
	case strings.Contains(file.sensorAttr, "intrusion"):
		s.label = "intrusion"
	default:
		v, err := getValueAsFloat(path)
		if err != nil {
			return err
		}

		s.Attributes = append(s.Attributes, Attribute{Name: file.sensorAttr, Value: v / s.scaleFactor})
	}

	return nil
}

// newSensor creates a new sensor representation from the given sensorFile.
func newSensor(file *sensorFile, chip *Chip) *Sensor {
	sensorType, scaleFactor, sensorUnits := file.getSensorType()

	return &Sensor{
		Chip:        chip,
		id:          file.sensorType,
		MonitorType: sensorType,
		scaleFactor: scaleFactor,
		units:       sensorUnits,
	}
}

// GetAllSensors returns a slice of Sensor objects, representing all detected
// chip sensors found on the host. If there were any errors in fetching chips or
// chip sensors, it will also return a non-nill composite error.
func GetAllSensors() ([]*Sensor, error) {
	var sensors []*Sensor

	chips, err := GetAllChips()
	for _, chip := range chips {
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
	filename   string
	sensorType string
	sensorAttr string
}

//nolint:cyclop,mnd
func (f *sensorFile) getSensorType() (sensorType MonitorType, scaleFactor float64, units string) {
	switch {
	case strings.Contains(f.sensorAttr, "intrusion"):
		return Intrusion, 1, ""
	case strings.Contains(f.sensorAttr, "alarm"):
		return Alarm, 1, ""
	case strings.Contains(f.sensorType, "temp"):
		return Temp, 1000, "°C"
	case strings.Contains(f.sensorType, "fan"):
		return Fan, 1, "rpm"
	case strings.Contains(f.sensorType, "in"):
		return Voltage, 1000, "V"
	case strings.Contains(f.sensorType, "pwm"):
		return PWM, 1, "Hz"
	case strings.Contains(f.sensorType, "curr"):
		return Current, 1, "A"
	case strings.Contains(f.sensorType, "power"):
		return Power, 1000, "W"
	case strings.Contains(f.sensorType, "energy"):
		return Energy, 1000, "J"
	case strings.Contains(f.sensorType, "humidity"):
		return Humidity, 1, "%"
	case strings.Contains(f.sensorType, "freq"):
		return Frequency, 1000, "Hz"
	default:
		return Unknown, 1, ""
	}
}

// getFileContents retrieves the contents of the given file as a string. If the
// contents cannot be read, it will return "unknown" and an error.
func getFileContents(path string) (string, error) {
	var (
		data []byte
		err  error
	)

	if data, err = os.ReadFile(path); err != nil {
		return "", fmt.Errorf("could not read contents of file: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func getValueAsString(p string) (string, error) {
	return getFileContents(p)
}

func getValueAsFloat(p string) (float64, error) {
	strValue, err := getFileContents(p)
	if err != nil {
		return 0, err
	}

	var floatValue float64

	floatValue, err = strconv.ParseFloat(strValue, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse as float: %w", err)
	}

	return floatValue, nil
}

func getValueAsBool(p string) (bool, error) {
	strValue, err := getFileContents(p)
	if err != nil {
		return false, err
	}

	var boolVal bool

	boolVal, err = strconv.ParseBool(strValue)
	if err != nil {
		return false, fmt.Errorf("could not parse as bool: %w", err)
	}

	return boolVal, nil
}

func getSensorFiles(hwMonPath string) ([]sensorFile, error) {
	fileList, err := os.ReadDir(hwMonPath)
	if err != nil {
		return nil, fmt.Errorf("could not read files at path %s: %w", hwMonPath, err)
	}

	files := make([]sensorFile, 0, len(fileList))
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
				sensorType, sensorAttr, ok := strings.Cut(file.Name(), "_")
				if !ok {
					return
				}
				// return any other file as a *sensorFile
				fileCh <- sensorFile{
					path:       hwMonPath,
					filename:   file.Name(),
					sensorType: sensorType,
					sensorAttr: sensorAttr,
				}
			}()
		}

		wg.Wait()
		close(fileCh)
	}()

	for sensorFile := range fileCh {
		files = append(files, sensorFile)
	}

	return files, nil
}
