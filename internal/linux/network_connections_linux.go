// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"errors"
	"net"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=networkProp -output network_connections_props_linux.go

const (
	dBusDest = "org.freedesktop.NetworkManager"
	dBusPath = "/org/freedesktop/NetworkManager"

	connectionState networkProp = iota
	connectionID
	connectionDevices
	connectionType
	connectionIPv4
	connectionIPv6
	addressIPv4
	addressIPv6
	wifiSSID
	wifiFrequency
	wifiSpeed
	wifiStrength
	wifiHWAddress
)

type networkProp int

func getNetProp(ctx context.Context, path dbus.ObjectPath, prop networkProp) (dbus.Variant, error) {
	deviceAPI, _ := FetchAPIFromContext(ctx)

	connIntr := "org.freedesktop.NetworkManager.Connection.Active"
	ipv4Intr := "org.freedesktop.NetworkManager.IP4Config"
	ipv6Intr := "org.freedesktop.NetworkManager.IP6Config"

	var dbusProp string
	switch prop {
	case connectionID:
		dbusProp = connIntr + ".Id"
	case connectionState:
		dbusProp = connIntr + ".State"
	case connectionType:
		dbusProp = connIntr + ".Type"
	case connectionDevices:
		dbusProp = connIntr + ".Devices"
	case connectionIPv4:
		dbusProp = connIntr + ".Ip4Config"
	case connectionIPv6:
		dbusProp = connIntr + ".Ip6Config"
	case addressIPv4:
		dbusProp = ipv4Intr + ".AddressData"
	case addressIPv6:
		dbusProp = ipv6Intr + ".AddressData"
	default:
		return dbus.MakeVariant(""), errors.New("unknown network property")
	}
	return deviceAPI.SystemBusRequest().
		Path(path).
		Destination(dBusDest).
		GetProp(dbusProp)
}

func getWifiProp(ctx context.Context, path dbus.ObjectPath, wifiProp networkProp) (dbus.Variant, error) {
	wirelessIntr := "org.freedesktop.NetworkManager.Device.Wireless"
	apIntr := "org.freedesktop.NetworkManager.AccessPoint"

	deviceAPI, _ := FetchAPIFromContext(ctx)

	var apPath dbus.ObjectPath
	ap, err := deviceAPI.SystemBusRequest().
		Path(path).
		Destination(dBusDest).
		GetProp(wirelessIntr + ".ActiveAccessPoint")
	if err != nil {
		return dbus.MakeVariant(""), err
	} else {
		apPath = dbus.ObjectPath((variantToValue[[]uint8](ap)))
		if !apPath.IsValid() {
			return dbus.MakeVariant(""), errors.New("AP DBus Path is invalid")
		}
	}

	var dbusProp string
	switch wifiProp {
	case wifiSSID:
		dbusProp = apIntr + ".Ssid"
	case wifiFrequency:
		dbusProp = apIntr + ".Frequency"
	case wifiSpeed:
		dbusProp = apIntr + ".MaxBitrate"
	case wifiStrength:
		dbusProp = apIntr + ".Strength"
	case wifiHWAddress:
		dbusProp = apIntr + ".HwAddress"
	default:
		return dbus.MakeVariant(""), errors.New("unknown wifi property")
	}
	return deviceAPI.SystemBusRequest().
		Path(apPath).
		Destination(dBusDest).
		GetProp(dbusProp)
}

func getIPAddrProp(ctx context.Context, connProp networkProp, path dbus.ObjectPath) (string, error) {
	var addrProp networkProp
	switch connProp {
	case connectionIPv4:
		addrProp = addressIPv4
	case connectionIPv6:
		addrProp = addressIPv6
	default:
		return "", errors.New("unknown address property")
	}
	if !path.IsValid() {
		return "", errors.New("invalid DBus path")
	}
	p, err := getNetProp(ctx, path, connProp)
	if err != nil {
		return "", err
	}
	switch configPath := p.Value().(type) {
	case dbus.ObjectPath:
		propValue, err := getNetProp(ctx, configPath, addrProp)
		if err != nil {
			return "", err
		}
		switch propValue.Value().(type) {
		case []map[string]dbus.Variant:
			addrs := propValue.Value().([]map[string]dbus.Variant)
			for _, a := range addrs {
				ip := net.ParseIP(a["address"].Value().(string))
				if ip.IsGlobalUnicast() {
					return ip.String(), nil
				}
			}
		}
	}
	return "", errors.New("no address found")
}

