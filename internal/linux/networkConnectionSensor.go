// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"net"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=dbusProp -output dbusConnProps.go -linecomment
const (
	networkManagerPath              = "/org/freedesktop/NetworkManager"
	networkManagerObject            = "org.freedesktop.NetworkManager"
	activeConnectionObject          = "org.freedesktop.NetworkManager.Connection.Active"
	accessPointObject               = "org.freedesktop.NetworkManager.AccessPoint"
	deviceObject                    = "org.freedesktop.NetworkManager.Device"
	wirelessDeviceObject            = "org.freedesktop.NetworkManager.Device.Wireless"
	ip4ConfigObject                 = "org.freedesktop.NetworkManager.IP4Config"
	ip6ConfigObject                 = "org.freedesktop.NetworkManager.IP6Config"
	connTypeProp           dbusProp = iota // org.freedesktop.NetworkManager.Connection.Active.Type
	connIDProp                             // org.freedesktop.NetworkManager.Connection.Active.Id
	connStateProp                          // org.freedesktop.NetworkManager.Connection.Active.State
	connDevicesProp                        // org.freedesktop.NetworkManager.Connection.Active.Devices
	wirelessDeviceProp                     // org.freedesktop.NetworkManager.Device.Wireless.ActiveAccessPoint
	connIPv4Prop                           // org.freedesktop.NetworkManager.Connection.Active.Ip4Config
	connIPv6Prop                           // org.freedesktop.NetworkManager.Connection.Active.Ip6Config
	addrIpv4Prop                           // org.freedesktop.NetworkManager.IP4Config.AddressData
	addrIpv6Prop                           // org.freedesktop.NetworkManager.IP6Config.AddressData
	wifiSSIDProp                           // org.freedesktop.NetworkManager.AccessPoint.Ssid
	wifiFreqProp                           // org.freedesktop.NetworkManager.AccessPoint.Frequency
	wifiSpeedProp                          // org.freedesktop.NetworkManager.AccessPoint.MaxBitrate
	wifiStrengthProp                       // org.freedesktop.NetworkManager.AccessPoint.Strength
	wifiHWAddrProp                         // org.freedesktop.NetworkManager.AccessPoint.HwAddress
)

type dbusProp int

// getConnType extracts the connection type string from DBus for a given DBus
// path to a connection object
func getConnType(path dbus.ObjectPath) string {
	if !path.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", path)
		return "Unknown"
	}
	v, err := NewBusRequest(SystemBus).
		Path(path).
		Destination(networkManagerObject).
		GetProp(connTypeProp.String())
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch type of connection.")
		return "Unknown"
	} else {
		return string(variantToValue[[]uint8](v))
	}
}

// getIPAddr extracts the IP address for the given version (4/6) from the given
// DBus path to an address object
func getIPAddr(connPath dbus.ObjectPath, ver int) string {
	if !connPath.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", connPath)
		return "Unknown"
	}
	var connProp, addrProp dbusProp
	switch ver {
	case 4:
		connProp = connIPv4Prop
		addrProp = addrIpv4Prop
	case 6:
		connProp = connIPv6Prop
		addrProp = addrIpv6Prop
	}
	v, err := NewBusRequest(SystemBus).
		Path(connPath).
		Destination(networkManagerObject).
		GetProp(connProp.String())
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch connection details.")
		return "Unknown"
	}
	addrPath, ok := v.Value().(dbus.ObjectPath)
	if !ok {
		log.Debug().Msgf("Cannot process value recieved from D-Bus. Got %T.", addrPath)
		return "Unknown"
	}
	v, err = NewBusRequest(SystemBus).
		Path(addrPath).
		Destination(networkManagerObject).
		GetProp(addrProp.String())
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch address details.")
		return "Unknown"
	}
	addrs, ok := v.Value().([]map[string]dbus.Variant)
	if !ok {
		log.Debug().Msgf("Cannot process value recieved from D-Bus. Got %T.", addrPath)
		return "Unknown"
	}
	for _, a := range addrs {
		ip := net.ParseIP(a["address"].Value().(string))
		if ip.IsGlobalUnicast() {
			return ip.String()
		}
	}
	log.Debug().Msg("Could not ascertain IP address.")
	return "Unknown"
}

