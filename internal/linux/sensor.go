// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

const (
	SOURCE_DBUS  = "D-Bus"
	SOURCE_PROCFS = "ProcFS"
)

// linuxSensor represents a generic sensor on the Linux platform. Most sensors
// will be able to use this struct, which satisfies the tracker.Sensor
// interface, alllowing them to be sent as a sensor to Home Assistant.
type linuxSensor struct {
	value  interface{}
	icon   string
	units  string
	source string
	sensorType
	diagnostic  bool
	deviceClass sensor.SensorDeviceClass
	stateClass  sensor.SensorStateClass
}

// linuxSensor satisfies the tracker.Sensor interface, allowing it to be sent as
// a sensor update to Home Assistant. Any of the methods below can be overridden
// by embedding linuxSensor in another struct and defining the needed function.

func (l *linuxSensor) Name() string {
	return l.String()
}

func (l *linuxSensor) ID() string {
	return strcase.ToSnake(l.String())
}

func (l *linuxSensor) State() interface{} {
	return l.value
}

func (l *linuxSensor) SensorType() sensor.SensorType {
	return sensor.TypeSensor
}

func (l *linuxSensor) Category() string {
	if l.diagnostic {
		return "diagnostic"
	}
	return ""
}

func (l *linuxSensor) DeviceClass() sensor.SensorDeviceClass {
	return l.deviceClass
}

func (l *linuxSensor) StateClass() sensor.SensorStateClass {
	return l.stateClass
}

func (l *linuxSensor) Icon() string {
	return l.icon
}

func (l *linuxSensor) Units() string {
	return l.units
}

func (l *linuxSensor) Attributes() interface{} {
	if l.source != "" {
		return struct {
			DataSource string `json:"Data Source"`
		}{
			DataSource: l.source,
		}
	}
	return nil
}
