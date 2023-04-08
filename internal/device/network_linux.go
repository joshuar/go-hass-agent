package device

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=sensorType -output network_sensortypes_linux.go

const (
	dBusDest                    = "org.freedesktop.NetworkManager"
	dBusPath                    = "/org/freedesktop/NetworkManager"
	dBusActiveConnectionsMethod = "org.freedesktop.NetworkManager.ActiveConnections"

	ConnectionState sensorType = iota
	WifiSSID
	WifiFrequency
	WifiSpeed
	WifiStrength
)

type sensorType int

type networkConnection struct {
	dBusPath        dbus.ObjectPath
	name            string
	devicesPath     []dbus.ObjectPath
	connectionType  string
	connectionState uint32
	*wifiDetails
}

func newNetworkConnection(ctx context.Context, path dbus.ObjectPath) *networkConnection {

	var propPath = "org.freedesktop.NetworkManager.Connection.Active"

	deviceAPI, _ := FromContext(ctx)

	thisConnection := &networkConnection{
		dBusPath: path,
	}

	connObj := deviceAPI.GetDBusObject(systemBus,
		dBusDest,
		thisConnection.dBusPath)

	name, _ := connObj.GetProperty(propPath + ".Id")
	thisConnection.name = name.Value().(string)

	connectionType, _ := connObj.
		GetProperty(propPath + ".Type")
	thisConnection.connectionType = connectionType.Value().(string)

	connectionState, _ := connObj.
		GetProperty(propPath + ".State")
	thisConnection.connectionState = connectionState.Value().(uint32)

	devices, _ := connObj.
		GetProperty(propPath + ".Devices")
	thisConnection.devicesPath = devices.Value().([]dbus.ObjectPath)

	return thisConnection
}

func (conn *networkConnection) marshallStateUpdate(api *deviceAPI, sensor sensorType) *networkSensor {
	// log.Debug().Caller().Msgf("Marshalling update for %v for connection %v", sensor.String(), conn.name)
	var value, attributes interface{}
	var path dbus.ObjectPath
	switch sensor {
	case ConnectionState:
		value = conn.connectionState
		path = conn.dBusPath
		attributes = &struct {
			ConnectionType string `json:"Connection Type"`
		}{
			ConnectionType: conn.connectionType,
		}
	case WifiSSID:
		value = conn.wifiDetails.ssid
		path = conn.wifiDetails.dbusPath
	case WifiFrequency:
		value = conn.wifiDetails.frequency
		path = conn.wifiDetails.dbusPath
	case WifiSpeed:
		value = conn.wifiDetails.linkSpeed
		path = conn.wifiDetails.dbusPath
	case WifiStrength:
		value = conn.wifiDetails.signalStrength
		path = conn.wifiDetails.dbusPath
	}
	return &networkSensor{
		connection:       conn.name,
		dbusPath:         path,
		sensorType:       sensor,
		sensorValue:      value,
		sensorAttributes: attributes,
	}
}

type wifiDetails struct {
	dbusPath       dbus.ObjectPath
	ssid           string
	frequency      uint32
	linkSpeed      uint32
	signalStrength uint8
}

func getWifiDetails(ctx context.Context, path dbus.ObjectPath) *wifiDetails {
	deviceAPI, _ := FromContext(ctx)

	wirelessIntr := "org.freedesktop.NetworkManager.Device.Wireless"
	apIntr := "org.freedesktop.NetworkManager.AccessPoint"

	ap := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		path,
		wirelessIntr+".ActiveAccessPoint")

	apPath := ap.Value().(dbus.ObjectPath)

	ssid := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		apPath,
		apIntr+".Ssid")

	freq := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		apPath,
		apIntr+".Frequency")

	speed := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		apPath,
		apIntr+".MaxBitrate")

	strength := deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		apPath,
		apIntr+".Strength")

	return &wifiDetails{
		ssid:           string(ssid.Value().([]byte)),
		frequency:      freq.Value().(uint32),
		linkSpeed:      speed.Value().(uint32),
		signalStrength: strength.Value().(uint8),
	}
}

type networkSensor struct {
	connection       string
	dbusPath         dbus.ObjectPath
	sensorType       sensorType
	sensorValue      interface{}
	sensorAttributes interface{}
}

// networkSensor implements hass.SensorUpdate

func (state *networkSensor) Group() string {
	return state.connection
}

func (state *networkSensor) Name() string {
	return state.sensorType.String()
}

func (state *networkSensor) Icon() string {
	// switch state.prop.name {
	// case Percentage:
	// 	if state.prop.value.(float64) >= 95 {
	// 		return "mdi:battery"
	// 	} else {
	// 		return fmt.Sprintf("mdi:battery-%d", int(math.Round(state.prop.value.(float64)/10)*10))
	// 	}
	// case EnergyRate:
	// 	if math.Signbit(state.prop.value.(float64)) {
	// 		return "mdi:battery-minus"
	// 	} else {
	// 		return "mdi:battery-plus"
	// 	}
	// default:
	return "mdi:wifi"
	// }
}

