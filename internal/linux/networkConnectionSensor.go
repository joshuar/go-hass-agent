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
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

const (
	dBusNMPath         = "/org/freedesktop/NetworkManager"
	dBusNMObj          = "org.freedesktop.NetworkManager"
	dBusObjConn        = "org.freedesktop.NetworkManager.Connection.Active"
	dBusObjAP          = "org.freedesktop.NetworkManager.AccessPoint"
	dBusObjDev         = "org.freedesktop.NetworkManager.Device"
	dBusObjWireless    = "org.freedesktop.NetworkManager.Device.Wireless"
	dBusObjIP4Cfg      = "org.freedesktop.NetworkManager.IP4Config"
	dBusObjIP6Cfg      = "org.freedesktop.NetworkManager.IP6Config"
	dBusPropConnType   = "org.freedesktop.NetworkManager.Connection.Active.Type"
	dBusPropConnID     = "org.freedesktop.NetworkManager.Connection.Active.Id"
	dBusPropConnState  = "org.freedesktop.NetworkManager.Connection.Active.State"
	dBusPropConnDevs   = "org.freedesktop.NetworkManager.Connection.Active.Devices"
	dBusPropConnAP     = "org.freedesktop.NetworkManager.Device.Wireless.ActiveAccessPoint"
	dBusPropConnIP4Cfg = "org.freedesktop.NetworkManager.Connection.Active.Ip4Config"
	dBusPropConnIP6Cfg = "org.freedesktop.NetworkManager.Connection.Active.Ip6Config"
	dBusPropIP4Addr    = "org.freedesktop.NetworkManager.IP4Config.AddressData"
	dBusPropIP6Addr    = "org.freedesktop.NetworkManager.IP6Config.AddressData"
	dBusPropAPSSID     = "org.freedesktop.NetworkManager.AccessPoint.Ssid"
	dBusPropAPFreq     = "org.freedesktop.NetworkManager.AccessPoint.Frequency"
	dBusPropAPSpd      = "org.freedesktop.NetworkManager.AccessPoint.MaxBitrate"
	dBusPropAPStr      = "org.freedesktop.NetworkManager.AccessPoint.Strength"
	dBusPropAPAddr     = "org.freedesktop.NetworkManager.AccessPoint.HwAddress"
)

// getActiveConns returns the list of active network connection D-Bus objects.
func getActiveConns() []dbus.ObjectPath {
	v, err := NewBusRequest(SystemBus).
		Path(dBusNMPath).
		Destination(dBusNMObj).
		GetProp(dBusNMObj + ".ActiveConnections")
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
// sensor.STATE_UNKNOWN.
func getConnType(p dbus.ObjectPath) string {
	if !p.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", p)
		return sensor.STATE_UNKNOWN
	}
	v, err := NewBusRequest(SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusPropConnType)
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch type of connection.")
		return sensor.STATE_UNKNOWN
	} else {
		return string(variantToValue[[]uint8](v))
	}
}

// getConnType extracts the connection name string from DBus for a given DBus
// path to a connection object. If it cannot find the name, it returns
// sensor.STATE_UNKNOWN.
func getConnName(p dbus.ObjectPath) string {
	if !p.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", p)
		return sensor.STATE_UNKNOWN
	}
	variant, err := NewBusRequest(SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusPropConnID)
	if err != nil {
		log.Error().Err(err).
			Msg("Could not fetch connection ID")
		return sensor.STATE_UNKNOWN
	}
	return string(variantToValue[[]uint8](variant))
}

