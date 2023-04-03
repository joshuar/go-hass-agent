package device

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=batteryProp -output battery_linux_props.go -trimprefix battery

const (
	upowerDBusDest         = "org.freedesktop.UPower"
	upowerDBusPath         = "/org/freedesktop/UPower"
	upowerGetDevicesMethod = "org.freedesktop.UPower.EnumerateDevices"
	upowerGetPropsMethod   = "org.freedesktop.DBus.Properties.GetAll"

	Percentage batteryProp = iota
	Temperature
	EnergyRate
	Voltage
	Energy
	batteryState
	NativePath
	batteryType
)

type batteryProp int

type upowerBattery struct {
	props map[batteryProp]dbus.Variant
}

func (s *upowerBattery) LevelPercent() float64 {
	return s.props[Percentage].Value().(float64)
}

func (s *upowerBattery) Temperature() float64 {
	return s.props[Temperature].Value().(float64)
}

func (s *upowerBattery) Health() string {
	return "good"
}

func (s *upowerBattery) Power() float64 {
	return s.props[EnergyRate].Value().(float64)
}

func (s *upowerBattery) Voltage() float64 {
	return s.props[Voltage].Value().(float64)
}

func (s *upowerBattery) Energy() float64 {
	return s.props[Energy].Value().(float64)
}

func (s *upowerBattery) ChargerType() string {
	return "Unknown"
}

func (s *upowerBattery) State() string {
	switch s.props[batteryState].Value().(uint32) {
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

func (s *upowerBattery) ID() string {
	return s.props[NativePath].Value().(string)
}

func (s *upowerBattery) Type() interface{} {
	return s.props[batteryType].Value()
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
		batteryID := string(v)
		batteryTracker[batteryID] = &upowerBattery{}
		batteryTracker[batteryID].props = make(map[batteryProp]dbus.Variant)
		for _, k := range []batteryProp{Percentage, Temperature, EnergyRate, Voltage, Energy, batteryState, NativePath, batteryType} {
			p, err := deviceAPI.GetDBusProp(systemBus, upowerDBusDest, dbus.ObjectPath(batteryID), "org.freedesktop.UPower.Device."+k.String())
			if err != nil {
				log.Debug().Caller().Msgf(err.Error())
			}
			batteryTracker[batteryID].props[k] = p
		}
		status <- batteryTracker[batteryID]

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
					for batteryProp := range batteryTracker[batteryID].props {
						if propName == batteryProp.String() {
							batteryTracker[batteryID].props[batteryProp] = propValue
							log.Debug().Caller().
								Msgf("Updating battery property %v to %v", batteryProp.String(), propValue.Value())
							status <- batteryTracker[batteryID]
						}
					}
				}
			},
		}
		deviceAPI.WatchEvents <- batteryChangeSignal
	}

	spew.Dump(batteryTracker)

	for {
		select {
		case <-status:
			log.Debug().Caller().
				Msg("Stopping Linux battery updater.")
			return
		}
	}
}