// getWifiProp will fetch the appropriate value for the given wifi sensor type from D-Bus
func getWifiProp(path dbus.ObjectPath, t sensorType) dbus.Variant {
	if !path.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", path)
		return dbus.MakeVariant("Unknown")
	}
	var wifiProp dbusProp
	switch t {
	case wifiSSID:
		wifiProp = wifiSSIDProp
	case wifiFrequency:
		wifiProp = wifiFreqProp
	case wifiSpeed:
		wifiProp = wifiSpeedProp
	case wifiStrength:
		wifiProp = wifiStrengthProp
	case wifiHWAddress:
		wifiProp = wifiHWAddrProp
	}
	var apPath dbus.ObjectPath
	ap, err := NewBusRequest(SystemBus).
		Path(path).
		Destination(networkManagerObject).
		GetProp(wirelessDeviceProp.String())
	if err != nil {
		log.Debug().Err(err).Msg("Unable to retrieve wireless device details from D-Bus.")
		return dbus.MakeVariant("Unknown")
	} else {
		apPath = dbus.ObjectPath((variantToValue[[]uint8](ap)))
		if !apPath.IsValid() {
			log.Debug().Msg("AP D-Bus Path is invalid")
			return dbus.MakeVariant("Unknown")
		}
	}
	v, err := NewBusRequest(SystemBus).
		Path(apPath).
		Destination(networkManagerObject).
		GetProp(wifiProp.String())
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch wifi property from D-Bus.")
		return dbus.MakeVariant("Unknown")
	}
	return v
}

type networkSensor struct {
	sensorValue      interface{}
	sensorAttributes interface{}
	sensorGroup      string
	path             dbus.ObjectPath
	sensorType       sensorType
}

// networkSensor implements hass.SensorUpdate

func (s *networkSensor) Name() string {
	if s.sensorType == connectionState {
		return s.sensorGroup + " State"
	}
	return s.sensorType.String()
}

func (s *networkSensor) ID() string {
	if s.sensorType == connectionState {
		return strcase.ToSnake(s.sensorGroup) + "_connection_state"
	}
	return strcase.ToSnake(s.sensorType.String())
}

