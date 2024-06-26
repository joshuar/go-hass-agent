// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"errors"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
)

const (
	DataSrcDbus   = "D-Bus"
	DataSrcProcfs = "ProcFS"
	DataSrcSysfs  = "SysFS"
)

var ErrUnimplemented = errors.New("unimplemented functionality")

// Sensor represents a generic sensor on the Linux platform. Most sensors
// will be able to use this struct, which satisfies the sensor.Sensor
// interface, alllowing them to be sent as a sensor to Home Assistant.
type Sensor struct {
	Value       any
	IconString  string
	UnitsString string
	SensorSrc   string
	SensorTypeValue
	IsBinary         bool
	IsDiagnostic     bool
	DeviceClassValue types.DeviceClass
	StateClassValue  types.StateClass
}

// linuxSensor satisfies the sensor.Sensor interface, allowing it to be sent as
// a sensor update to Home Assistant. Any of the methods below can be overridden
// by embedding linuxSensor in another struct and defining the needed function.

func (l *Sensor) Name() string {
	return l.SensorTypeValue.String()
}

func (l *Sensor) ID() string {
	return strcase.ToSnake(l.SensorTypeValue.String())
}

func (l *Sensor) State() any {
	return l.Value
}

func (l *Sensor) SensorType() types.SensorClass {
	if l.IsBinary {
		return types.BinarySensor
	}

	return types.Sensor
}

func (l *Sensor) Category() string {
	if l.IsDiagnostic {
		return "diagnostic"
	}

	return ""
}

func (l *Sensor) DeviceClass() types.DeviceClass {
	return l.DeviceClassValue
}

func (l *Sensor) StateClass() types.StateClass {
	return l.StateClassValue
}

func (l *Sensor) Icon() string {
	return l.IconString
}

func (l *Sensor) Units() string {
	return l.UnitsString
}

func (l *Sensor) Attributes() map[string]any {
	attributes := make(map[string]any)

	if l.SensorSrc != "" {
		attributes["data_source"] = l.SensorSrc
	}

	if l.UnitsString != "" {
		attributes["native_unit_of_measurement"] = l.UnitsString
	}

	return attributes
}

func (l *Sensor) String() string {
	var sensorStr strings.Builder

	fmt.Fprintf(&sensorStr, "Name: %s (ID: %s)", l.Name(), l.ID())

	if l.DeviceClass() > 0 {
		fmt.Fprintf(&sensorStr, " Device Class: %s", l.DeviceClass().String())
	}

	if l.StateClass() > 0 {
		fmt.Fprintf(&sensorStr, " State Class: %s", l.DeviceClass().String())
	}

	fmt.Fprintf(&sensorStr, " Value: %v", l.Value)

	if l.UnitsString != "" {
		fmt.Fprintf(&sensorStr, " %s", l.UnitsString)
	}

	if l.Attributes() != nil {
		fmt.Fprintf(&sensorStr, " Attributes: %v", l.Attributes())
	}

	return sensorStr.String()
}
