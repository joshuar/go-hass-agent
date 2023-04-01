package device

import (
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

type upowerBattery struct {
	details map[string]dbus.Variant
}

func (s *upowerBattery) LevelPercent() float64 {
	return s.details["Percentage"].Value().(float64)
}

func (s *upowerBattery) Temperature() float64 {
	return s.details["Temperature"].Value().(float64)
}

func (s *upowerBattery) Health() string {
	return "good"
}

func (s *upowerBattery) Power() float64 {
	return s.details["EnergyRate"].Value().(float64)
}

func (s *upowerBattery) Voltage() float64 {
	return s.details["Voltage"].Value().(float64)
}

func (s *upowerBattery) Energy() float64 {
	return s.details["Energy"].Value().(float64)
}

func (s *upowerBattery) ChargerType() string {
	return "Unknown"
}

func (s *upowerBattery) State() string {
	switch s.details["State"].Value().(uint32) {
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
	return s.details["NativePath"].Value().(string)
}

func (s *upowerBattery) Type() interface{} {
	return s.details["Type"].Value()
}

func BatteryUpdater(status chan interface{}) {

	conn, err := ConnectSystemDBus()
	if err != nil {
		log.Debug().Caller().
			Msgf("Could not connect to DBus to monitor batteries: %v", err)
		return
	}
	defer conn.Close()

	obj := conn.Object("org.freedesktop.UPower", "/org/freedesktop/UPower")
	var batteryList []dbus.ObjectPath
	err = obj.Call("org.freedesktop.UPower.EnumerateDevices", 0).Store(&batteryList)
	if err != nil {
		log.Debug().Caller().
			Msgf("Unable to find all battery devices: %v", err)
		return
	}

	ticker := time.NewTicker(time.Second * 30)
	tickerDone := make(chan bool)

	for {
		select {
		case <-tickerDone:
			return
		case <-ticker.C:
			for i := 0; i < len(batteryList); i++ {
				battery := &upowerBattery{}
				var batteryInfo map[string]dbus.Variant
				obj := conn.Object("org.freedesktop.UPower", batteryList[i])
				err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.UPower.Device").Store(&batteryInfo)
				if err != nil {
					log.Debug().Caller().
						Msgf("Could not get properties for battery %s: %v", batteryList[i], err)
				} else {
					battery.details = batteryInfo
					status <- battery
				}
			}
		case <-status:
			log.Debug().Caller().
				Msg("Stopping Linux battery updater.")
			conn.Close()
			tickerDone <- true
		}
	}
}
