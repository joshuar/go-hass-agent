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

type SensorType int

// Chip represents a sensor chip exposed by the Linux kernel hardware monitoring
// API. These are retrieved from the directories in the sysfs /sys/devices tree
// under /sys/class/hwmon/hwmon*.
type Chip struct {
	Name    string
	Sensors []*Sensor
	chipID  int
}

func (c *Chip) update(newID int) {
	c.chipID = newID
	c.Name += " " + strconv.Itoa(c.chipID)
	for i := range c.Sensors {
		c.Sensors[i].chip = c.Name
	}
}

func processChip(path string) (*Chip, error) {
	n, err := getFileContents(filepath.Join(path, "name"))
	if err != nil {
		return nil, err
	}

	c := &Chip{
		Name: n,
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
	lastID := make(map[string]int)
	var mu sync.Mutex
	for _, f := range files {
		p.Go(func() error {
			chip, err := processChip(filepath.Join(hwmonPath, f.Name()))
			mu.Lock()
			defer mu.Unlock()
			if _, ok := lastID[chip.Name]; ok {
				lastID[chip.Name]++
			}
			chip.update(lastID[chip.Name])
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
	chip        string
	deviceModel string
	label       string
	id          string
	units       string
	Attributes  []Attribute
	scaleFactor float64
	value       float64
	SensorType  SensorType
}

// Value returns the sensor value.
func (s *Sensor) Value() float64 {
	return s.value / s.scaleFactor
}

// Name returns a name for the sensor. It will be derived from the chip name
// plus either any label, else name of the sensor itself.
func (s *Sensor) Name() string {
	c := cases.Title(language.AmericanEnglish)
	var chipFormatted string
	if s.deviceModel != "" {
		chipFormatted = s.deviceModel
	} else {
		chipFormatted = c.String(strings.ReplaceAll(s.chip, "_", " "))
	}
	idFormatted := c.String(s.id)
	labelFormatted := c.String(s.label)
	switch {
	case s.SensorType == Alarm || s.SensorType == Intrusion:
		return chipFormatted + " " + idFormatted + " " + labelFormatted
	case s.label != "":
		return chipFormatted + " " + labelFormatted
	default:
		return chipFormatted + " " + idFormatted
	}
}

func (s *Sensor) ID() string {
	if s.SensorType == Alarm || s.SensorType == Intrusion {
		return strcase.ToSnake(s.chip + "_" + s.id + "_" + s.SensorType.String())
	}
	return strcase.ToSnake(s.chip + "_" + s.id)
}

// Units returns the units for the value of this sensor.
func (s *Sensor) Units() string {
	return s.units
}

// String will format the sensor name and value as a pretty string.
func (s *Sensor) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %.3f %s [%s] (id: %s)", s.Name(), s.Value(), s.Units(), s.SensorType, s.ID())
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
	switch {
	case file.sensorAttr == "label":
		l, err := file.getValueAsString()
		if err != nil {
			return err
		}
		s.label = l
	case file.sensorAttr == "input":
		v, err := file.getValueAsFloat()
		if err != nil {
			return err
		}
		s.value = v
	case strings.Contains(file.sensorAttr, "alarm"):
		v, err := file.getValueAsFloat()
		if err != nil {
			return err
		}
		s.value = v
		s.label = "alarm"
	case strings.Contains(file.sensorAttr, "intrusion"):
		v, err := file.getValueAsFloat()
		if err != nil {
			return err
		}
		s.value = v
		s.label = "intrusion"
	default:
		v, err := file.getValueAsFloat()
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

func (f *sensorFile) getValueAsString() (string, error) {
	return getFileContents(filepath.Join(f.path, f.filename))
}

func (f *sensorFile) getValueAsFloat() (float64, error) {
	strValue, err := getFileContents(filepath.Join(f.path, f.filename))
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strValue, 64)
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
	chip, err := getFileContents(filepath.Join(path, "name"))
	if err != nil {
		return nil, err
	}

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
				chip:        chip,
				deviceModel: deviceModel,
				id:          sensorFile.sensorType,
				SensorType:  t,
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