func (state *networkSensor) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (state *networkSensor) DeviceClass() hass.SensorDeviceClass {
	switch state.sensorType {
	case WifiFrequency:
		return hass.Frequency
	case WifiSpeed:
		return hass.Data_rate
	default:
		return 0
	}
}

func (state *networkSensor) StateClass() hass.SensorStateClass {
	switch state.sensorType {
	case WifiFrequency:
		fallthrough
	case WifiSpeed:
		fallthrough
	case WifiStrength:
		return hass.Measurement
	default:
		return 0
	}
}

func (state *networkSensor) State() interface{} {
	switch state.sensorType {
	case ConnectionState:
		return stateToString(state.sensorValue.(uint32))
	}
	return state.sensorValue
}

func (state *networkSensor) Units() string {
	switch state.sensorType {
	case WifiFrequency:
		return "MHz"
	case WifiSpeed:
		return "Kb/s"
	case WifiStrength:
		return "%"
	default:
		return ""
	}
}

func (state *networkSensor) Category() string {
	return "diagnostic"
}

func (state *networkSensor) Attributes() interface{} {
	return state.sensorAttributes
}

func stateToString(state uint32) string {
	switch state {
	case 0:
		return "Unknown"
	case 1:
		return "Activating"
	case 2:
		return "Online"
	case 3:
		return "Deactivating"
	case 4:
		return "Offline"
	default:
		return "Unknown"
	}
}

func (s *networkSensor) marshalStateWatch() *DBusWatchRequest {
	var intr, event string
	var handler func(*dbus.Signal)
	switch s.sensorType {
	case ConnectionState:
		intr = "org.freedesktop.NetworkManager.Connection.Active"
		event = intr + ".StateChanged"
		handler = func(s *dbus.Signal) {
			log.Debug().Msgf("Got signal %s with %v", s.Name, s.Body)
		}
	}
	return &DBusWatchRequest{
		bus: systemBus,
		match: DBusSignalMatch{
			path: s.dbusPath,
			intr: intr,
		},
		event:        event,
		eventHandler: handler,
	}
}

func NetworkUpdater(ctx context.Context, status chan interface{}, done chan struct{}) {

	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor batteries.")
		return
	}

	connList := deviceAPI.GetDBusProp(systemBus, dBusDest, dBusPath, dBusActiveConnectionsMethod)

	connTracker := make(map[string]*networkConnection)

	for _, path := range connList.Value().([]dbus.ObjectPath) {
		conn := newNetworkConnection(ctx, path)
		id := conn.name
		connTracker[id] = conn

		switch connTracker[id].connectionType {
		case "802-11-wireless":
			log.Debug().Msgf("Fetching wifi details for %s at %s", id, connTracker[id].devicesPath[0])
			connTracker[id].wifiDetails = getWifiDetails(ctx, connTracker[id].devicesPath[0])
			wifiSSIDState := connTracker[id].
				marshallStateUpdate(deviceAPI, WifiSSID)
			status <- wifiSSIDState
			wifiFrequency := connTracker[id].
				marshallStateUpdate(deviceAPI, WifiFrequency)
			status <- wifiFrequency
			wifiSpeed := connTracker[id].
				marshallStateUpdate(deviceAPI, WifiSpeed)
			status <- wifiSpeed
			wifiStrength := connTracker[id].
				marshallStateUpdate(deviceAPI, WifiStrength)
			status <- wifiStrength
			// spew.Dump(connTracker[id].wifiDetails)
		}
		// connectionInfo, err := deviceAPI.GetDBusData(systemBus,
		// 	dBusDest, v, "org.freedesktop.DBus.Properties.GetAll", "org.freedesktop.NetworkManager.Connection.Active")
		// connectionTracker[connectionID].props = connectionInfo.(map[string]interface{})

		// ipv4Config, err := deviceAPI.GetDBusData(systemBus,
		// 	dBusDest, connectionTracker[connectionID].props["Ip4Config"].(dbus.ObjectPath), "org.freedesktop.DBus.Properties.GetAll", "org.freedesktop.NetworkManager.IP4Config")
		// connectionTracker[connectionID].addressDetails = &addressDetails{
		// 	ipv4: ipv4Config.(map[string]interface{}),
		// }
		connState := connTracker[id].
			marshallStateUpdate(deviceAPI, ConnectionState)
		status <- connState
		connStateWatch := connState.marshalStateWatch()
		deviceAPI.WatchEvents <- connStateWatch
		// spew.Dump(connTracker[path])
	}

	<-done
	log.Debug().Caller().
		Msg("Stopping Linux battery updater.")
}