// getConnState extracts and interprets the connection state string from DBus
// for a given DBus path to a connection object. If it cannot determine the
// state, it returns sensor.STATE_UNKNOWN.
func getConnState(p dbus.ObjectPath) string {
	state, err := NewBusRequest(SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusPropConnState)
	if err != nil {
		log.Error().Err(err).
			Msgf("Invalid connection state %v for network connection %s.", state.Value(), p)
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
	return sensor.STATE_UNKNOWN
}

// getConnDevice returns the D-Bus object representing the device for the given connection object.
func getConnDevice(p dbus.ObjectPath) dbus.ObjectPath {
	variant, err := NewBusRequest(SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusPropConnDevs)
	if err != nil {
		log.Error().Err(err).
			Msgf("Unable to fetch device for connection %s.", p)
		return sensor.STATE_UNKNOWN
	} else {
		// ! this conversion might yield unexpected results
		devicePath := variantToValue[[]dbus.ObjectPath](variant)[0]
		if !devicePath.IsValid() {
			log.Debug().Msgf("Invalid device path for connection %s.", p)
			return sensor.STATE_UNKNOWN
		}
		return devicePath
	}
}

// getIPAddr extracts the IP address for the given version (4/6) from the given
// DBus path to an address object. If it cannot find the address, it returns
// sensor.STATE_UNKNOWN.
func getIPAddr(p dbus.ObjectPath, ver int) string {
	if !p.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", p)
		return sensor.STATE_UNKNOWN
	}
	var connProp, addrProp string
	switch ver {
	case 4:
		connProp = dBusPropConnIP4Cfg
		addrProp = dBusPropIP4Addr
	case 6:
		connProp = dBusPropConnIP6Cfg
		addrProp = dBusPropIP6Addr
	}
	v, err := NewBusRequest(SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(connProp)
	if err != nil {
		return sensor.STATE_UNKNOWN
	}
	addrPath, ok := v.Value().(dbus.ObjectPath)
	if !ok {
		log.Debug().Msgf("Cannot process value recieved from D-Bus. Got %T.", addrPath)
		return sensor.STATE_UNKNOWN
	}
	v, err = NewBusRequest(SystemBus).
		Path(addrPath).
		Destination(dBusNMObj).
		GetProp(addrProp)
	if err != nil {
		return sensor.STATE_UNKNOWN
	}
	addrs, ok := v.Value().([]map[string]dbus.Variant)
	if !ok {
		log.Debug().Msgf("Cannot process value recieved from D-Bus. Got %T.", addrPath)
		return sensor.STATE_UNKNOWN
	}
	for _, a := range addrs {
		ip := net.ParseIP(a["address"].Value().(string))
		if ip.IsGlobalUnicast() {
			return ip.String()
		}
	}
	log.Debug().Msg("Could not ascertain IP address.")
	return sensor.STATE_UNKNOWN
}

// getWifiProp will fetch the appropriate value for the given wifi sensor type
// from D-Bus. If it cannot find the type, it returns sensor.STATE_UNKNOWN.
func getWifiProp(p dbus.ObjectPath, t sensorType) interface{} {
	if !p.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", p)
		return sensor.STATE_UNKNOWN
	}
	var prop string
	switch t {
	case wifiSSID:
		prop = dBusPropAPSSID
	case wifiFrequency:
		prop = dBusPropAPFreq
	case wifiSpeed:
		prop = dBusPropAPSpd
	case wifiStrength:
		prop = dBusPropAPStr
	case wifiHWAddress:
		prop = dBusPropAPAddr
	}
	var path dbus.ObjectPath
	ap, err := NewBusRequest(SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusPropConnAP)
	if err != nil {
		log.Debug().Err(err).Msg("Unable to retrieve wireless device details from D-Bus.")
		return sensor.STATE_UNKNOWN
	} else {
		path = dbus.ObjectPath((variantToValue[[]uint8](ap)))
		if !path.IsValid() {
			log.Debug().Msg("AP D-Bus Path is invalid")
			return sensor.STATE_UNKNOWN
		}
	}
	v, err := NewBusRequest(SystemBus).
		Path(path).
		Destination(dBusNMObj).
		GetProp(prop)
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch wifi property from D-Bus.")
		return dbus.MakeVariant(sensor.STATE_UNKNOWN)
	}
	switch t {
	case wifiSSID, wifiHWAddress:
		return string(variantToValue[[]uint8](v))
	case wifiFrequency, wifiSpeed, wifiStrength:
		return variantToValue[uint32](v)
	}

	return sensor.STATE_UNKNOWN
}

