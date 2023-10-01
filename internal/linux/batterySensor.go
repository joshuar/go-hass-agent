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
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

const (
	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"
)

type upowerBattery struct {
	props    map[sensorType]dbus.Variant
	dBusPath dbus.ObjectPath
}

func (b *upowerBattery) updateProp(ctx context.Context, t sensorType) {
	var p string
	switch t {
	case battType:
		p = "Type"
	case battPercentage:
		p = "Percentage"
	case battTemp:
		p = "Temperature"
	case battVoltage:
		p = "Voltage"
	case battEnergy:
		p = "Energy"
	case battEnergyRate:
		p = "EnergyRate"
	case battState:
		p = "State"
	case battNativePath:
		p = "NativePath"
	case battLevel:
		p = "BatteryLevel"
	case battModel:
		p = "Model"
	}
	propValue, err := NewBusRequest(ctx, SystemBus).
		Path(b.dBusPath).
		Destination(upowerDBusDest).
		GetProp("org.freedesktop.UPower.Device." + p)
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not update property %s. Not found?", p)
	} else {
		b.props[t] = propValue
	}
}

func (b *upowerBattery) getProp(t sensorType) interface{} {
	return b.props[t].Value()
}

func (b *upowerBattery) marshalBatteryStateUpdate(ctx context.Context, t sensorType) *upowerBatteryState {
	state := &upowerBatteryState{
		batteryID: b.getProp(battNativePath).(string),
		model:     b.getProp(battModel).(string),
		prop: upowerBatteryProp{
			name:  t,
			value: b.getProp(t),
		},
	}
	switch t {
	case battEnergyRate:
		b.updateProp(ctx, battVoltage)
		b.updateProp(ctx, battEnergy)
		state.attributes = &struct {
			DataSource string  `json:"Data Source"`
			Voltage    float64 `json:"Voltage"`
			Energy     float64 `json:"Energy"`
		}{
			Voltage:    b.getProp(battVoltage).(float64),
			Energy:     b.getProp(battEnergy).(float64),
			DataSource: srcDbus,
		}
	case battPercentage, battLevel:
		state.attributes = &struct {
			Type       string `json:"Battery Type"`
			DataSource string `json:"Data Source"`
		}{
			Type:       stringType(b.getProp(battType).(uint32)),
			DataSource: srcDbus,
		}
	}
	return state
}

type upowerBatteryProp struct {
	value interface{}
	name  sensorType
}

type upowerBatteryState struct {
	attributes interface{}
	prop       upowerBatteryProp
	batteryID  string
	model      string
}

// uPowerBatteryState implements hass.SensorUpdate

func (state *upowerBatteryState) Name() string {
	return state.model + " " + state.prop.name.String()
}

func (state *upowerBatteryState) ID() string {
	return state.batteryID + "_" + strings.ToLower(strcase.ToSnake(state.prop.name.String()))
}

func (state *upowerBatteryState) Icon() string {
	switch state.prop.name {
	case battPercentage:
		if state.prop.value.(float64) >= 95 {
			return "mdi:battery"
		} else {
			return fmt.Sprintf("mdi:battery-%d", int(math.Round(state.prop.value.(float64)/10)*10))
		}
	case battEnergyRate:
		if math.Signbit(state.prop.value.(float64)) {
			return "mdi:battery-minus"
		} else {
			return "mdi:battery-plus"
		}
	default:
		return "mdi:battery"
	}
}

func (state *upowerBatteryState) SensorType() sensor.SensorType {
	return sensor.TypeSensor
}

func (state *upowerBatteryState) DeviceClass() sensor.SensorDeviceClass {
	switch state.prop.name {
	case battPercentage:
		return sensor.SensorBattery
	case battTemp:
		return sensor.SensorTemperature
	case battEnergyRate:
		return sensor.SensorPower
	default:
		return 0
	}
}

