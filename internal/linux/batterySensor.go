// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass/deviceClass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensorType"
	"github.com/joshuar/go-hass-agent/internal/hass/stateClass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=batteryProp -output batterySensorProps.go -linecomment

const (
	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"

	battType     batteryProp = iota + 1 // Battery Type
	percentage                          // Battery Level
	temperature                         // Battery Temperature
	voltage                             // Battery Voltage
	energy                              // Battery Energy
	energyRate                          // Battery Power
	battState                           // Battery State
	nativePath                          // Battery Path
	batteryLevel                        // Battery Level
	model                               // Battery Model
)

type batteryProp int

type upowerBattery struct {
	props    map[batteryProp]dbus.Variant
	dBusPath dbus.ObjectPath
}

func (b *upowerBattery) updateProp(ctx context.Context, prop batteryProp) {
	var p string
	switch prop {
	case battType:
		p = "Type"
	case percentage:
		p = "Percentage"
	case temperature:
		p = "Temperature"
	case voltage:
		p = "Voltage"
	case energy:
		p = "Energy"
	case energyRate:
		p = "EnergyRate"
	case battState:
		p = "State"
	case nativePath:
		p = "NativePath"
	case batteryLevel:
		p = "BatteryLevel"
	case model:
		p = "Model"
	}
	propValue, err := NewBusRequest(ctx, systemBus).
		Path(b.dBusPath).
		Destination(upowerDBusDest).
		GetProp("org.freedesktop.UPower.Device." + p)
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not update property %s. Not found?", p)
	} else {
		b.props[prop] = propValue
	}
}

func (b *upowerBattery) getProp(prop batteryProp) interface{} {
	return b.props[prop].Value()
}

func (b *upowerBattery) marshalBatteryStateUpdate(ctx context.Context, prop batteryProp) *upowerBatteryState {
	// log.Debug().Caller().Msgf("Marshalling update for %v for battery %v", prop.String(), b.getProp(NativePath).(string))
	state := &upowerBatteryState{
		batteryID: b.getProp(nativePath).(string),
		model:     b.getProp(model).(string),
		prop: upowerBatteryProp{
			name:  prop,
			value: b.getProp(prop),
		},
	}
	switch prop {
	case energyRate:
		b.updateProp(ctx, voltage)
		b.updateProp(ctx, energy)
		state.attributes = &struct {
			Voltage float64 `json:"Voltage"`
			Energy  float64 `json:"Energy"`
		}{
			Voltage: b.getProp(voltage).(float64),
			Energy:  b.getProp(energy).(float64),
		}
	case percentage:
		fallthrough
	case batteryLevel:
		state.attributes = &struct {
			Type string `json:"Battery Type"`
		}{
			Type: stringType(b.getProp(battType).(uint32)),
		}
	}
	return state
}

type upowerBatteryProp struct {
	value interface{}
	name  batteryProp
}

type upowerBatteryState struct {
	attributes interface{}
	prop       upowerBatteryProp
	batteryID  string
	model      string
}

// uPowerBatteryState implements hass.SensorUpdate

func (state *upowerBatteryState) Name() string {
	return state.model + state.prop.name.String()
}

func (state *upowerBatteryState) ID() string {
	return state.batteryID + "_" + strings.ToLower(strcase.ToSnake(state.prop.name.String()))
}

func (state *upowerBatteryState) Icon() string {
	switch state.prop.name {
	case percentage:
		if state.prop.value.(float64) >= 95 {
			return "mdi:battery"
		} else {
			return fmt.Sprintf("mdi:battery-%d", int(math.Round(state.prop.value.(float64)/10)*10))
		}
	case energyRate:
		if math.Signbit(state.prop.value.(float64)) {
			return "mdi:battery-minus"
		} else {
			return "mdi:battery-plus"
		}
	default:
		return "mdi:battery"
	}
}

func (state *upowerBatteryState) SensorType() sensorType.SensorType {
	return sensorType.TypeSensor
}

func (state *upowerBatteryState) DeviceClass() deviceClass.SensorDeviceClass {
	switch state.prop.name {
	case percentage:
		return deviceClass.SensorBattery
	case temperature:
		return deviceClass.SensorTemperature
	case energyRate:
		return deviceClass.SensorPower
	default:
		return 0
	}
}

func (state *upowerBatteryState) StateClass() stateClass.SensorStateClass {
	switch state.prop.name {
	case percentage:
		fallthrough
	case temperature:
		fallthrough
	case energyRate:
		return stateClass.StateMeasurement
	default:
		return 0
	}
}

func (state *upowerBatteryState) State() interface{} {
	switch state.prop.name {
	case voltage:
		fallthrough
	case temperature:
		fallthrough
	case energy:
		fallthrough
	case energyRate:
		fallthrough
	case percentage:
		return state.prop.value.(float64)
	case battState:
		return stringState(state.prop.value.(uint32))
	case batteryLevel:
		return stringLevel(state.prop.value.(uint32))
	default:
		return state.prop.value.(string)
	}
}

