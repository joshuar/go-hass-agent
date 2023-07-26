// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

const (
	networkManagerPath     = "/org/freedesktop/NetworkManager"
	networkManagerObject   = "org.freedesktop.NetworkManager"
	activeConnectionObject = "org.freedesktop.NetworkManager.Connection.Active"
	accessPointObject      = "org.freedesktop.NetworkManager.AccessPoint"
	deviceObject           = "org.freedesktop.NetworkManager.Device"
	wirelessDeviceObject   = "org.freedesktop.NetworkManager.Device.Wireless"
	ip4ConfigObject        = "org.freedesktop.NetworkManager.IP4Config"
	ip6ConfigObject        = "org.freedesktop.NetworkManager.IP6Config"
)

func (t sensorType) dbusProp() string {
	switch t {
	case connectionID:
		return activeConnectionObject + ".Id"
	case connectionState:
		return activeConnectionObject + ".State"
	case connectionType:
		return activeConnectionObject + ".Type"
	case connectionDevices:
		return activeConnectionObject + ".Devices"
	case connectionIPv4:
		return activeConnectionObject + ".Ip4Config"
	case connectionIPv6:
		return activeConnectionObject + ".Ip6Config"
	case addressIPv4:
		return ip4ConfigObject + ".AddressData"
	case addressIPv6:
		return ip6ConfigObject + ".AddressData"
	case wifiSSID:
		return accessPointObject + ".Ssid"
	case wifiFrequency:
		return accessPointObject + ".Frequency"
	case wifiSpeed:
		return accessPointObject + ".MaxBitrate"
	case wifiStrength:
		return accessPointObject + ".Strength"
	case wifiHWAddress:
		return accessPointObject + ".HwAddress"
	default:
		return ""
	}
}

func getNetProp(ctx context.Context, p dbus.ObjectPath, t sensorType) (dbus.Variant, error) {
	return NewBusRequest(SystemBus).
		Path(p).
		Destination(networkManagerObject).
		GetProp(t.dbusProp())
}

