package device

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=BatteryProp -output battery_props_linux.go -trimprefix batt

const (
	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"

	battType BatteryProp = iota + 1
	Percentage
	Temperature
	Voltage
	Energy
	EnergyRate
	battState
	NativePath
	BatteryLevel
)

type BatteryProp int

type upowerBattery struct {
	dBusPath dbus.ObjectPath
	props    map[BatteryProp]dbus.Variant
}

func (b *upowerBattery) updateProp(api *deviceAPI, prop BatteryProp) {
	propValue := api.GetDBusProp(systemBus, upowerDBusDest, b.dBusPath, "org.freedesktop.UPower.Device."+prop.String())
	if propValue.Value() == "" {
		log.Debug().Caller().
			Msgf("Could not update property %s. Not found?", prop.String())
	}
	b.props[prop] = propValue
}

func (b *upowerBattery) getProp(prop BatteryProp) interface{} {
	return b.props[prop].Value()
}

func (b *upowerBattery) marshallStateUpdate(api *deviceAPI, prop BatteryProp) *upowerBatteryState {
	// log.Debug().Caller().Msgf("Marshalling update for %v for battery %v", prop.String(), b.getProp(NativePath).(string))
	state := &upowerBatteryState{
		batteryID: b.getProp(NativePath).(string),
		prop: upowerBatteryProp{
			name:  prop,
			value: b.getProp(prop),
		},
	}
	switch prop {
	case EnergyRate:
		b.updateProp(api, Voltage)
		b.updateProp(api, Energy)
		state.attributes = &struct {
			Voltage float64 `json:"Voltage"`
			Energy  float64 `json:"Energy"`
		}{
			Voltage: b.getProp(Voltage).(float64),
			Energy:  b.getProp(Energy).(float64),
		}
	case Percentage:
		fallthrough
	case BatteryLevel:
		state.attributes = &struct {
			Type string `json:"Battery Type"`
		}{
			Type: stringType(b.getProp(battType).(uint32)),
		}
	}
	return state
}

type upowerBatteryProp struct {
	name  BatteryProp
	value interface{}
}

type upowerBatteryState struct {
	batteryID  string
	prop       upowerBatteryProp
	attributes interface{}
}

// uPowerBatteryState implements hass.SensorUpdate

func (state *upowerBatteryState) Name() string {
	switch state.prop.name {
	case Percentage:
		fallthrough
	case BatteryLevel:
		return state.batteryID + " Battery Level"
	case battState:
		return state.batteryID + " Battery State"
	case Temperature:
		return state.batteryID + " Battery Temperature"
	case EnergyRate:
		return state.batteryID + " Battery Power"
	default:
		return state.batteryID + strcase.ToDelimited(state.prop.name.String(), ' ')
	}
}

func (state *upowerBatteryState) ID() string {
	switch state.prop.name {
	case Percentage:
		fallthrough
	case BatteryLevel:
		return state.batteryID + "_battery_level"
	case battState:
		return state.batteryID + "_battery_state"
	case Temperature:
		return state.batteryID + "_battery_temperature"
	case EnergyRate:
		return state.batteryID + "_battery_power"
	default:
		return state.batteryID + "_" + strings.ToLower(strcase.ToSnake(state.prop.name.String()))
	}
}

func (state *upowerBatteryState) Icon() string {
	switch state.prop.name {
	case Percentage:
		if state.prop.value.(float64) >= 95 {
			return "mdi:battery"
		} else {
			return fmt.Sprintf("mdi:battery-%d", int(math.Round(state.prop.value.(float64)/10)*10))
		}
	case EnergyRate:
		if math.Signbit(state.prop.value.(float64)) {
			return "mdi:battery-minus"
		} else {
			return "mdi:battery-plus"
		}
	default:
		return "mdi:battery"
	}
}

func (state *upowerBatteryState) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (state *upowerBatteryState) DeviceClass() hass.SensorDeviceClass {
	switch state.prop.name {
	case Percentage:
		return hass.SensorBattery
	case Temperature:
		return hass.SensorTemperature
	case EnergyRate:
		return hass.SensorPower
	default:
		return 0
	}
}

func (state *upowerBatteryState) StateClass() hass.SensorStateClass {
	switch state.prop.name {
	case Percentage:
		fallthrough
	case Temperature:
		fallthrough
	case EnergyRate:
		return hass.Measurement
	default:
		return 0
	}
}

