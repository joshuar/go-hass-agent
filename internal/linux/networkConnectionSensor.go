// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"
	"net"

	"github.com/godbus/dbus/v5"
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

//go:generate stringer -type=dbusProp -output dbusConnProps.go -linecomment
const (
	networkManagerPath     = "/org/freedesktop/NetworkManager"
	networkManagerObject   = "org.freedesktop.NetworkManager"
	activeConnectionObject = "org.freedesktop.NetworkManager.Connection.Active"
	accessPointObject      = "org.freedesktop.NetworkManager.AccessPoint"
	deviceObject           = "org.freedesktop.NetworkManager.Device"
	wirelessDeviceObject   = "org.freedesktop.NetworkManager.Device.Wireless"
	ip4ConfigObject        = "org.freedesktop.NetworkManager.IP4Config"
	ip6ConfigObject        = "org.freedesktop.NetworkManager.IP6Config"
	connTypeProp           = "org.freedesktop.NetworkManager.Connection.Active.Type"
	connIDProp             = "org.freedesktop.NetworkManager.Connection.Active.Id"
	connStateProp          = "org.freedesktop.NetworkManager.Connection.Active.State"
	connDevicesProp        = "org.freedesktop.NetworkManager.Connection.Active.Devices"
	wirelessDeviceProp     = "org.freedesktop.NetworkManager.Device.Wireless.ActiveAccessPoint"
	connIPv4Prop           = "org.freedesktop.NetworkManager.Connection.Active.Ip4Config"
	connIPv6Prop           = "org.freedesktop.NetworkManager.Connection.Active.Ip6Config"
	addrIpv4Prop           = "org.freedesktop.NetworkManager.IP4Config.AddressData"
	addrIpv6Prop           = "org.freedesktop.NetworkManager.IP6Config.AddressData"
	wifiSSIDProp           = "org.freedesktop.NetworkManager.AccessPoint.Ssid"
	wifiFreqProp           = "org.freedesktop.NetworkManager.AccessPoint.Frequency"
	wifiSpeedProp          = "org.freedesktop.NetworkManager.AccessPoint.MaxBitrate"
	wifiStrengthProp       = "org.freedesktop.NetworkManager.AccessPoint.Strength"
	wifiHWAddrProp         = "org.freedesktop.NetworkManager.AccessPoint.HwAddress"
)

// getActiveConns returns the list of active network connection D-Bus objects.
func getActiveConns() []dbus.ObjectPath {
	v, err := NewBusRequest(SystemBus).
		Path(networkManagerPath).
		Destination(networkManagerObject).
		GetProp(networkManagerObject + ".ActiveConnections")
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not retrieve active connection list.")
		return nil
	}
	if l, ok := v.Value().([]dbus.ObjectPath); !ok {
		log.Warn().Msgf("Could not interpret active connection list, got %T", l)
	} else {
		return l
	}
	return nil
}

// getConnType extracts the connection type string from DBus for a given DBus
// path to a connection object. If it cannot find the type, it returns
// "Unknown".
func getConnType(path dbus.ObjectPath) string {
	if !path.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", path)
		return "Unknown"
	}
	v, err := NewBusRequest(SystemBus).
		Path(path).
		Destination(networkManagerObject).
		GetProp(connTypeProp)
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch type of connection.")
		return "Unknown"
	} else {
		return string(variantToValue[[]uint8](v))
	}
}

// getConnType extracts the connection name string from DBus for a given DBus
// path to a connection object. If it cannot find the name, it returns
// "Unknown".
func getConnName(path dbus.ObjectPath) string {
	if !path.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", path)
		return "Unknown"
	}
	variant, err := NewBusRequest(SystemBus).
		Path(path).
		Destination(networkManagerObject).
		GetProp(connIDProp)
	if err != nil {
		log.Error().Err(err).
			Msg("Could not fetch connection ID")
		return "Unknown"
	}
	return string(variantToValue[[]uint8](variant))
}