type networkSensor struct {
	sensorGroup      string
	sensorType       networkProp
	sensorValue      interface{}
	sensorAttributes interface{}
}

// networkSensor implements hass.SensorUpdate

func (state *networkSensor) Name() string {
	switch state.sensorType {
	case connectionState:
		return state.sensorGroup + " State"
	case wifiSSID:
		return "Wi-Fi Connection"
	case wifiHWAddress:
		return "Wi-Fi BSSID"
	case wifiFrequency:
		return "Wi-Fi Frequency"
	case wifiSpeed:
		return "Wi-Fi Link Speed"
	case wifiStrength:
		return "Wi-Fi Signal Strength"
	default:
		prettySensorName := strcase.ToDelimited(state.sensorType.String(), ' ')
		log.Debug().Caller().
			Msgf("Unexpected sensor %s with type %s.",
				prettySensorName, state.sensorType.String())
		return state.sensorGroup + " " + prettySensorName
	}
}

func (state *networkSensor) ID() string {
	switch state.sensorType {
	case connectionState:
		return strcase.ToSnake(state.sensorGroup) + "_connection_state"
	case wifiSSID:
		return "wifi_connection"
	case wifiHWAddress:
		return "wifi_bssid"
	case wifiFrequency:
		return "wifi_frequency"
	case wifiSpeed:
		return "wifi_link_speed"
	case wifiStrength:
		return "wifi_signal_strength"
	default:
		snakeSensorName := strcase.ToSnake(state.sensorType.String())
		return strcase.ToSnake(state.sensorGroup) + "_" + snakeSensorName
	}
}

func (state *networkSensor) Icon() string {
	switch state.sensorType {
	case connectionState:
		switch state.sensorValue {
		case "Online":
			return "mdi:network"
		case "Offline":
			return "mdi:network-off"
		case "Activating":
			return "mdi:plus-network"
		case "Deactivating":
			return "mdi:minus-network"
		default:
			return "mdi:help-network"
		}
	case wifiSSID:
		fallthrough
	case wifiHWAddress:
		fallthrough
	case wifiFrequency:
		fallthrough
	case wifiSpeed:
		return "mdi:wifi"
	case wifiStrength:
		switch s := state.sensorValue.(uint32); {
		case s <= 25:
			return "mdi:wifi-strength-1"
		case s > 25 && s <= 50:
			return "mdi:wifi-strength-2"
		case s > 50 && s <= 75:
			return "mdi:wifi-strength-3"
		case s > 75:
			return "mdi:wifi-strength-4"
		}
	}
	return "mdi:network"
}

func (state *networkSensor) SensorType() hass.SensorType {
	return hass.TypeSensor
}

func (state *networkSensor) DeviceClass() hass.SensorDeviceClass {
	switch state.sensorType {
	case wifiFrequency:
		return hass.Frequency
	case wifiSpeed:
		return hass.Data_rate
	default:
		return 0
	}
}

func (state *networkSensor) StateClass() hass.SensorStateClass {
	switch state.sensorType {
	case wifiFrequency:
		fallthrough
	case wifiSpeed:
		fallthrough
	case wifiStrength:
		return hass.StateMeasurement
	default:
		return 0
	}
}

func (state *networkSensor) State() interface{} {
	return state.sensorValue
}