func (state *upowerBatteryState) StateClass() sensor.SensorStateClass {
	switch state.prop.name {
	case battPercentage, battTemp, battEnergyRate:
		return sensor.StateMeasurement
	default:
		return 0
	}
}

func (state *upowerBatteryState) State() interface{} {
	switch state.prop.name {
	case battVoltage, battTemp, battEnergy, battEnergyRate, battPercentage:
		return state.prop.value.(float64)
	case battState:
		return stringState(state.prop.value.(uint32))
	case battLevel:
		return stringLevel(state.prop.value.(uint32))
	default:
		return state.prop.value.(string)
	}
}

func (state *upowerBatteryState) Units() string {
	switch state.prop.name {
	case battPercentage:
		return "%"
	case battTemp:
		return "°C"
	case battEnergyRate:
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
		return sensor.StateUnknown
	}
}

func stringType(t uint32) string {
	switch t {
	case 0:
		return sensor.StateUnknown
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
		return sensor.StateUnknown
	}
}

func stringLevel(l uint32) string {
	switch l {
	case 0:
		return sensor.StateUnknown
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
		return sensor.StateUnknown
	}
}

func BatteryUpdater(ctx context.Context, tracker device.SensorTracker) {
	batteryList := NewBusRequest(ctx, SystemBus).
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
		batteryTracker[batteryID].props = make(map[sensorType]dbus.Variant)
		batteryTracker[batteryID].updateProp(ctx, battNativePath)
		batteryTracker[batteryID].updateProp(ctx, battType)
		batteryTracker[batteryID].updateProp(ctx, battModel)

		// Standard battery properties as sensors
		for _, prop := range []sensorType{battState} {
			batteryTracker[batteryID].updateProp(ctx, prop)
			stateUpdate := batteryTracker[batteryID].marshalBatteryStateUpdate(ctx, prop)
			if stateUpdate != nil {
				if err := tracker.UpdateSensors(ctx, stateUpdate); err != nil {
					log.Error().Err(err).Msg("Could not update battery sensor.")
				}
			}
		}

		// For some battery types, track additional properties as sensors
		if batteryTracker[batteryID].getProp(battType).(uint32) == 2 {
			for _, prop := range []sensorType{battPercentage, battTemp, battEnergyRate} {
				batteryTracker[batteryID].updateProp(ctx, prop)
				stateUpdate := batteryTracker[batteryID].marshalBatteryStateUpdate(ctx, prop)
				if stateUpdate != nil {
					if err := tracker.UpdateSensors(ctx, stateUpdate); err != nil {
						log.Error().Err(err).Msg("Could not update battery sensor.")
					}
				}
			}
		} else {
			batteryTracker[batteryID].updateProp(ctx, battLevel)
			if batteryTracker[batteryID].getProp(battLevel).(uint32) != 1 {
				stateUpdate := batteryTracker[batteryID].marshalBatteryStateUpdate(ctx, battLevel)
				if stateUpdate != nil {
					if err := tracker.UpdateSensors(ctx, stateUpdate); err != nil {
						log.Error().Err(err).Msg("Could not update battery sensor.")
					}
				}
			}
		}

		// Create a DBus signal match to watch for property changes for this
		// battery. If a property changes, check it is one we want to track and
		// if so, update the battery's state in batteryTracker and send the
		// update back to Home Assistant.
		err := NewBusRequest(ctx, SystemBus).
			Path(v).
			Match([]dbus.MatchOption{
				dbus.WithMatchObjectPath(v),
				dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
			}).
			Event("org.freedesktop.DBus.Properties.PropertiesChanged").
			Handler(func(s *dbus.Signal) {
				var sensors []interface{}
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
								sensors = append(sensors, stateUpdate)
							}
						}
					}
				}
				if len(sensors) > 0 {
					if err := tracker.UpdateSensors(ctx, sensors...); err != nil {
						log.Error().Err(err).Msg("Could not update battery sensor.")
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
