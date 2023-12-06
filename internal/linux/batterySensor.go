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
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
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
	if !b.dBusPath.IsValid() {
		log.Warn().Msg("No D-Bus path for battery.")
		return
	}
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
	propValue, err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(b.dBusPath).
		Destination(upowerDBusDest).
		GetProp("org.freedesktop.UPower.Device." + p)
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not update property %s. Not found?", p)
	} else if propValue.Value() != nil {
		b.setValue(t, propValue)
	}
}

func (b *upowerBattery) getValue(t sensorType) dbus.Variant {
	if b != nil {
		if v, ok := b.props[t]; ok {
			return v
		}
	}
	return dbus.MakeVariant("")
}

func (b *upowerBattery) setValue(t sensorType, v dbus.Variant) {
	if b != nil {
		if _, ok := b.props[t]; ok {
			b.props[t] = v
		}
	}
}

func (b *upowerBattery) marshalBatteryStateUpdate(ctx context.Context, t sensorType) *upowerBatteryState {
	if b == nil {
		return nil
	}
	id := dbushelpers.VariantToValue[string](b.getValue(battNativePath))
	if id == "" {
		log.Warn().Msg("Battery does not have a usable path. Will not monitor.")
		return nil
	}
	state := &upowerBatteryState{
		batteryID: id,
		model:     dbushelpers.VariantToValue[string](b.getValue(battModel)),
		prop: upowerBatteryProp{
			name:  t,
			value: b.getValue(t),
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
			Voltage:    dbushelpers.VariantToValue[float64](b.getValue(battVoltage)),
			Energy:     dbushelpers.VariantToValue[float64](b.getValue(battEnergy)),
			DataSource: srcDbus,
		}
	case battPercentage, battLevel:
		state.attributes = &struct {
			Type       string `json:"Battery Type"`
			DataSource string `json:"Data Source"`
		}{
			Type:       battTypeAsString(dbushelpers.VariantToValue[uint32](b.getValue(battType))),
			DataSource: srcDbus,
		}
	}
	return state
}

func newBattery(ctx context.Context, path dbus.ObjectPath) *upowerBattery {
	b := &upowerBattery{
		dBusPath: path,
		props:    make(map[sensorType]dbus.Variant),
	}
	b.updateProp(ctx, battNativePath)
	b.updateProp(ctx, battType)
	b.updateProp(ctx, battModel)
	b.updateProp(ctx, battState)
	return b
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
		return battStateAsString(state.prop.value.(uint32))
	case battLevel:
		return battLevelAsString(state.prop.value.(uint32))
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

func battStateAsString(state uint32) string {
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

func battTypeAsString(t uint32) string {
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

func battLevelAsString(l uint32) string {
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
	batteryList := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(upowerDBusPath).
		Destination(upowerDBusDest).
		GetData(upowerGetDevicesMethod).AsObjectPathList()
	if len(batteryList) == 0 {
		log.Warn().
			Msg("Unable to get any battery devices from D-Bus. Battery sensor will not run.")
		return
	}
	for _, path := range batteryList {
		// Track this battery in batteryTracker.
		battery := newBattery(ctx, path)

		// Track state
		if stateUpdate := battery.marshalBatteryStateUpdate(ctx, battState); stateUpdate != nil {
			if err := tracker.UpdateSensors(ctx, stateUpdate); err != nil {
				log.Error().Err(err).Msgf("Could not update battery %s.", battState.String())
			}
		}

		// For some battery types, track additional properties as sensors
		if dbushelpers.VariantToValue[uint32](battery.getValue(battType)) == 2 {
			for _, prop := range []sensorType{battPercentage, battTemp, battEnergyRate} {
				battery.updateProp(ctx, prop)
				if stateUpdate := battery.marshalBatteryStateUpdate(ctx, prop); stateUpdate != nil {
					if err := tracker.UpdateSensors(ctx, stateUpdate); err != nil {
						log.Error().Err(err).Msgf("Could not update battery %s.", prop.String())
					}
				}
			}
		} else {
			battery.updateProp(ctx, battLevel)
			if dbushelpers.VariantToValue[uint32](battery.getValue(battLevel)) != 1 {
				if stateUpdate := battery.marshalBatteryStateUpdate(ctx, battLevel); stateUpdate != nil {
					if err := tracker.UpdateSensors(ctx, stateUpdate); err != nil {
						log.Error().Err(err).Msgf("Could not update battery %s.", battLevel.String())
					}
				}
			}
		}

		// Create a DBus signal match to watch for property changes for this
		// battery. If a property changes, check it is one we want to track and
		// if so, update the battery's state in batteryTracker and send the
		// update back to Home Assistant.
		err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
			Path(path).
			Match([]dbus.MatchOption{
				dbus.WithMatchObjectPath(path),
				dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
			}).
			Event("org.freedesktop.DBus.Properties.PropertiesChanged").
			Handler(func(s *dbus.Signal) {
				var sensors []interface{}
				props := s.Body[1].(map[string]dbus.Variant)
				for propName, propValue := range props {
					for BatteryProp := range battery.props {
						if propName == BatteryProp.String() {
							battery.props[BatteryProp] = propValue
							log.Debug().Caller().
								Msgf("Updating battery property %v to %v", BatteryProp.String(), propValue.Value())
							if stateUpdate := battery.marshalBatteryStateUpdate(ctx, BatteryProp); stateUpdate != nil {
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