// handleConn will treat the given path as a connection object and create a
// sensor for the connection state, as well as any additional sensors for the
// connection type.
func handleConn(ctx context.Context, p dbus.ObjectPath, t device.SensorTracker) {
	s := newNetworkSensor(p, connectionState, nil)
	if s.sensorGroup == "lo" {
		log.Trace().Caller().Msgf("Ignoring state update for connection %s.", s.sensorGroup)
	} else {
		t.UpdateSensors(ctx, s)
	}
	extraSensors := make(chan *networkSensor)
	go func() {
		handleConnType(p, extraSensors)
	}()
	for p := range extraSensors {
		t.UpdateSensors(ctx, p)
	}
}

// handleConnType will treat the given path as a connection object and extra
// additional sensors based on the type of connection. If the connection type
// has no additional sensors available, nothing is returned.
func handleConnType(p dbus.ObjectPath, extraSensors chan *networkSensor) {
	defer close(extraSensors)
	connType := getConnType(p)
	switch connType {
	case "802-11-wireless":
		devicePath := getConnDevice(p)
		if devicePath == sensor.STATE_UNKNOWN {
			return
		}
		wifiProps := []sensorType{wifiSSID, wifiHWAddress, wifiFrequency, wifiSpeed, wifiStrength}
		for _, prop := range wifiProps {
			s := newNetworkSensor(devicePath, prop, nil)
			extraSensors <- s
		}
	case sensor.STATE_UNKNOWN:
		fallthrough
	default:
		log.Trace().Caller().Msgf("Unhandled connection type %s (%s).", connType, p)
	}
}

// handleProps will treat a list of properties as sensor updates, where the property names
// and values were recieved directly from D-Bus as a list.
func handleProps(ctx context.Context, p dbus.ObjectPath, props map[string]dbus.Variant, tracker device.SensorTracker) {
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
				Msgf("Unhandled property %v changed to %v (%s).", propName, propValue, p)
			return
		}
		tracker.UpdateSensors(ctx, newNetworkSensor(p, propType, value))
	}

}

type networkSensor struct {
	sensorGroup string
	objectPath  dbus.ObjectPath
	linuxSensor
}

func newNetworkSensor(p dbus.ObjectPath, asType sensorType, value interface{}) *networkSensor {
	s := &networkSensor{
		objectPath: p,
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
			s.value = getWifiProp(p, s.sensorType)
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
			DataSource:     SOURCE_DBUS,
		}
	}
	return struct {
		DataSource string `json:"Data Source"`
	}{
		DataSource: SOURCE_DBUS,
	}
}

func NetworkConnectionsUpdater(ctx context.Context, tracker device.SensorTracker) {
	connList := getActiveConns()
	if connList == nil {
		log.Error().Msg("No active connections.")
		return
	}
	for _, path := range connList {
		handleConn(ctx, path, tracker)
	}

	err := NewBusRequest(SystemBus).
		Path(dBusNMPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace(dBusNMPath),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(func(s *dbus.Signal) {
			if !s.Path.IsValid() || s.Path == "/" {
				log.Trace().Caller().Msgf("Invalid D-Bus object path %s.", s.Path)
				return
			}
			if len(s.Body) == 0 {
				log.Trace().Caller().Msg("No signal body recieved.")
				return
			}
			obj, ok := s.Body[0].(string)
			if !ok {
				log.Trace().Caller().Msgf("Unhandled signal body of type %T (%v).", obj, s)
				return
			}
			switch obj {
			case dBusNMObj:
				// TODO: handle this object
			case dBusObjConn:
				handleConn(ctx, s.Path, tracker)
			case dBusObjDev:
				// TODO: handle this object
			case dBusObjAP:
				fallthrough
			case dBusObjWireless:
				if updatedProps, ok := s.Body[1].(map[string]dbus.Variant); ok {
					handleProps(ctx, s.Path, updatedProps, tracker)
				}
			case dBusObjIP4Cfg:
				fallthrough
			case dBusObjIP6Cfg:
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