// getConnState extracts and interprets the connection state string from DBus
// for a given DBus path to a connection object. If it cannot determine the
// state, it returns "Unknown".
func getConnState(path dbus.ObjectPath) string {
	state, err := NewBusRequest(SystemBus).
		Path(path).
		Destination(networkManagerObject).
		GetProp(connStateProp)
	if err != nil {
		log.Error().Err(err).
			Msgf("Invalid connection state %v for network connection %s.", state.Value(), path)
	} else {
		switch state.Value().(uint32) {
		case 4:
			return "Offline"
		case 3:
			return "Deactivating"
		case 2:
			return "Online"
		case 1:
			return "Activating"
		}
	}
	return "Unknown"
}

// getConnDevice returns the D-Bus object representing the device for the given connection object.
func getConnDevice(path dbus.ObjectPath) dbus.ObjectPath {
	variant, err := NewBusRequest(SystemBus).
		Path(path).
		Destination(networkManagerObject).
		GetProp(connDevicesProp)
	if err != nil {
		log.Error().Err(err).
			Msgf("Unable to fetch device for connection %s.", path)
		return "Unknown"
	} else {
		// ! this conversion might yield unexpected results
		devicePath := variantToValue[[]dbus.ObjectPath](variant)[0]
		if !devicePath.IsValid() {
			log.Debug().Msgf("Invalid device path for connection %s.", path)
			return "Unknown"
		}
		return devicePath
	}
}

// getIPAddr extracts the IP address for the given version (4/6) from the given
// DBus path to an address object. If it cannot find the address, it returns
// "Unknown".
func getIPAddr(connPath dbus.ObjectPath, ver int) string {
	if !connPath.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", connPath)
		return "Unknown"
	}
	var connProp, addrProp string
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
		GetProp(connProp)
	if err != nil {
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
		GetProp(addrProp)
	if err != nil {
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

// getWifiProp will fetch the appropriate value for the given wifi sensor type
// from D-Bus. If it cannot find the type, it returns "Unknown".
func getWifiProp(path dbus.ObjectPath, t sensorType) interface{} {
	if !path.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", path)
		return "Unknown"
	}
	var wifiProp string
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
		GetProp(wirelessDeviceProp)
	if err != nil {
		log.Debug().Err(err).Msg("Unable to retrieve wireless device details from D-Bus.")
		return "Unknown"
	} else {
		apPath = dbus.ObjectPath((variantToValue[[]uint8](ap)))
		if !apPath.IsValid() {
			log.Debug().Msg("AP D-Bus Path is invalid")
			return "Unknown"
		}
	}
	v, err := NewBusRequest(SystemBus).
		Path(apPath).
		Destination(networkManagerObject).
		GetProp(wifiProp)
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch wifi property from D-Bus.")
		return dbus.MakeVariant("Unknown")
	}
	switch t {
	case wifiSSID, wifiHWAddress:
		return string(variantToValue[[]uint8](v))
	case wifiFrequency, wifiSpeed, wifiStrength:
		return variantToValue[uint32](v)
	}

	return "Unknown"
}

// handleConn will treat the given path as a connection object and create a
// sensor for the connection state, as well as any additional sensors for the
// connection type.
func handleConn(path dbus.ObjectPath, status chan interface{}) {
	s := newNetworkSensor(path, connectionState, nil)
	if s.sensorGroup == "lo" {
		log.Trace().Caller().Msgf("Ignoring state update for connection %s.", s.sensorGroup)
	} else {
		status <- s
	}
	extraSensors := make(chan *networkSensor)
	go func() {
		handleConnType(path, extraSensors)
	}()
	for p := range extraSensors {
		status <- p
	}
}

// handleConnType will treat the given path as a connection object and extra
// additional sensors based on the type of connection. If the connection type
// has no additional sensors available, nothing is returned.
func handleConnType(path dbus.ObjectPath, extraSensors chan *networkSensor) {
	defer close(extraSensors)
	connType := getConnType(path)
	switch connType {
	case "802-11-wireless":
		devicePath := getConnDevice(path)
		if devicePath == "Unknown" {
			return
		}
		wifiProps := []sensorType{wifiSSID, wifiHWAddress, wifiFrequency, wifiSpeed, wifiStrength}
		for _, prop := range wifiProps {
			s := newNetworkSensor(devicePath, prop, nil)
			extraSensors <- s
		}
	case "Unknown":
		fallthrough
	default:
		log.Trace().Caller().Msgf("Unhandled connection type %s (%s).", connType, path)
	}
}

