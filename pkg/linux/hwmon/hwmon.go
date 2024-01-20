// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hwmon

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
)

type SensorType int

// Chip represents a sensor chip exposed by the Linux kernel hardware monitoring
// API. These are retrieved from the directories in the sysfs /sys/devices tree
// under /sys/class/hwmon/hwmon*.
type Chip struct {
	Name    string
	Sensors []Sensor
}

// Sensor represents a single sensor exposed by a sensor chip. A Sensor may have
// a label, which is a formatted name of the sensor, otherwise it will just have
// a name. The Sensor will also have a value. It may also have zero or more
// Attributes, which are additional measurements like max/min/avg of the value.
type Sensor struct {
	chip, label, name string
	value             float64
	stype             SensorType
	Attributes        []Attribute
}

// Value returns the sensor value.
func (s *Sensor) Value() float64 {
	return s.value
}

// Name returns a name for the sensor. It will be derived from the chip name
// plus either any label, else name of the sensor itself.
func (s *Sensor) Name() string {
	if s.label != "" {
		return s.chip + " " + s.label
	}
	return s.chip + " " + s.name
}

// String will format the sensor name and value as a string.
func (s *Sensor) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %.f [%s]", s.Name(), s.value, s.stype)
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

func (s *Sensor) update(d details) error {
	switch d.suffix {
	case "input":
		v, err := strconv.ParseFloat(d.value, 64)
		if err != nil {
			return err
		}
		s.value = v
	case "label":
		s.label = d.value
	default:
		v, err := strconv.ParseFloat(d.value, 64)
		if err != nil {
			return err
		}
		s.Attributes = append(s.Attributes, Attribute{Name: d.suffix, Value: v})
	}
	return nil
}

func newSensor(chip, name string) *Sensor {
	return &Sensor{
		chip:  chip,
		name:  name,
		stype: getType(name),
	}
}

// Attribute represents an attribute of a sensor, like its max, min or average
// value.
type Attribute struct {
	Name  string
	Value float64
}

// String will format the attribute name and value as a string.
func (a *Attribute) String() string {
	return fmt.Sprintf("%s: %.f", a.Name, a.Value)
}

type details struct {
	prefix string
	suffix string
	value  string
}

func getSensorDetails(path, file string) (*details, error) {
	pfx, sfx, ok := strings.Cut(file, "_")
	if !ok {
		return nil, fmt.Errorf("%s: not a sensor file", file)
	}
	strValue, err := getValue(filepath.Join(path, file))
	if err != nil {
		return nil, err
	}
	return &details{prefix: pfx, suffix: sfx, value: strValue}, nil
}

func processSensors(path string) (sensorCh <-chan Sensor, errCh <-chan error) {
	c := make(chan Sensor)
	errc := make(chan error, 1)
	smap := make(map[string]*Sensor)
	var wg sync.WaitGroup

	files, err := os.ReadDir(path)
	if err != nil {
		errc <- err
		close(c)
		close(errc)
		return c, errc
	}

	namePrefix, err := getValue(filepath.Join(path, "name"))
	if err != nil {
		errc <- err
		close(c)
		close(errc)
		return c, errc
	}

	dc := make(chan details)
	var mu sync.Mutex
	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range dc {
			mu.Lock()
			if _, ok := smap[d.prefix]; !ok {
				smap[d.prefix] = newSensor(namePrefix, d.prefix)
			}
			if err := smap[d.prefix].update(d); err != nil {
				errc <- err
			}
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(dc)
		var wgFiles sync.WaitGroup
		for _, f := range files {
			wgFiles.Add(1)
			go func(f fs.DirEntry) {
				defer wgFiles.Done()
				d, _ := getSensorDetails(path, f.Name())
				if d != nil {
					dc <- *d
				}
			}(f)
			wgFiles.Wait()
		}
	}()

	go func() {
		wg.Wait()
		mu.Lock()
		for _, s := range smap {
			c <- *s
		}
		mu.Unlock()
		close(c)
		close(errc)
	}()
	return c, errc
}

func processChip(path string) (chipCh <-chan Chip, errCh <-chan error) {
	chipc := make(chan Chip, 1)
	errc := make(chan error, 1)

	n, err := getValue(filepath.Join(path, "name"))
	if err != nil {
		errc <- err
		close(chipc)
		return chipc, errc
	}

	c := Chip{
		Name: n,
	}

	var wg sync.WaitGroup
	s, e := processSensors(path)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for sensor := range s {
			c.Sensors = append(c.Sensors, sensor)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range e {
			errc <- err
		}
	}()
	go func() {
		wg.Wait()
		chipc <- c
		close(chipc)
		close(errc)
	}()
	return chipc, errc
}

func getValue(p string) (string, error) {
	var b []byte
	var err error
	if b, err = os.ReadFile(p); err != nil {
		return "unknown", err
	}
	return strings.TrimSpace(string(b)), nil
}

func getType(n string) SensorType {
	switch {
	case strings.Contains(n, "temp"):
		return Temp
	case strings.Contains(n, "fan"):
		return Fan
	case strings.Contains(n, "in"):
		return Voltage
	case strings.Contains(n, "pwm"):
		return PWM
	case strings.Contains(n, "curr"):
		return Current
	case strings.Contains(n, "power"):
		return Power
	case strings.Contains(n, "energy"):
		return Energy
	case strings.Contains(n, "humidity"):
		return Humidity
	case strings.Contains(n, "freq"):
		return Frequency
	default:
		return Unknown
	}
}

// GetAllSensors returns a slice of Sensor objects, representing all detected
// chip sensors found on the host.
func GetAllSensors() []Sensor {
	files, err := os.ReadDir(hwmonPath)
	if err != nil {
		println(err)
		return nil
	}

	var wg sync.WaitGroup
	var sensors []Sensor

	for _, f := range files {
		c, _ := processChip(filepath.Join(hwmonPath, f.Name()))
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chip := range c {
				sensors = append(sensors, chip.Sensors...)
			}
		}()
	}

	wg.Wait()
	return sensors
}