func (s *networkSensor) Icon() string {
	switch s.sensorType {
	case connectionState:
		switch s.sensorValue {
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
		switch s := s.sensorValue.(uint32); {
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

func (s *networkSensor) SensorType() sensor.SensorType {
	return sensor.TypeSensor
}

func (s *networkSensor) DeviceClass() sensor.SensorDeviceClass {
	switch s.sensorType {
	case wifiFrequency:
		return sensor.Frequency
	case wifiSpeed:
		return sensor.Data_rate
	default:
		return 0
	}
}

func (s *networkSensor) StateClass() sensor.SensorStateClass {
	switch s.sensorType {
	case wifiFrequency:
		fallthrough
	case wifiSpeed:
		fallthrough
	case wifiStrength:
		return sensor.StateMeasurement
	default:
		return 0
	}
}

func (s *networkSensor) State() interface{} {
	return s.sensorValue
}

func (s *networkSensor) Units() string {
	switch s.sensorType {
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

func (s *networkSensor) Category() string {
	return "diagnostic"
}

func (s *networkSensor) Attributes() interface{} {
	if s.sensorType == connectionState {
		return &struct {
			ConnectionType string `json:"Connection Type"`
			Ipv4           string `json:"IPv4 Address"`
			Ipv6           string `json:"IPv6 Address"`
			DataSource     string `json:"Data Source"`
		}{
			ConnectionType: getConnType(s.path),
			Ipv4:           getIPAddr(s.path, 4),
			Ipv6:           getIPAddr(s.path, 6),
			DataSource:     "D-Bus",
		}
	}
	return struct {
		DataSource string `json:"Data Source"`
	}{
		DataSource: "D-Bus",
	}
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

func marshalNetworkStateUpdate(t sensorType, dbusPath dbus.ObjectPath, g string, v dbus.Variant) *networkSensor {
	var value, attributes interface{}
	var path dbus.ObjectPath
	switch t {
	case connectionState:
		connState := variantToValue[uint32](v)
		value = stateToString(connState)
		path = dbusPath
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
		sensorGroup:      g,
		sensorType:       t,
		sensorValue:      value,
		sensorAttributes: attributes,
		path:             path,
	}
}

func NetworkConnectionsUpdater(ctx context.Context, status chan interface{}) {
	connList, err := NewBusRequest(SystemBus).
		Path(networkManagerPath).
		Destination(networkManagerObject).
		GetProp(networkManagerObject + ".ActiveConnections")
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve active connection list.")
		return
	}
	for _, conn := range connList.Value().([]dbus.ObjectPath) {
		processConnectionState(conn, status)
		processConnectionType(conn, status)
	}

	err = NewBusRequest(SystemBus).
		Path(networkManagerPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(networkManagerPath),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(func(s *dbus.Signal) {
			if len(s.Body) == 0 {
				return
			}
			obj, ok := s.Body[0].(string)
			if !ok {
				log.Debug().Msgf("Unhandled signal body of type %T.", obj)
				return
			}
			switch obj {
			case networkManagerObject:
				updatedProps := s.Body[1].(map[string]dbus.Variant)
				if _, ok := updatedProps["ActiveConnections"]; ok {
					log.Trace().Caller().
						Msg("Processing active connections changes.")
					// for _, conn := range c.Value().([]dbus.ObjectPath) {
					// 	processConnectionState(ctx, conn, status)
					// 	processConnectionType(ctx, conn, status)
					// }
				}
			case activeConnectionObject:
				log.Trace().Caller().
					Msgf("Processing active connections %s.", s.Path)
				processConnectionState(s.Path, status)
				processConnectionType(s.Path, status)
			case deviceObject:
				updatedProps := s.Body[1].(map[string]dbus.Variant)
				if c, ok := updatedProps["ActiveConnection"]; ok {
					log.Trace().Caller().
						Msgf("Processing device connection update %s.", c.String())
					// processConnectionState(
					// 	ctx, c.Value().(dbus.ObjectPath), status)
					// processConnectionType(
					// 	ctx, c.Value().(dbus.ObjectPath), status)
				}
			case accessPointObject:
				fallthrough
			case wirelessDeviceObject:
				updatedProps := s.Body[1].(map[string]dbus.Variant)
				processWifiProps(ctx, updatedProps, s.Path, status)
			case ip4ConfigObject:
				fallthrough
			case ip6ConfigObject:
				device := ipConfigToDevice(ctx, s.Path, obj)
				log.Trace().Caller().
					Msgf("Device %s was updated.", device)
				// connection := deviceToConnection(ctx, device)
			case "org.freedesktop.NetworkManager.Device.Statistics":
				// no-op, too noisy
			default:
				log.Trace().Caller().Msgf("Unhandled signal %v", s)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create network connections DBus watch.")
	}
}

func processConnectionState(path dbus.ObjectPath, status chan interface{}) {
	if !path.IsValid() || path == "/" {
		log.Debug().Msgf("Invalid D-Bus object path %s.", path)
		return
	}
	var variant dbus.Variant
	var err error
	variant, err = NewBusRequest(SystemBus).
		Path(path).
		Destination(networkManagerObject).
		GetProp(connIDProp.String())
	if err != nil {
		log.Error().Err(err).
			Msgf("Could not fetch properties of network connection %s.", path)
	} else {
		name := string(variantToValue[[]uint8](variant))
		if name == "lo" {
			return
		}
		variant, err = NewBusRequest(SystemBus).
			Path(path).
			Destination(networkManagerObject).
			GetProp(connStateProp.String())
		if err != nil {
			log.Error().Err(err).
				Msgf("Invalid connection state %v for network connection %s.", variant.Value(), path)
		} else {
			connState := marshalNetworkStateUpdate(connectionState, path, name, variant)
			status <- connState
		}
	}
}

func processConnectionType(path dbus.ObjectPath, status chan interface{}) {
	var variant dbus.Variant
	var err error
	connType := getConnType(path)
	switch connType {
	case "802-11-wireless":
		variant, err = NewBusRequest(SystemBus).
			Path(path).
			Destination(networkManagerObject).
			GetProp(connDevicesProp.String())
		if err != nil {
			log.Error().Err(err).
				Msgf("Unable to fetch device for connection %s.", path)
		} else {
			// ! this conversion might yield unexpected results
			devicePath := variantToValue[[]dbus.ObjectPath](variant)[0]
			if devicePath.IsValid() {
				wifiProps := []sensorType{wifiSSID, wifiHWAddress, wifiFrequency, wifiSpeed, wifiStrength}
				for _, prop := range wifiProps {
					propState := marshalNetworkStateUpdate(
						prop,
						devicePath,
						"wifi",
						getWifiProp(devicePath, prop))
					status <- propState
				}
			}
		}
	case "Unknown":
		fallthrough
	default:
		log.Trace().Caller().Msgf("Unhandled connection type %s (%s).", connType, path)
	}
}

func processWifiProps(ctx context.Context, props map[string]dbus.Variant, path dbus.ObjectPath, status chan interface{}) {
	if path.IsValid() {
		for propName, propValue := range props {
			var propType sensorType
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
				log.Trace().Caller().
					Msgf("Unhandled property %v changed to %v (%s).", propName, propValue, path)
			}
			if propType != 0 {
				propState := marshalNetworkStateUpdate(
					propType,
					path,
					"wifi",
					propValue)
				status <- propState
			}
		}
	}
}

func deviceToConnection(ctx context.Context, networkDevicePath dbus.ObjectPath) dbus.ObjectPath {
	variant, err := NewBusRequest(SystemBus).
		Path(networkDevicePath).
		Destination(networkManagerObject).
		GetProp(networkManagerObject + ".ActiveConnection")
	conn := dbus.ObjectPath(variantToValue[[]uint8](variant))
	if err != nil || !conn.IsValid() {
		return ""
	} else {
		return conn
	}
}

func ipConfigToDevice(ctx context.Context, ipConfigPath dbus.ObjectPath, ipConfigType string) dbus.ObjectPath {
	var configProp string
	switch {
	case strings.Contains(ipConfigType, "IP4Config"):
		configProp = "Ip4Config"
	case strings.Contains(ipConfigType, "IP6Config"):
		configProp = "Ip6Config"
	}

	deviceList := NewBusRequest(SystemBus).
		Path(networkManagerPath).
		Destination(networkManagerObject).
		GetData(networkManagerObject + ".GetDevices").
		AsObjectPathList()
	if deviceList == nil {
		log.Error().
			Msg("Could not list devices from network manager.")
		return ""
	}
	if len(deviceList) > 0 {
		for _, devicePath := range deviceList {
			c, err := NewBusRequest(SystemBus).
				Path(devicePath).
				Destination(networkManagerObject).
				GetProp(deviceObject + "." + configProp)
			if err != nil {
				log.Error().Err(err).
					Msg("Could not retrieve device config.")
			}
			deviceConfig := string(variantToValue[[]uint8](c))
			if deviceConfig == string(ipConfigPath) {
				return devicePath
			}
		}
	}
	return ""
}
