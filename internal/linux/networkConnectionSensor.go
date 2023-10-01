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
	"golang.org/x/sync/errgroup"
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

type connAttributes struct {
	ConnectionType string `json:"Connection Type,omitempty"`
	Ipv4           string `json:"IPv4 Address,omitempty"`
	Ipv6           string `json:"IPv6 Address,omitempty"`
	DataSource     string `json:"Data Source"`
}

// getActiveConns returns the list of active network connection D-Bus objects.
func getActiveConns(ctx context.Context) []dbus.ObjectPath {
	v, err := NewBusRequest(ctx, SystemBus).
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
func getConnType(ctx context.Context, p dbus.ObjectPath) string {
	if !p.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", p)
		return sensor.StateUnknown
	}
	v, err := NewBusRequest(ctx, SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusPropConnType)
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch type of connection.")
		return sensor.StateUnknown
	} else {
		return string(variantToValue[[]uint8](v))
	}
}

// getConnType extracts the connection name string from DBus for a given DBus
// path to a connection object. If it cannot find the name, it returns
// sensor.STATE_UNKNOWN.
func getConnName(ctx context.Context, p dbus.ObjectPath) string {
	if !p.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", p)
		return sensor.StateUnknown
	}
	variant, err := NewBusRequest(ctx, SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusPropConnID)
	if err != nil {
		log.Error().Err(err).
			Msg("Could not fetch connection ID")
		return sensor.StateUnknown
	}
	return string(variantToValue[[]uint8](variant))
}

// getConnState extracts and interprets the connection state string from DBus
// for a given DBus path to a connection object. If it cannot determine the
// state, it returns sensor.STATE_UNKNOWN.
func getConnState(ctx context.Context, p dbus.ObjectPath) string {
	state, err := NewBusRequest(ctx, SystemBus).
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
	return sensor.StateUnknown
}

// getConnDevice returns the D-Bus object representing the device for the given connection object.
func getConnDevice(ctx context.Context, p dbus.ObjectPath) dbus.ObjectPath {
	variant, err := NewBusRequest(ctx, SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusPropConnDevs)
	if err != nil {
		log.Error().Err(err).
			Msgf("Unable to fetch device for connection %s.", p)
		return sensor.StateUnknown
	} else {
		// ! this conversion might yield unexpected results
		devicePath := variantToValue[[]dbus.ObjectPath](variant)[0]
		if !devicePath.IsValid() {
			log.Debug().Msgf("Invalid device path for connection %s.", p)
			return sensor.StateUnknown
		}
		return devicePath
	}
}

// getIPAddr extracts the IP address for the given version (4/6) from the given
// DBus path to an address object. If it cannot find the address, it returns
// sensor.STATE_UNKNOWN.
func getIPAddr(ctx context.Context, p dbus.ObjectPath, ver int) string {
	if !p.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", p)
		return sensor.StateUnknown
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
	v, err := NewBusRequest(ctx, SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(connProp)
	if err != nil {
		return sensor.StateUnknown
	}
	addrPath, ok := v.Value().(dbus.ObjectPath)
	if !ok {
		log.Debug().Msgf("Cannot process value received from D-Bus. Got %T.", addrPath)
		return sensor.StateUnknown
	}
	v, err = NewBusRequest(ctx, SystemBus).
		Path(addrPath).
		Destination(dBusNMObj).
		GetProp(addrProp)
	if err != nil {
		return sensor.StateUnknown
	}
	addrs, ok := v.Value().([]map[string]dbus.Variant)
	if !ok {
		log.Debug().Msgf("Cannot process value received from D-Bus. Got %T.", addrPath)
		return sensor.StateUnknown
	}
	for _, a := range addrs {
		ip := net.ParseIP(a["address"].Value().(string))
		if ip.IsGlobalUnicast() {
			return ip.String()
		}
	}
	return sensor.StateUnknown
}

// getWifiProp will fetch the appropriate value for the given wifi sensor type
// from D-Bus. If it cannot find the type, it returns sensor.STATE_UNKNOWN.
func getWifiProp(ctx context.Context, p dbus.ObjectPath, t sensorType) interface{} {
	if !p.IsValid() {
		log.Debug().Msgf("Invalid D-Bus object path %s.", p)
		return sensor.StateUnknown
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
	ap, err := NewBusRequest(ctx, SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusPropConnAP)
	if err != nil {
		log.Debug().Err(err).Msg("Unable to retrieve wireless device details from D-Bus.")
		return sensor.StateUnknown
	} else {
		path = dbus.ObjectPath((variantToValue[[]uint8](ap)))
		if !path.IsValid() {
			log.Debug().Msg("AP D-Bus Path is invalid")
			return sensor.StateUnknown
		}
	}
	v, err := NewBusRequest(ctx, SystemBus).
		Path(path).
		Destination(dBusNMObj).
		GetProp(prop)
	if err != nil {
		log.Debug().Err(err).Msg("Could not fetch wifi property from D-Bus.")
		return dbus.MakeVariant(sensor.StateUnknown)
	}
	switch t {
	case wifiSSID, wifiHWAddress:
		return string(variantToValue[[]uint8](v))
	case wifiFrequency, wifiSpeed, wifiStrength:
		return variantToValue[uint32](v)
	}

	return sensor.StateUnknown
}

// handleConn will treat the given path as a connection object and create a
// sensor for the connection state, as well as any additional sensors for the
// connection type.
func handleConn(ctx context.Context, p dbus.ObjectPath, t device.SensorTracker) error {
	s := newNetworkSensor(ctx, p, connectionState, nil)
	if s.sensorGroup == "lo" {
		log.Trace().Caller().Msgf("Ignoring state update for connection %s.", s.sensorGroup)
	} else {
		return t.UpdateSensors(ctx, s)
	}

	extraSensors := make(chan *networkSensor)
	go func() {
		handleConnType(ctx, p, extraSensors)
	}()

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		for p := range extraSensors {
			return t.UpdateSensors(ctx, p)
		}
		return nil
	})
	return g.Wait()
}