func (state *networkSensor) Units() string {
	switch state.sensorType {
	case wifiFrequency:
		return "MHz"
	case wifiSpeed:
		return "kB/s"
	case wifiStrength:
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

func marshalNetworkStateUpdate(ctx context.Context, sensor networkProp, path dbus.ObjectPath, group string, v dbus.Variant) *networkSensor {
	var value, attributes interface{}
	switch sensor {
	case connectionState:
		connState := variantToValue[uint32](v)
		value = stateToString(connState)
		connTypeVariant, err := getNetProp(ctx, path, connectionType)
		var connType string
		if err != nil {
			connType = "Unknown"
		} else {
			connType = string(variantToValue[[]uint8](connTypeVariant))
		}
		var ip4Addr, ip6Addr, addr string
		addr, err = getIPAddrProp(ctx, connectionIPv4, path)
		if err != nil {
			ip4Addr = ""
		} else {
			ip4Addr = addr
		}
		addr, err = getIPAddrProp(ctx, connectionIPv6, path)
		if err != nil {
			ip6Addr = ""
		} else {
			ip6Addr = addr
		}
		attributes = &struct {
			ConnectionType string `json:"Connection Type"`
			Ipv4           string `json:"IPv4 Address"`
			Ipv6           string `json:"IPv6 Address"`
		}{
			ConnectionType: connType,
			Ipv4:           ip4Addr,
			Ipv6:           ip6Addr,
		}
	case wifiSSID:
		value = string(variantToValue[[]uint8](v))
	case wifiHWAddress:
		value = string(variantToValue[[]uint8](v))
	case wifiFrequency:
		value = variantToValue[uint32](v)
	case wifiSpeed:
		value = variantToValue[uint32](v)
	case wifiStrength:
		value = variantToValue[uint32](v)
	}
	return &networkSensor{
		sensorGroup:      group,
		sensorType:       sensor,
		sensorValue:      value,
		sensorAttributes: attributes,
	}
}

func NetworkConnectionsUpdater(ctx context.Context, status chan interface{}) {
	deviceAPI, err := FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to DBus.")
		return
	}

	deviceList := deviceAPI.SystemBusRequest().
		Path(dBusPath).
		Destination(dBusDest).
		GetData("org.freedesktop.NetworkManager.GetDevices").
		AsObjectPathList()
	if deviceList == nil {
		log.Debug().Err(err).Caller().
			Msg("Could not list devices from network manager.")
		return
	}
	if len(deviceList) > 0 {
		for _, device := range deviceList {
			conn := deviceActiveConnection(ctx, device)
			if conn != "" {
				processConnectionState(ctx, conn, status)
				processConnectionType(ctx, conn, status)
			}
		}
	}

	// Set up a DBus watch for connection state changes
	activeConnDBusPath := dbus.ObjectPath(dBusPath + "/ActiveConnection")
	connStateDBusMatch := []dbus.MatchOption{
		dbus.WithMatchPathNamespace(activeConnDBusPath),
	}
	connStateHandler := func(s *dbus.Signal) {
		if s.Path.IsValid() {
			switch {
			case s.Name == "org.freedesktop.NetworkManager.Connection.Active.StateChanged":
				processConnectionState(ctx, s.Path, status)
				processConnectionType(ctx, s.Path, status)
			}
		}
	}
	// NewDBusWatchRequest().
	// 	System().
	deviceAPI.SystemBusRequest().
		Path(activeConnDBusPath).
		Match(connStateDBusMatch).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(connStateHandler).
		AddWatch()
		// Add(deviceAPI)

	// Set up a DBus watch for Wi-Fi state changes
	apDbusPath := dbus.ObjectPath(dBusPath + "/AccessPoint")
	wifiStateDBusMatch := []dbus.MatchOption{
		dbus.WithMatchPathNamespace(apDbusPath),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
	}
	wifiStateHandler := func(s *dbus.Signal) {
		if s.Path.IsValid() {
			updatedProps := s.Body[1].(map[string]dbus.Variant)
			for propName, propValue := range updatedProps {
				var propType networkProp
				switch propName {
				case "Ssid":
					propType = wifiSSID
				case "HwAddress":
					propType = wifiHWAddress
				case "Frequency":
					propType = wifiFrequency
				case "Bitrate":
					propType = wifiSpeed
				case "Strength":
					propType = wifiStrength
				default:
					log.Debug().Msgf("Unhandled property %v changed to %v", propName, propValue)
				}
				if propType != 0 {
					propState := marshalNetworkStateUpdate(ctx,
						propType,
						s.Path,
						"wifi",
						propValue)
					status <- propState
				}
			}
		}
	}
	deviceAPI.SystemBusRequest().
		Path(apDbusPath).
		Match(wifiStateDBusMatch).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(wifiStateHandler).
		AddWatch()

	// Add a DBus watch for global connectivity changes. If global connectivity
	// is established, check and update external IP sensor.
	// networkStateWatch := &DBusWatchRequest{
	// 	bus:  systemBus,
	// 	path: dBusPath,
	// 	match: []dbus.MatchOption{
	// 		dbus.WithMatchPathNamespace(dBusPath),
	// 		dbus.WithMatchInterface(dBusDest),
	// 	},
	// 	event: "org.freedesktop.NetworkManager.Statechanged",
	// 	eventHandler: func(s *dbus.Signal) {
	// 		switch state := s.Body[0].(type) {
	// 		case uint32:
	// 			if state == 70 {
	// 				device.UpdateExternalIPSensors(ctx, status)
	// 			}
	// 		}
	// 	},
	// }
	// deviceAPI.WatchEvents <- networkStateWatch

	// catchAllWatch := &DBusWatchRequest{
	// 	bus:  systemBus,
	// 	path: "/org/freedesktop/NetworkManager",
	// 	match: []dbus.MatchOption{
	// 		dbus.WithMatchPathNamespace("/org/freedesktop/NetworkManager"),
	// 	},
	// 	event: "org.freedesktop.DBus.Properties.PropertiesChanged",
	// 	eventHandler: func(s *dbus.Signal) {
	// 		switch prop := s.Body[0].(type) {
	// 		case string:
	// 			propsChanged := s.Body[1].(map[string]dbus.Variant)
	// 			switch prop {
	// 			case "org.freedesktop.NetworkManager":
	// 				if connList, ok := propsChanged["ActiveConnections"]; ok {
	// 					spew.Dump(connList)
	// 				}
	// 			case "org.freedesktop.NetworkManager.Device.Statistics":
	// 				// no-op
	// 			case "org.freedesktop.NetworkManager.AccessPoint":
	// 				// no-op
	// 			default:
	// 				spew.Dump(s)
	// 			}
	// 		}
	// 	},
	// }
	// deviceAPI.WatchEvents <- catchAllWatch

}

func deviceActiveConnection(ctx context.Context, networkDevicePath dbus.ObjectPath) dbus.ObjectPath {
	deviceAPI, err := FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to DBus.")
		return ""
	}

	variant, err := deviceAPI.SystemBusRequest().
		Path(networkDevicePath).
		Destination(dBusDest).
		GetProp("org.freedesktop.NetworkManager.Device.ActiveConnection")
	conn := dbus.ObjectPath(variantToValue[[]uint8](variant))
	if err != nil || !conn.IsValid() {
		return ""
	} else {
		return conn
	}
}

