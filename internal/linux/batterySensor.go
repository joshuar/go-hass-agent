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
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
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
			Type:       batteryType(dbushelpers.VariantToValue[uint32](b.getValue(battType))).String(),
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
	linuxSensor
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
		pc, ok := state.prop.value.(float64)
		if !ok {
			return "mdi:battery-unknown"
		}
		if pc >= 95 {
			return "mdi:battery"
		} else {
			return fmt.Sprintf("mdi:battery-%d", int(math.Round(pc/10)*10))
		}
	case battEnergyRate:
		er, ok := state.prop.value.(float64)
		if !ok {
			return "mdi:battery"
		}
		if math.Signbit(er) {
			return "mdi:battery-minus"
		} else {
			return "mdi:battery-plus"
		}
	default:
		return "mdi:battery"
	}
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
	propType := state.prop.name
	rawValue := state.prop.value
	if rawValue == nil {
		return sensor.StateUnknown
	}
	switch propType {
	case battVoltage, battTemp, battEnergy, battEnergyRate, battPercentage:
		if value, ok := rawValue.(float64); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
	case battState:
		if value, ok := rawValue.(battChargeState); !ok {
			return sensor.StateUnknown
		} else {
			return value.String()
		}
	case battLevel:
		if value, ok := rawValue.(batteryLevel); !ok {
			return sensor.StateUnknown
		} else {
			return value.String()
		}
	default:
		if value, ok := rawValue.(string); !ok {
			return sensor.StateUnknown
		} else {
			return value
		}
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

func BatteryUpdater(ctx context.Context) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)
	batteryList := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(upowerDBusPath).
		Destination(upowerDBusDest).
		GetData(upowerGetDevicesMethod).AsObjectPathList()
	if len(batteryList) == 0 {
		log.Warn().
			Msg("Unable to get any battery devices from D-Bus. Battery sensor will not run.")
		close(sensorCh)
		return sensorCh
	}

	for _, path := range batteryList {
		// Track this battery in batteryTracker.
		battery := newBattery(ctx, path)

		// Track state
		if stateUpdate := battery.marshalBatteryStateUpdate(ctx, battState); stateUpdate != nil {
			sensorCh <- stateUpdate
		}

		// For some battery types, track additional properties as sensors
		if dbushelpers.VariantToValue[uint32](battery.getValue(battType)) == 2 {
			for _, prop := range []sensorType{battPercentage, battTemp, battEnergyRate} {
				battery.updateProp(ctx, prop)
				if stateUpdate := battery.marshalBatteryStateUpdate(ctx, prop); stateUpdate != nil {
					sensorCh <- stateUpdate
				}
			}
		} else {
			battery.updateProp(ctx, battLevel)
			if dbushelpers.VariantToValue[uint32](battery.getValue(battLevel)) != 1 {
				if stateUpdate := battery.marshalBatteryStateUpdate(ctx, battLevel); stateUpdate != nil {
					sensorCh <- stateUpdate
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
				props, ok := s.Body[1].(map[string]dbus.Variant)
				if !ok {
					log.Debug().Msg("Could not map received signal to battery properties.")
					return
				}
				for propName, propValue := range props {
					for BatteryProp := range battery.props {
						if propName == BatteryProp.String() {
							battery.props[BatteryProp] = propValue
							log.Debug().Caller().
								Msgf("Updating battery property %v to %v", BatteryProp.String(), propValue.Value())
							if stateUpdate := battery.marshalBatteryStateUpdate(ctx, BatteryProp); stateUpdate != nil {
								sensorCh <- stateUpdate
							}
						}
					}
				}
			}).
			AddWatch(ctx)
		if err != nil {
			log.Debug().Caller().Err(err).
				Msg("Failed to create DBus battery property watch.")
			close(sensorCh)
			return sensorCh
		}
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped battery sensors.")
	}()
	return sensorCh
}
