package device

import (
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=currentState

const (
	Unknown          currentState = 0
	Charging         currentState = 1
	Discharging      currentState = 2
	Empty            currentState = 3
	FullyCharged     currentState = 4
	PendingCharge    currentState = 5
	PendingDischarge currentState = 6
)

type currentState int

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
	return s.details["State"].Value().(currentState).String()
}

func (s *upowerBattery) ID() string {
	return s.details["NativePath"].Value().(string)
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
	done := make(chan bool)

	for {
		select {
		case <-done:
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
		}
	}
}
