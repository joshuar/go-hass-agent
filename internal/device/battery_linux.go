package device

import (
	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/logging"
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

// LevelPercent() float64
// Temperature() float64
// Health() string
// Power() float64
// ChargerType() string
// State() string
// ID() string
// Attributes() interface{}

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

	conn := ConnectSystemDBus()

	obj := conn.Object("org.freedesktop.UPower", "/org/freedesktop/UPower")
	var batteryList []dbus.ObjectPath
	err := obj.Call("org.freedesktop.UPower.EnumerateDevices", 0).Store(&batteryList)
	logging.CheckError(err)
	// var monitoringList []string
	for i := 0; i < len(batteryList); i++ {
		battery := &upowerBattery{}
		var batteryInfo map[string]dbus.Variant
		obj := conn.Object("org.freedesktop.UPower", batteryList[i])
		err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.UPower.Device").Store(&batteryInfo)
		logging.CheckError(err)
		// spew.Dump(batteryInfo)
		battery.details = batteryInfo
		status <- battery

		// monitoringList = append(monitoringList, fmt.Sprintf("type='signal',member='PropertiesChanged',path='%s',interface='org.freedesktop.UPower.Device'", batteryList[i]))
	}
	// var flag uint = 0

	// call := conn.BusObject().Call("org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, monitoringList, flag)
	// logging.CheckError(call.Err)

	// c := make(chan *dbus.Message, 10)
	// conn.Eavesdrop(c)
	// log.Debug().Caller().Msg("Monitoring D-Bus for app changes.")

	// var batteryProps interface{}
	// obj := conn.Object("org.freedesktop.UPower", batteryList[i])
	// err = obj.Call("org.freedesktop.DBus.Properties.GetAll", 0).Store(&batteryProps)
	// logging.CheckError(err)
	// spew.Dump(batteryProps)
	// for range c {
	// 	spew.Dump(c)
	// }

}