func processConnectionState(ctx context.Context, conn dbus.ObjectPath, status chan interface{}) {
	var variant dbus.Variant
	var err error
	variant, err = getNetProp(ctx, conn, connectionID)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msgf("Invalid connection %s", conn)
	} else {
		name := string(variantToValue[[]uint8](variant))
		if conn != "/" && name != "lo" {
			variant, err = getNetProp(ctx, conn, connectionState)
			if err != nil {
				log.Debug().Err(err).Caller().
					Msgf("Invalid connection state %v.", variant.Value())
			} else {
				connState := marshalNetworkStateUpdate(ctx, connectionState, conn, name, variant)
				status <- connState
			}
		}
	}
}

func processConnectionType(ctx context.Context, conn dbus.ObjectPath, status chan interface{}) {
	var variant dbus.Variant
	var err error
	variant, err = getNetProp(ctx, conn, connectionType)
	if err != nil {
		log.Debug().Err(err).Msg("Invalid connection type.")
	} else {
		connType := string(variantToValue[[]uint8](variant))
		switch connType {
		case "802-11-wireless":
			variant, err = getNetProp(ctx, conn, connectionDevices)
			if err != nil {
				log.Debug().Err(err).Caller().
					Msg("Invalid connection device.")
			} else {
				// ! this conversion might yield unexpected results
				devicePath := variantToValue[[]dbus.ObjectPath](variant)[0]
				if devicePath.IsValid() {
					wifiProps := []networkProp{wifiSSID, wifiHWAddress, wifiFrequency, wifiSpeed, wifiStrength}
					for _, prop := range wifiProps {
						propValue, err := getWifiProp(ctx, devicePath, prop)
						if err != nil {
							log.Debug().Err(err).Caller().
								Msg("Invalid wifi property.")
						} else {
							propState := marshalNetworkStateUpdate(ctx,
								prop,
								devicePath,
								"wifi",
								propValue)
							status <- propState
						}
					}
				}
			}
		}
	}
}
