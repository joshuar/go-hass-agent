// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"github.com/iancoleman/strcase"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

const (
	DataSrcDbus   = "D-Bus"
	DataSrcProcfs = "ProcFS"
	DataSrcSysfs  = "SysFS"
)

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
	DeviceClassValue sensor.SensorDeviceClass
	StateClassValue  sensor.SensorStateClass
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

func (l *Sensor) SensorType() sensor.SensorType {
	if l.IsBinary {
		return sensor.TypeBinary
	}
	return sensor.TypeSensor
}

func (l *Sensor) Category() string {
	if l.IsDiagnostic {
		return "diagnostic"
	}
	return ""
}

func (l *Sensor) DeviceClass() sensor.SensorDeviceClass {
	return l.DeviceClassValue
}

func (l *Sensor) StateClass() sensor.SensorStateClass {
	return l.StateClassValue
}

func (l *Sensor) Icon() string {
	return l.IconString
}

func (l *Sensor) Units() string {
	return l.UnitsString
}

func (l *Sensor) Attributes() any {
	if l.SensorSrc != "" {
		return struct {
			DataSource string `json:"Data Source"`
		}{
			DataSource: l.SensorSrc,
		}
	}
	return nil
}