func (state *upowerBatteryState) State() interface{} {
	switch state.prop.name {
	case Voltage:
		fallthrough
	case Temperature:
		fallthrough
	case Energy:
		fallthrough
	case EnergyRate:
		fallthrough
	case Percentage:
		return state.prop.value.(float64)
	case battState:
		return stringState(state.prop.value.(uint32))
	case BatteryLevel:
		return stringLevel(state.prop.value.(uint32))
	default:
		return state.prop.value.(string)
	}
}

func (state *upowerBatteryState) Units() string {
	switch state.prop.name {
	case Percentage:
		return "%"
	case Temperature:
		return "°C"
	case EnergyRate:
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

func BatteryUpdater(ctx context.Context, status chan interface{}, done chan struct{}) {

	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor batteries.")
		return
	}

	batteryList := deviceAPI.GetDBusDataAsList(systemBus, upowerDBusDest, upowerDBusPath, upowerGetDevicesMethod, "")
	if batteryList == nil {
		log.Debug().Caller().
			Msg("Unable to any battery devices")
		return
	}

	batteryTracker := make(map[string]*upowerBattery)
	for _, v := range batteryList {

		// Track this battery in batteryTracker.
		batteryID := v
		batteryTracker[batteryID] = &upowerBattery{
			dBusPath: dbus.ObjectPath(v),
		}
		batteryTracker[batteryID].props = make(map[BatteryProp]dbus.Variant)
		batteryTracker[batteryID].updateProp(deviceAPI, NativePath)
		batteryTracker[batteryID].updateProp(deviceAPI, battType)

		// Standard battery properties as sensors
		for _, prop := range []BatteryProp{battState} {
			batteryTracker[batteryID].updateProp(deviceAPI, prop)
			stateUpdate := batteryTracker[batteryID].marshallStateUpdate(deviceAPI, prop)
			if stateUpdate != nil {
				status <- stateUpdate
			}
		}

		// For some battery types, track additional properties as sensors
		if batteryTracker[batteryID].getProp(battType).(uint32) == 2 {
			for _, prop := range []BatteryProp{Percentage, Temperature, EnergyRate} {
				batteryTracker[batteryID].updateProp(deviceAPI, prop)
				stateUpdate := batteryTracker[batteryID].marshallStateUpdate(deviceAPI, prop)
				if stateUpdate != nil {
					status <- stateUpdate
				}
			}
		} else {
			batteryTracker[batteryID].updateProp(deviceAPI, BatteryLevel)
			if batteryTracker[batteryID].getProp(BatteryLevel).(uint32) != 1 {
				stateUpdate := batteryTracker[batteryID].marshallStateUpdate(deviceAPI, BatteryLevel)
				if stateUpdate != nil {
					status <- stateUpdate
				}
			}
		}

		// Create a DBus signal match to watch for property changes for this
		// battery. If a property changes, check it is one we want to track and
		// if so, update the battery's state in batteryTracker and send the
		// update back to Home Assistant.
		batteryChangeSignal := &DBusWatchRequest{
			bus: systemBus,
			match: DBusSignalMatch{
				path: dbus.ObjectPath(v),
				intr: "org.freedesktop.DBus.Properties",
			},
			event: "org.freedesktop.DBus.Properties.PropertiesChanged",
			eventHandler: func(s *dbus.Signal) {
				batteryID := string(s.Path)
				props := s.Body[1].(map[string]dbus.Variant)
				for propName, propValue := range props {
					for BatteryProp := range batteryTracker[batteryID].props {
						if propName == BatteryProp.String() {
							batteryTracker[batteryID].props[BatteryProp] = propValue
							log.Debug().Caller().
								Msgf("Updating battery property %v to %v", BatteryProp.String(), propValue.Value())
							stateUpdate := batteryTracker[batteryID].marshallStateUpdate(deviceAPI, BatteryProp)
							if stateUpdate != nil {
								status <- stateUpdate
							}
						}
					}
				}
			},
		}
		deviceAPI.WatchEvents <- batteryChangeSignal
	}

	// <-done
	// log.Debug().Caller().
	// 	Msg("Stopping Linux battery updater.")
}
