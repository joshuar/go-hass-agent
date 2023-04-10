// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package device

import (
	"context"
	"net"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=networkProp -output network_props_linux.go

const (
	dBusDest                    = "org.freedesktop.NetworkManager"
	dBusPath                    = "/org/freedesktop/NetworkManager"
	dBusActiveConnectionsMethod = "org.freedesktop.NetworkManager.ActiveConnections"

	ConnectionState networkProp = iota
	ConnectionID
	ConnectionDevices
	ConnectionType
	ConnectionIPv4
	ConnectionIPv6
	AddressIPv4
	AddressIPv6
	WifiSSID
	WifiFrequency
	WifiSpeed
	WifiStrength
	WifiHWAddress
)

type networkProp int

func getNetProp(ctx context.Context, path dbus.ObjectPath, prop networkProp) dbus.Variant {
	deviceAPI, _ := FromContext(ctx)

	connIntr := "org.freedesktop.NetworkManager.Connection.Active"
	ipv4Intr := "org.freedesktop.NetworkManager.IP4Config"
	ipv6Intr := "org.freedesktop.NetworkManager.IP6Config"
	wirelessIntr := "org.freedesktop.NetworkManager.Device.Wireless"
	apIntr := "org.freedesktop.NetworkManager.AccessPoint"

	var dbusProp string
	switch prop {
	case ConnectionID:
		dbusProp = connIntr + ".Id"
	case ConnectionState:
		dbusProp = connIntr + ".State"
	case ConnectionType:
		dbusProp = connIntr + ".Type"
	case ConnectionDevices:
		dbusProp = connIntr + ".Devices"
	case ConnectionIPv4:
		dbusProp = connIntr + ".Ip4Config"
	case ConnectionIPv6:
		dbusProp = connIntr + ".Ip6Config"
	case AddressIPv4:
		dbusProp = ipv4Intr + ".AddressData"
	case AddressIPv6:
		dbusProp = ipv6Intr + ".AddressData"
	case WifiSSID:
		ap := deviceAPI.GetDBusProp(systemBus,
			dBusDest,
			path,
			wirelessIntr+".ActiveAccessPoint")
		apPath := ap.Value().(dbus.ObjectPath)
		dbusProp = apIntr + ".Ssid"
		path = apPath
	case WifiFrequency:
		ap := deviceAPI.GetDBusProp(systemBus,
			dBusDest,
			path,
			wirelessIntr+".ActiveAccessPoint")

		apPath := ap.Value().(dbus.ObjectPath)
		dbusProp = apIntr + ".Frequency"
		path = apPath
	case WifiSpeed:
		ap := deviceAPI.GetDBusProp(systemBus,
			dBusDest,
			path,
			wirelessIntr+".ActiveAccessPoint")

		apPath := ap.Value().(dbus.ObjectPath)
		dbusProp = apIntr + ".MaxBitrate"
		path = apPath
	case WifiStrength:
		ap := deviceAPI.GetDBusProp(systemBus,
			dBusDest,
			path,
			wirelessIntr+".ActiveAccessPoint")

		apPath := ap.Value().(dbus.ObjectPath)
		dbusProp = apIntr + ".Strength"
		path = apPath
	case WifiHWAddress:
		ap := deviceAPI.GetDBusProp(systemBus,
			dBusDest,
			path,
			wirelessIntr+".ActiveAccessPoint")

		apPath := ap.Value().(dbus.ObjectPath)
		dbusProp = apIntr + ".HwAddress"
		path = apPath
	}

	return deviceAPI.GetDBusProp(systemBus,
		dBusDest,
		path,
		dbusProp)
}

func getIPAddrProp(ctx context.Context, connProp networkProp, path dbus.ObjectPath) string {
	var addrProp networkProp
	switch connProp {
	case ConnectionIPv4:
		addrProp = AddressIPv4
	case ConnectionIPv6:
		addrProp = AddressIPv6
	}
	p := getNetProp(ctx, path, connProp)
	switch configPath := p.Value().(type) {
	case dbus.ObjectPath:
		propValue := getNetProp(ctx, configPath, addrProp)
		switch propValue.Value().(type) {
		case []map[string]dbus.Variant:
			addrs := propValue.Value().([]map[string]dbus.Variant)
			for _, a := range addrs {
				ip := net.ParseIP(a["address"].Value().(string))
				if ip.IsGlobalUnicast() {
					return ip.String()
				}
			}
		}
	default:
		return ""
	}
	return ""
}

type networkSensor struct {
	connection       string
	sensorType       networkProp
	sensorValue      interface{}
	sensorAttributes interface{}
}

// networkSensor implements hass.SensorUpdate

func (state *networkSensor) Name() string {
	switch state.sensorType {
	case ConnectionState:
		return state.connection + " State"
	case WifiSSID:
		return state.connection + " Wifi Connection"
	case WifiFrequency:
		return state.connection + " Wifi Frequency"
	case WifiSpeed:
		return state.connection + " Wifi Speed"
	case WifiStrength:
		return state.connection + " Wifi Signal Strength"
	default:
		return state.connection + " " + strcase.ToDelimited(state.sensorType.String(), ' ')
	}
}

func (state *networkSensor) ID() string {
	switch state.sensorType {
	case ConnectionState:
		return state.connection + "_connection_state"
	case WifiSSID:
		return state.connection + "_wifi_connection"
	case WifiFrequency:
		return state.connection + "_wifi_frequency"
	case WifiSpeed:
		return state.connection + "_wifi_speed"
	case WifiStrength:
		return state.connection + "_wifi_signal_strength"
	default:
		return state.connection + "_" + strcase.ToSnake(state.sensorType.String())
	}
}

func (state *networkSensor) Icon() string {
	switch state.sensorType {
	case ConnectionState:
		switch state.sensorValue {
		case "Online":
			return "mdi:network"
		case "Offline":
			return "mdi:close-network"
		default:
			return "mdi:help-network"
		}
	case WifiSSID:
		fallthrough
	case WifiFrequency:
		fallthrough
	case WifiSpeed:
		fallthrough
	case WifiStrength:
		return "mdi:wifi"
	}
	return "mdi:network"
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
	return state.sensorValue
}

func (state *networkSensor) Units() string {
	switch state.sensorType {
	case WifiFrequency:
		return "MHz"
	case WifiSpeed:
		return "kB/s"
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

func stateToString(state interface{}) string {
	switch state := state.(type) {
	case string:
		return "Unknown"
	case uint32:
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
	default:
		return "Unknown"
	}
}

func marshallNetworkStateUpdate(ctx context.Context, sensor networkProp, path dbus.ObjectPath, conn string, v dbus.Variant) *networkSensor {
	// log.Debug().Caller().Msgf("Marshalling update for %v for connection %v", sensor.String(), conn.name)
	var value, attributes interface{}
	switch sensor {
	case ConnectionState:
		value = stateToString(v.Value())
		attributes = &struct {
			ConnectionType string `json:"Connection Type"`
			Ipv4           string `json:"IPv4 Address"`
			Ipv6           string `json:"IPv6 Address"`
		}{
			ConnectionType: getNetProp(ctx, path, ConnectionType).Value().(string),
			Ipv4:           getIPAddrProp(ctx, ConnectionIPv4, path),
			Ipv6:           getIPAddrProp(ctx, ConnectionIPv6, path),
		}
	case WifiSSID:
		value = string(v.Value().([]byte))
		attributes = &struct {
			Bssid string `json:"BSSID"`
		}{
			Bssid: getNetProp(ctx, path, WifiHWAddress).Value().(string),
		}
	case WifiFrequency:
		value = v.Value().(uint32)
	case WifiSpeed:
		value = v.Value().(uint32)
	case WifiStrength:
		value = v.Value().(uint8)
	}
	return &networkSensor{
		connection:       conn,
		sensorType:       sensor,
		sensorValue:      value,
		sensorAttributes: attributes,
	}
}

func NetworkUpdater(ctx context.Context, status chan interface{}) {

	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor batteries.")
		return
	}

	connList := deviceAPI.GetDBusProp(systemBus, dBusDest, dBusPath, dBusActiveConnectionsMethod).Value().([]dbus.ObjectPath)

	if len(connList) > 0 {
		processConnections(ctx, connList, status)
	}

	log.Debug().Msg("Watching for connection changes.")
	connStateWatch := &DBusWatchRequest{
		bus:  systemBus,
		path: "/org/freedesktop/NetworkManager/ActiveConnection",
		match: []dbus.MatchOption{
			dbus.WithMatchPathNamespace("/org/freedesktop/NetworkManager/ActiveConnection"),
		},
		event: "org.freedesktop.DBus.Properties.PropertiesChanged",
		eventHandler: func(s *dbus.Signal) {
			switch {
			case s.Name == "org.freedesktop.NetworkManager.Connection.Active.StateChanged":
				name := getNetProp(ctx, s.Path, ConnectionID).Value().(string)
				if name != "" {
					connState := marshallNetworkStateUpdate(ctx, ConnectionState, s.Path, name, dbus.MakeVariant(s.Body[0]))
					status <- connState
				}
			}
		},
	}
	deviceAPI.WatchEvents <- connStateWatch
}

func processConnections(ctx context.Context, connList []dbus.ObjectPath, status chan interface{}) {

	// var activeConnectionIntr = "org.freedesktop.NetworkManager.Connection.Active"

	deviceAPI, deviceAPIExists := FromContext(ctx)
	if !deviceAPIExists {
		log.Debug().Caller().
			Msg("Could not connect to DBus to monitor batteries.")
		return
	}

	for _, path := range connList {
		name := getNetProp(ctx, path, ConnectionID).Value().(string)
		log.Debug().Msgf("Processing connection %s", name)
		// Fetch and monitor connection state
		state := getNetProp(ctx, path, ConnectionState)
		var stateValue uint32
		if state.Store(&stateValue) != nil {
			log.Debug().Msgf("Connection %s is not active.", name)
			return
		}
		connState := marshallNetworkStateUpdate(ctx, ConnectionState, path, name, state)
		status <- connState

		// Get connection type and then fetch and monitor additional type
		// dependent properties
		connType := getNetProp(ctx, path, ConnectionType)
		switch connType.Value().(string) {
		case "802-11-wireless":
			dp := getNetProp(ctx, path, ConnectionDevices)
			devicePath := dp.Value().([]dbus.ObjectPath)[0]

			wifiProps := []networkProp{WifiSSID, WifiFrequency, WifiSpeed, WifiStrength}

			for _, prop := range wifiProps {
				propValue := getNetProp(ctx, devicePath, prop)
				propState := marshallNetworkStateUpdate(ctx,
					prop,
					devicePath,
					name,
					propValue)
				status <- propState
			}

			wifiStateWatch := &DBusWatchRequest{
				bus:  systemBus,
				path: "/org/freedesktop/NetworkManager/AccessPoint",
				match: []dbus.MatchOption{
					dbus.WithMatchPathNamespace("/org/freedesktop/NetworkManager/AccessPoint"),
					dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
				},
				event: "org.freedesktop.DBus.Properties.PropertiesChanged",
				eventHandler: func(s *dbus.Signal) {
					updatedProps := s.Body[1].(map[string]dbus.Variant)
					for propName, propValue := range updatedProps {
						var propType networkProp
						switch propName {
						case "Ssid":
							propType = WifiSSID
						case "Frequency":
							propType = WifiFrequency
						case "Bitrate":
							propType = WifiSpeed
						case "Strength":
							propType = WifiStrength
						default:
							log.Debug().Msgf("Unhandled property %v changed to %v", propName, propValue)
						}
						if propType != 0 {
							propState := marshallNetworkStateUpdate(ctx,
								propType,
								devicePath,
								name,
								propValue)
							status <- propState
						}
					}
				},
			}
			deviceAPI.WatchEvents <- wifiStateWatch
		default:
			log.Debug().Msgf("No extra properties for %s", connType.Value().(string))
		}
	}
	log.Debug().Msg("Done processing connections.")
}