// handleConnType will treat the given path as a connection object and extra
// additional sensors based on the type of connection. If the connection type
// has no additional sensors available, nothing is returned.
func handleConnType(ctx context.Context, p dbus.ObjectPath, extraSensors chan *networkSensor) {
	defer close(extraSensors)
	connType := getConnType(ctx, p)
	switch connType {
	case "802-11-wireless":
		devicePath := getConnDevice(ctx, p)
		if devicePath == sensor.StateUnknown {
			return
		}
		wifiProps := []sensorType{wifiSSID, wifiHWAddress, wifiFrequency, wifiSpeed, wifiStrength}
		for _, prop := range wifiProps {
			s := newNetworkSensor(ctx, devicePath, prop, nil)
			extraSensors <- s
		}
	// case sensor.StateUnknown:
	// 	fallthrough
	default:
		log.Trace().Caller().Msgf("Unhandled connection type %s (%s).", connType, p)
	}
}

// handleProps will treat a list of properties as sensor updates, where the property names
// and values were received directly from D-Bus as a list.
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
		if err := tracker.UpdateSensors(ctx, newNetworkSensor(ctx, p, propType, value)); err != nil {
			log.Error().Err(err).Str("prop", propName).
				Msg("Could not update connection property")
		}
	}
}

type networkSensor struct {
	attributes  connAttributes
	sensorGroup string
	objectPath  dbus.ObjectPath
	linuxSensor
}

func newNetworkSensor(ctx context.Context, p dbus.ObjectPath, asType sensorType, value interface{}) *networkSensor {
	s := &networkSensor{
		objectPath: p,
	}
	s.sensorType = asType
	s.diagnostic = true
	s.value = value
	s.attributes.DataSource = srcDbus
	switch s.sensorType {
	case connectionState:
		s.sensorGroup = getConnName(ctx, s.objectPath)
		if value == nil {
			s.value = getConnState(ctx, s.objectPath)
		}
		s.attributes.Ipv4 = getIPAddr(ctx, s.objectPath, 4)
		s.attributes.Ipv6 = getIPAddr(ctx, s.objectPath, 6)
		s.attributes.ConnectionType = getConnType(ctx, s.objectPath)
	case wifiFrequency, wifiSSID, wifiSpeed, wifiHWAddress, wifiStrength:
		s.sensorGroup = "wifi"
		if value == nil {
			s.value = getWifiProp(ctx, p, s.sensorType)
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
	case wifiSSID, wifiHWAddress, wifiFrequency, wifiSpeed:
		return "mdi:wifi"
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
	case wifiFrequency, wifiSpeed, wifiStrength:
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
	return s.attributes
}

func NetworkConnectionsUpdater(ctx context.Context, tracker device.SensorTracker) {
	connList := getActiveConns(ctx)
	if connList == nil {
		log.Error().Msg("No active connections.")
		return
	}
	for _, path := range connList {
		if err := handleConn(ctx, path, tracker); err != nil {
			log.Error().Err(err).Str("dBusPath", string(path)).
				Msg("Could not process connection.")
		}
	}

	err := NewBusRequest(ctx, SystemBus).
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
				log.Trace().Caller().Msg("No signal body received.")
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
				if err := handleConn(ctx, s.Path, tracker); err != nil {
					log.Error().Err(err).Str("dBusPath", string(s.Path)).
						Msg("Could not process connection.")
				}
			case dBusObjDev:
				// TODO: handle this object
			case dBusObjAP, dBusObjWireless:
				if updatedProps, ok := s.Body[1].(map[string]dbus.Variant); ok {
					handleProps(ctx, s.Path, updatedProps, tracker)
				}
			case dBusObjIP4Cfg, dBusObjIP6Cfg:
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
