// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hwmon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:generate stringer -type=SensorType -output sensorType_strings.go
const (
	hwmonPath            = "/sys/class/hwmon"
	Unknown   SensorType = iota
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

// SensorType represents the type of sensor. For example, a temp sensor, a fan
// sensor, etc.
type SensorType int

// Chip represents a sensor chip exposed by the Linux kernel hardware monitoring
// API. These are retrieved from the directories in the sysfs /sys/devices tree
// under /sys/class/hwmon/hwmon*.
type Chip struct {
	Name    string
	id      string
	Sensors []*Sensor
}

func processChip(path string) (*Chip, error) {
	n, err := getFileContents(filepath.Join(path, "name"))
	if err != nil {
		return nil, err
	}

	c := &Chip{
		Name: n,
		id:   filepath.Base(path),
	}

	sensors, err := getSensors(path)
	c.Sensors = sensors

	return c, err
}

// GetAllChips will return a slice of Chips containing their sensors. If there
// are any errors in parsing chip or sensor values, it will return a non-nill
// composite error as well.
func GetAllChips() ([]*Chip, error) {
	var chips []*Chip

	files, err := os.ReadDir(hwmonPath)
	if err != nil {
		return nil, err
	}

	p := pool.New().WithErrors()
	for _, f := range files {
		p.Go(func() error {
			chip, err := processChip(filepath.Join(hwmonPath, f.Name()))
			chips = append(chips, chip)
			return err
		})
	}
	err = p.Wait()
	return chips, err
}

// Sensor represents a single sensor exposed by a sensor chip. A Sensor may have
// a label, which is a formatted name of the sensor, otherwise it will just have
// a name. The Sensor will also have a value. It may also have zero or more
// Attributes, which are additional measurements like max/min/avg of the value.
type Sensor struct {
	chipLabel   string
	chipID      string
	deviceModel string
	label       string
	id          string
	units       string
	// SysFSPath is the base path under the hwmon tree in /sys that contains this sensor.
	SysFSPath string
	// Attributes is a slice of additional attributes, such as max, min, crit
	// values for the sensor.
	Attributes  []Attribute
	scaleFactor float64
	SensorType  SensorType
}

// Value returns the sensor value. This will be either a bool for alarm and
// intrusion sensors, or a float64 for all other types of sensors.
func (s *Sensor) Value() any {
	var path string
	switch s.SensorType {
	case Alarm:
		path = filepath.Join(s.SysFSPath, s.id+"_alarm")
		value, err := getValueAsBool(path)
		if err != nil {
			log.Debug().Err(err).Str("sensor", s.Name()).Msg("Problem fetching sensor value.")
			return nil
		}
		return value
	case Intrusion:
		path = filepath.Join(s.SysFSPath, s.id+"_intrusion")
		value, err := getValueAsBool(path)
		if err != nil {
			log.Debug().Err(err).Str("sensor", s.Name()).Msg("Problem fetching sensor value.")
			return nil
		}
		return value
	default:
		path = filepath.Join(s.SysFSPath, s.id+"_input")
		value, err := getValueAsFloat(path)
		if err != nil {
			log.Debug().Err(err).Str("sensor", s.Name()).Msg("Problem fetching sensor value.")
			return 0.0
		}
		return value / s.scaleFactor
	}
}

// Name returns a formatted string as the name for the sensor. It will be
// derived from the chip name plus either any label, else name of the sensor
// itself.
func (s *Sensor) Name() string {
	capitaliser := cases.Title(language.English)
	var name strings.Builder
	if s.deviceModel != "" {
		name.WriteString(s.deviceModel)
	} else {
		name.WriteString("Hardware Sensor")
		if s.chipLabel != "" {
			name.WriteString(" ")
			name.WriteString(capitaliser.String(strings.ReplaceAll(s.chipLabel, "_", " ")))
		}
	}
	name.WriteString(" ")
	if s.SensorType == Alarm || s.SensorType == Intrusion {
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

// Chip returns a formatted string for identifying the chip to which this sensor
// belongs.
func (s *Sensor) Chip() string {
	if s.deviceModel != "" {
		return s.deviceModel
	}
	if s.chipLabel != "" {
		return s.chipLabel
	}
	return s.chipID
}

// ID returns a formatted string that can be used as a unique identifier for
// this sensor. This will be some combination of the chip and sensor details, as
// appropriate.
func (s *Sensor) ID() string {
	var id strings.Builder
	id.WriteString(s.chipID)
	id.WriteString("_")
	id.WriteString(s.chipLabel)
	id.WriteString("_")
	id.WriteString(s.id)
	if s.SensorType == Alarm || s.SensorType == Intrusion {
		id.WriteString("_")
		id.WriteString(s.SensorType.String())
	}
	return strcase.ToSnake(id.String())
}

// Units returns the units for the value of this sensor.
func (s *Sensor) Units() string {
	return s.units
}

// String will format the sensor name and value as a pretty string.
func (s *Sensor) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %v %s [%s] (id: %s, path: %s, chip: %s)", s.Name(), s.Value(), s.Units(), s.SensorType, s.ID(), s.SysFSPath, s.Chip())
	for i, a := range s.Attributes {
		if i == 0 {
			fmt.Fprintf(&b, " (")
		}
		b.WriteString(a.String())
		if i < len(s.Attributes)-1 {
			fmt.Fprintf(&b, ", ")
		}
		if i == len(s.Attributes)-1 {
			fmt.Fprintf(&b, ")")
		}
	}
	return b.String()
}

func (s *Sensor) updateFromFile(file *sensorFile) error {
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

func (f *sensorFile) getSensorType() (sensorType SensorType, scaleFactor float64, units string) {
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

func getSensors(path string) ([]*Sensor, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// retrieve the chip name
	var chipLabel, chipID string
	l, err := getFileContents(filepath.Join(path, "name"))
	if err == nil {
		chipLabel = l
	}
	chipID = filepath.Base(path)

	var deviceModel string
	fh, err := os.Stat(filepath.Join(path, "device", "model"))
	if err == nil && fh.Mode().IsRegular() {
		m, err := getFileContents(filepath.Join(path, "device", "model"))
		if err == nil {
			deviceModel = m
		}
	}

	// gather all valid sensor files
	var allSensorFiles []*sensorFile
	for _, f := range files {
		// ignore directories
		if f.IsDir() {
			continue
		}
		// ignore files that can't be parsed as a sensor
		sensorType, sensorAttr, ok := strings.Cut(f.Name(), "_")
		if !ok {
			continue
		}
		allSensorFiles = append(allSensorFiles, &sensorFile{
			path:       path,
			filename:   f.Name(),
			sensorType: sensorType,
			sensorAttr: sensorAttr,
		})
	}

	// generate a map of allSensors from the files.
	allSensors := make(map[string]*Sensor)
	var mu sync.Mutex
	genSensorsPool := pool.New().WithErrors()
	for _, sensorFile := range allSensorFiles {
		genSensorsPool.Go(func() error {
			t, sf, u := sensorFile.getSensorType()
			trackerID := sensorFile.sensorType + "_" + t.String()
			mu.Lock()
			defer mu.Unlock()
			// if this sensor is already tracked, update it from the sensorFile contents
			if _, ok := allSensors[trackerID]; ok {
				return allSensors[trackerID].updateFromFile(sensorFile)
			}
			// otherwise, its a new sensor, start tracking it
			allSensors[trackerID] = &Sensor{
				chipLabel:   chipLabel,
				chipID:      chipID,
				deviceModel: deviceModel,
				id:          sensorFile.sensorType,
				SensorType:  t,
				SysFSPath:   path,
				scaleFactor: sf,
				units:       u,
			}
			return allSensors[trackerID].updateFromFile(sensorFile)
		})
	}
	err = genSensorsPool.Wait()
	var s []*Sensor
	for _, sensor := range allSensors {
		s = append(s, sensor)
	}
	return s, err
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

// getFileContents retrieves the contents of the given file as a string. If the
// contents cannot be read, it will return "unknown" and an error.
func getFileContents(p string) (string, error) {
	var b []byte
	var err error
	if b, err = os.ReadFile(p); err != nil {
		return "unknown", err
	}
	return strings.TrimSpace(string(b)), nil
}

func getValueAsString(p string) (string, error) {
	return getFileContents(p)
}

func getValueAsFloat(p string) (float64, error) {
	strValue, err := getFileContents(p)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strValue, 64)
}

func getValueAsBool(p string) (bool, error) {
	strValue, err := getFileContents(p)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(strValue)
}