func (state *upowerBatteryState) Units() string {
	switch state.prop.name {
	case percentage:
		return "%"
	case temperature:
		return "°C"
	case energyRate:
		return "W"
	default:
		return ""
	}
}

func (state *upowerBatteryState) Category() string {
	return "diagnostic"
}

func (state *upowerBatteryState) Attributes() interface{} {
	return state.attributes
}

func stringState(state uint32) string {
	switch state {
	case 1:
		return "Charging"
	case 2:
		return "Discharging"
	case 3:
		return "Empty"
	case 4:
		return "Fully Charged"
	case 5:
		return "Pending Charge"
	case 6:
		return "Pending Discharge"
	default:
		return "Unknown"
	}
}

func stringType(t uint32) string {
	switch t {
	case 0:
		return "Unknown"
	case 1:
		return "Line Power"
	case 2:
		return "Battery"
	case 3:
		return "Ups"
	case 4:
		return "Monitor"
	case 5:
		return "Mouse"
	case 6:
		return "Keyboard"
	case 7:
		return "Pda"
	case 8:
		return "Phone"
	default:
		return "Unknown"
	}
}

func stringLevel(l uint32) string {
	switch l {
	case 0:
		return "Unknown"
	case 1:
		return "None"
	case 3:
		return "Low"
	case 4:
		return "Critical"
	case 6:
		return "Normal"
	case 7:
		return "High"
	case 8:
		return "Full"
	default:
		return "Unknown"
	}
}

func BatteryUpdater(ctx context.Context, status chan interface{}) {
	batteryList := NewBusRequest(ctx, systemBus).
		Path(upowerDBusPath).
		Destination(upowerDBusDest).
		GetData(upowerGetDevicesMethod).AsObjectPathList()
	if batteryList == nil {
		log.Debug().Caller().
			Msg("Unable to get any battery devices from DBus.")
		return
	}

	batteryTracker := make(map[string]*upowerBattery)
	for _, v := range batteryList {

		// Track this battery in batteryTracker.
		batteryID := string(v)
		batteryTracker[batteryID] = &upowerBattery{
			dBusPath: v,
		}
		batteryTracker[batteryID].props = make(map[batteryProp]dbus.Variant)
		batteryTracker[batteryID].updateProp(ctx, nativePath)
		batteryTracker[batteryID].updateProp(ctx, battType)
		batteryTracker[batteryID].updateProp(ctx, model)

		// Standard battery properties as sensors
		for _, prop := range []batteryProp{battState} {
			batteryTracker[batteryID].updateProp(ctx, prop)
			stateUpdate := batteryTracker[batteryID].marshalBatteryStateUpdate(ctx, prop)
			if stateUpdate != nil {
				status <- stateUpdate
			}
		}

		// For some battery types, track additional properties as sensors
		if batteryTracker[batteryID].getProp(battType).(uint32) == 2 {
			for _, prop := range []batteryProp{percentage, temperature, energyRate} {
				batteryTracker[batteryID].updateProp(ctx, prop)
				stateUpdate := batteryTracker[batteryID].marshalBatteryStateUpdate(ctx, prop)
				if stateUpdate != nil {
					status <- stateUpdate
				}
			}
		} else {
			batteryTracker[batteryID].updateProp(ctx, batteryLevel)
			if batteryTracker[batteryID].getProp(batteryLevel).(uint32) != 1 {
				stateUpdate := batteryTracker[batteryID].marshalBatteryStateUpdate(ctx, batteryLevel)
				if stateUpdate != nil {
					status <- stateUpdate
				}
			}
		}

		// Create a DBus signal match to watch for property changes for this
		// battery. If a property changes, check it is one we want to track and
		// if so, update the battery's state in batteryTracker and send the
		// update back to Home Assistant.
		err := NewBusRequest(ctx, systemBus).
			Path(dbus.ObjectPath(v)).
			Match([]dbus.MatchOption{
				dbus.WithMatchObjectPath(dbus.ObjectPath(v)),
				dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
			}).
			Event("org.freedesktop.DBus.Properties.PropertiesChanged").
			Handler(func(s *dbus.Signal) {
				batteryID := string(s.Path)
				props := s.Body[1].(map[string]dbus.Variant)
				for propName, propValue := range props {
					for BatteryProp := range batteryTracker[batteryID].props {
						if propName == BatteryProp.String() {
							batteryTracker[batteryID].props[BatteryProp] = propValue
							log.Debug().Caller().
								Msgf("Updating battery property %v to %v", BatteryProp.String(), propValue.Value())
							stateUpdate := batteryTracker[batteryID].marshalBatteryStateUpdate(ctx, BatteryProp)
							if stateUpdate != nil {
								status <- stateUpdate
							}
						}
					}
				}
			}).
			AddWatch(ctx)
		if err != nil {
			log.Debug().Caller().Err(err).
				Msg("Failed to create DBus battery property watch.")
		}
	}
}