func getIPAddrProp(ctx context.Context, t sensorType, p dbus.ObjectPath) (string, error) {
	var addrProp sensorType
	switch t {
	case connectionIPv4:
		addrProp = addressIPv4
	case connectionIPv6:
		addrProp = addressIPv6
	default:
		return "", errors.New("unknown address property")
	}
	if !p.IsValid() {
		return "", errors.New("invalid DBus path")
	}
	v, err := getNetProp(ctx, p, t)
	if err != nil {
		return "", err
	}
	switch configPath := v.Value().(type) {
	case dbus.ObjectPath:
		v, err := getNetProp(ctx, configPath, addrProp)
		if err != nil {
			return "", err
		}
		switch v.Value().(type) {
		case []map[string]dbus.Variant:
			addrs := v.Value().([]map[string]dbus.Variant)
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
	sensorValue      interface{}
	sensorAttributes interface{}
	sensorGroup      string
	sensorType       sensorType
}

// networkSensor implements hass.SensorUpdate

func (state *networkSensor) Name() string {
	if state.sensorType == connectionState {
		return state.sensorGroup + " State"
	}
	return state.sensorType.String()
}

func (state *networkSensor) ID() string {
	if state.sensorType == connectionState {
		return strcase.ToSnake(state.sensorGroup) + "_connection_state"
	}
	return strcase.ToSnake(state.sensorType.String())
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

func (state *networkSensor) SensorType() sensor.SensorType {
	return sensor.TypeSensor
}

func (state *networkSensor) DeviceClass() sensor.SensorDeviceClass {
	switch state.sensorType {
	case wifiFrequency:
		return sensor.Frequency
	case wifiSpeed:
		return sensor.Data_rate
	default:
		return 0
	}
}

func (state *networkSensor) StateClass() sensor.SensorStateClass {
	switch state.sensorType {
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

func marshalNetworkStateUpdate(ctx context.Context, t sensorType, p dbus.ObjectPath, g string, v dbus.Variant) *networkSensor {
	var value, attributes interface{}
	switch t {
	case connectionState:
		connState := variantToValue[uint32](v)
		value = stateToString(connState)
		connTypeVariant, err := getNetProp(ctx, p, connectionType)
		var connType string
		if err != nil {
			connType = "Unknown"
		} else {
			connType = string(variantToValue[[]uint8](connTypeVariant))
		}
		var ip4Addr, ip6Addr, addr string
		addr, err = getIPAddrProp(ctx, connectionIPv4, p)
		if err != nil {
			ip4Addr = ""
		} else {
			ip4Addr = addr
		}
		addr, err = getIPAddrProp(ctx, connectionIPv6, p)
		if err != nil {
			ip6Addr = ""
		} else {
			ip6Addr = addr
		}
		attributes = &struct {
			ConnectionType string `json:"Connection Type"`
			Ipv4           string `json:"IPv4 Address"`
			Ipv6           string `json:"IPv6 Address"`
			DataSource     string `json:"Data Source"`
		}{
			ConnectionType: connType,
			Ipv4:           ip4Addr,
			Ipv6:           ip6Addr,
			DataSource:     "D-Bus",
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
	if attributes == nil {
		attributes = struct {
			DataSource string `json:"Data Source"`
		}{
			DataSource: "D-Bus",
		}
	}
	return &networkSensor{
		sensorGroup:      g,
		sensorType:       t,
		sensorValue:      value,
		sensorAttributes: attributes,
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
		processConnectionState(ctx, conn, status)
		processConnectionType(ctx, conn, status)
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
			switch obj := s.Body[0].(type) {
			case string:
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
					// spew.Dump(s.Body)
				case activeConnectionObject:
					log.Trace().Caller().
						Msgf("Processing active connections %s.", s.Path)
					processConnectionState(ctx, s.Path, status)
					processConnectionType(ctx, s.Path, status)
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
				case "org.freedesktop.NetworkManager.DnsManager":
					// no-op
				case "org.freedesktop.NetworkManager.Device.Statistics":
					// no-op
					// default:
					// 	spew.Dump(s)
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create network connections DBus watch.")
	}
}

func processConnectionState(ctx context.Context, conn dbus.ObjectPath, status chan interface{}) {
	var variant dbus.Variant
	var err error
	variant, err = getNetProp(ctx, conn, connectionID)
	if err != nil {
		log.Error().Err(err).
			Msgf("Could not fetch properties of network connection %s.", conn)
	} else {
		name := string(variantToValue[[]uint8](variant))
		if conn != "/" && name != "lo" {
			variant, err = getNetProp(ctx, conn, connectionState)
			if err != nil {
				log.Error().Err(err).
					Msgf("Invalid connection state %v for network connection %s.", variant.Value(), conn)
			} else {
				connState := marshalNetworkStateUpdate(ctx, connectionState, conn, name, variant)
				status <- connState
			}
		}
	}
}

func processConnectionType(ctx context.Context, conn dbus.ObjectPath, status chan interface{}) {
	getWifiProp := func(p dbus.ObjectPath, t sensorType) (dbus.Variant, error) {
		var apPath dbus.ObjectPath
		ap, err := NewBusRequest(SystemBus).
			Path(p).
			Destination(networkManagerObject).
			GetProp(wirelessDeviceObject + ".ActiveAccessPoint")
		if err != nil {
			return dbus.MakeVariant(""), err
		} else {
			apPath = dbus.ObjectPath((variantToValue[[]uint8](ap)))
			if !apPath.IsValid() {
				return dbus.MakeVariant(""), errors.New("AP DBus Path is invalid")
			}
		}

		return NewBusRequest(SystemBus).
			Path(apPath).
			Destination(networkManagerObject).
			GetProp(t.dbusProp())
	}

	var variant dbus.Variant
	var err error
	variant, err = getNetProp(ctx, conn, connectionType)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to fetch connection type for %s.", conn)
	} else {
		connType := string(variantToValue[[]uint8](variant))
		switch connType {
		case "802-11-wireless":
			variant, err = getNetProp(ctx, conn, connectionDevices)
			if err != nil {
				log.Error().Err(err).
					Msgf("Unable to fetch device for connection %s.", conn)
			} else {
				// ! this conversion might yield unexpected results
				devicePath := variantToValue[[]dbus.ObjectPath](variant)[0]
				if devicePath.IsValid() {
					wifiProps := []sensorType{wifiSSID, wifiHWAddress, wifiFrequency, wifiSpeed, wifiStrength}
					for _, prop := range wifiProps {
						propValue, err := getWifiProp(devicePath, prop)
						if err != nil {
							log.Warn().Err(err).
								Msgf("Invalid wifi property %s for connection %s.", prop.String(), conn)
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
				propState := marshalNetworkStateUpdate(ctx,
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