// handleProps will treat a list of properties as sensor updates, where the property names
// and values were recieved directly from D-Bus as a list.
func handleProps(path dbus.ObjectPath, props map[string]dbus.Variant, status chan interface{}) {
	var propType sensorType
	var value interface{}
	for propName, propValue := range props {
		switch propName {
		case "Ssid":
			propType = wifiSSID
			value = variantToValue[[]uint8](propValue)
		case "HwAddress":
			propType = wifiHWAddress
			value = variantToValue[[]uint8](propValue)
		case "Frequency":
			propType = wifiFrequency
			value = variantToValue[uint32](propValue)
		case "Bitrate":
			propType = wifiSpeed
			value = variantToValue[uint32](propValue)
		case "Strength":
			propType = wifiStrength
			value = variantToValue[uint32](propValue)
		default:
			log.Trace().Caller().
				Msgf("Unhandled property %v changed to %v (%s).", propName, propValue, path)
			return
		}
		status <- newNetworkSensor(path, propType, value)
	}

}

type networkSensor struct {
	sensorGroup string
	objectPath  dbus.ObjectPath
	linuxSensor
}

func newNetworkSensor(path dbus.ObjectPath, asType sensorType, value interface{}) *networkSensor {
	s := &networkSensor{
		objectPath: path,
	}
	s.sensorType = asType
	s.diagnostic = true
	s.value = value
	switch s.sensorType {
	case connectionState:
		s.sensorGroup = getConnName(s.objectPath)
		if value == nil {
			s.value = getConnState(s.objectPath)
		}
	case wifiFrequency, wifiSSID, wifiSpeed, wifiHWAddress, wifiStrength:
		s.sensorGroup = "wifi"
		if value == nil {
			s.value = getWifiProp(path, s.sensorType)
		}
	}
	return s
}

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
		switch s.value {
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
		switch s := s.value.(uint32); {
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

func (s *networkSensor) Attributes() interface{} {
	if s.sensorType == connectionState {
		return &struct {
			ConnectionType string `json:"Connection Type"`
			Ipv4           string `json:"IPv4 Address"`
			Ipv6           string `json:"IPv6 Address"`
			DataSource     string `json:"Data Source"`
		}{
			ConnectionType: getConnType(s.objectPath),
			Ipv4:           getIPAddr(s.objectPath, 4),
			Ipv6:           getIPAddr(s.objectPath, 6),
			DataSource:     "D-Bus",
		}
	}
	return struct {
		DataSource string `json:"Data Source"`
	}{
		DataSource: "D-Bus",
	}
}

func NetworkConnectionsUpdater(ctx context.Context, status chan interface{}) {
	connList := getActiveConns()
	if connList == nil {
		log.Debug().Msg("No active connections.")
		return
	}
	for _, path := range connList {
		handleConn(path, status)
	}

	err := NewBusRequest(SystemBus).
		Path(networkManagerPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(networkManagerPath),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(func(s *dbus.Signal) {
			if !s.Path.IsValid() || s.Path == "/" {
				log.Debug().Msgf("Invalid D-Bus object path %s.", s.Path)
				return
			}
			if len(s.Body) == 0 {
				log.Debug().Msg("No signal body.")
				return
			}
			obj, ok := s.Body[0].(string)
			if !ok {
				log.Trace().Caller().Msgf("Unhandled signal body of type %T (%v).", obj, s)
				return
			}
			switch obj {
			case networkManagerObject:
				// TODO: handle this object
			case activeConnectionObject:
				handleConn(s.Path, status)
			case deviceObject:
				// TODO: handle this object
			case accessPointObject:
				fallthrough
			case wirelessDeviceObject:
				if updatedProps, ok := s.Body[1].(map[string]dbus.Variant); ok {
					handleProps(s.Path, updatedProps, status)
				}
			case ip4ConfigObject:
				fallthrough
			case ip6ConfigObject:
				// TODO: handle these objects
			case "org.freedesktop.NetworkManager.Device.Statistics":
				// no-op, too noisy
			default:
				log.Trace().Caller().Msgf("Unhandled object %s.", obj)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create network connections DBus watch.")
	}
}
