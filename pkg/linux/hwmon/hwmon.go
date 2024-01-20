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
	chip        string
	label       string
	id          string
	units       string
	Attributes  []Attribute
	scaleFactor float64
	value       float64
	stype       SensorType
}

// Value returns the sensor value.
func (s *Sensor) Value() float64 {
	return s.value / s.scaleFactor
}

// Name returns a name for the sensor. It will be derived from the chip name
// plus either any label, else name of the sensor itself.
func (s *Sensor) Name() string {
	if s.label != "" {
		return s.chip + " " + s.label
	}
	return s.chip + " " + s.id
}

// Units returns the units for the value of this sensor.
func (s *Sensor) Units() string {
	return s.units
}

// String will format the sensor name and value as a pretty string.
func (s *Sensor) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %.2f %s [%s]", s.Name(), s.Value(), s.Units(), s.stype)
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
	switch d.item {
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
		s.Attributes = append(s.Attributes, Attribute{Name: d.item, Value: v / s.scaleFactor})
	}
	return nil
}

func newSensor(chip, name string) *Sensor {
	t, f, u := parseType(name)
	return &Sensor{
		chip:        chip,
		id:          name,
		stype:       t,
		scaleFactor: f,
		units:       u,
	}
}

// Attribute represents an attribute of a sensor, like its max, min or average
// value.
type Attribute struct {
	Name  string
	Value float64
}

// String will format the attribute name and value as a pretty string.
func (a *Attribute) String() string {
	return fmt.Sprintf("%s: %.2f", a.Name, a.Value)
}

type details struct {
	id    string
	item  string
	value string
}

func getDetails(path, file string) (*details, error) {
	i, t, ok := strings.Cut(file, "_")
	if !ok {
		return nil, fmt.Errorf("%s: not a sensor file", file)
	}
	v, err := getValue(filepath.Join(path, file))
	if err != nil {
		return nil, err
	}
	return &details{id: i, item: t, value: v}, nil
}

func getValue(p string) (string, error) {
	var b []byte
	var err error
	if b, err = os.ReadFile(p); err != nil {
		return "unknown", err
	}
	return strings.TrimSpace(string(b)), nil
}

func parseType(n string) (sensorType SensorType, scaleFactor float64, units string) {
	switch {
	case strings.Contains(n, "temp"):
		return Temp, 1000, "°C"
	case strings.Contains(n, "fan"):
		return Fan, 1, "rpm"
	case strings.Contains(n, "in"):
		return Voltage, 1000, "V"
	case strings.Contains(n, "pwm"):
		return PWM, 1, "Hz"
	case strings.Contains(n, "curr"):
		return Current, 1, "A"
	case strings.Contains(n, "power"):
		return Power, 1000, "W"
	case strings.Contains(n, "energy"):
		return Energy, 1000, "J"
	case strings.Contains(n, "humidity"):
		return Humidity, 1, "%"
	case strings.Contains(n, "freq"):
		return Frequency, 1000, "Hz"
	default:
		return Unknown, 1, ""
	}
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
			if _, ok := smap[d.id]; !ok {
				smap[d.id] = newSensor(namePrefix, d.id)
			}
			if err := smap[d.id].update(d); err != nil {
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
				d, _ := getDetails(path, f.Name())
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

// GetAllSensors returns a slice of Sensor objects, representing all detected
// chip sensors found on the host.
func GetAllSensors() []Sensor {
	files, err := os.ReadDir(hwmonPath)
	if err != nil {
		println(err)
		return nil
	}

	var sensors []Sensor

	for _, f := range files {
		c, _ := processChip(filepath.Join(hwmonPath, f.Name()))
		chip := <-c
		sensors = append(sensors, chip.Sensors...)
	}

	return sensors
}
