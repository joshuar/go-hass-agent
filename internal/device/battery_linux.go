package device

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=BatteryProp -output battery_props_linux.go -trimprefix battery

const (
	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"

	// Note: order is important!
	batteryType BatteryProp = iota
	Percentage
	Temperature
	Voltage
	Energy
	EnergyRate
	batteryState
)

type BatteryProp int

type upowerBattery struct {
	name  string
	props map[BatteryProp]dbus.Variant
}

func (b *upowerBattery) MarshallStateUpdate(prop BatteryProp) *upowerBatteryState {
	// We don't send Energy or Voltage updates separately.
	// These are sent as part of the EnergyRate (Power) update.
	if prop == Energy || prop == Voltage || prop == batteryType {
		return nil
	}
	log.Debug().Caller().Msgf("Marshalling update for %v for battery %v", prop.String(), b.name)
	state := &upowerBatteryState{
		batteryID: b.name,
		prop: upowerBatteryProp{
			kind:  prop,
			value: b.props[prop].Value(),
		},
	}
	switch prop {
	case EnergyRate:
		state.attributes = &struct {
			Voltage float64 `json:"Voltage"`
			Energy  float64 `json:"Energy"`
		}{
			Voltage: b.props[Voltage].Value().(float64),
			Energy:  b.props[Energy].Value().(float64),
		}
	case Percentage:
		state.attributes = &struct {
			Type string `json:"Battery Type"`
		}{
			Type: stringType(b.props[batteryType].Value().(uint32)),
		}
	}
	return state
}

type upowerBatteryProp struct {
	kind  BatteryProp
	value interface{}
}

type upowerBatteryState struct {
	batteryID  string
	prop       upowerBatteryProp
	attributes interface{}
}

func (state *upowerBatteryState) ID() string {
	return state.batteryID
}

func (state *upowerBatteryState) Type() BatteryProp {
	return state.prop.kind
}

func (state *upowerBatteryState) Value() interface{} {
	switch state.prop.kind {
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
	case batteryState:
		return stringState(state.prop.value.(uint32))
	default:
		return state.prop.value
	}
}

func (state *upowerBatteryState) ExtraValues() interface{} {
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

func BatteryUpdater(ctx context.Context, status chan interface{}) {

	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor batteries.")
		return
	}

	batteryList, err := deviceAPI.GetDBusData(systemBus, upowerDBusDest, upowerDBusPath, upowerGetDevicesMethod)
	if err != nil {
		log.Debug().Caller().
			Msgf("Unable to find all battery devices: %v", err)
		return
	}

	batteryTracker := make(map[string]*upowerBattery)
	for _, v := range batteryList.([]dbus.ObjectPath) {

		// Populate the batteryTracker map with the battery's current
		// properties. Send it to Home Assistant.
		batteryID := string(v)
		batteryTracker[batteryID] = &upowerBattery{}
		batteryTracker[batteryID].props = make(map[BatteryProp]dbus.Variant)
		p, _ := deviceAPI.GetDBusProp(systemBus, upowerDBusDest, dbus.ObjectPath(batteryID), "org.freedesktop.UPower.Device.NativePath")
		batteryTracker[batteryID].name = p.Value().(string)

		for _, prop := range []BatteryProp{batteryType, Percentage, Temperature, Voltage, Energy, EnergyRate, batteryState} {
			propValue, err := deviceAPI.GetDBusProp(systemBus, upowerDBusDest, dbus.ObjectPath(batteryID), "org.freedesktop.UPower.Device."+prop.String())
			if err != nil {
				log.Debug().Caller().Msgf(err.Error())
			}
			batteryTracker[batteryID].props[prop] = propValue
			stateUpdate := batteryTracker[batteryID].MarshallStateUpdate(prop)
			if stateUpdate != nil {
				status <- stateUpdate
			}
		}
		// status <- batteryTracker[batteryID]

		// Create a DBus signal match to watch for property changes for this
		// battery. If a property changes, check it is one we want to track and
		// if so, update the battery's state in batteryTracker and send the
		// update back to Home Assistant.
		batteryChangeSignal := &DBusWatchRequest{
			bus: systemBus,
			match: DBusSignalMatch{
				path: v,
				intr: "org.freedesktop.DBus.Properties",
			},
			event: "org.freedesktop.DBus.Properties.PropertiesChanged",
			eventHandler: func(s *dbus.Signal) {
				log.Debug().Caller().Msg("Recieved changed battery state.")
				batteryID := string(s.Path)
				props := s.Body[1].(map[string]dbus.Variant)
				for propName, propValue := range props {
					for BatteryProp := range batteryTracker[batteryID].props {
						if propName == BatteryProp.String() {
							batteryTracker[batteryID].props[BatteryProp] = propValue
							log.Debug().Caller().
								Msgf("Updating battery property %v to %v", BatteryProp.String(), propValue.Value())

							stateUpdate := batteryTracker[batteryID].MarshallStateUpdate(BatteryProp)
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

	for {
		select {
		case <-status:
			log.Debug().Caller().
				Msg("Stopping Linux battery updater.")
			return
		}
	}
}
