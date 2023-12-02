// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
)

var wifiProps = map[string]*wifiSensor{
	"Ssid": {
		linuxSensor: linuxSensor{
			sensorType: wifiSSID,
			diagnostic: true,
		},
	},
	"HwAddress": {
		linuxSensor: linuxSensor{
			sensorType: wifiHWAddress,
			diagnostic: true,
		},
	},
	"MaxBitrate": {
		linuxSensor: linuxSensor{
			sensorType:  wifiSpeed,
			units:       "kB/s",
			deviceClass: sensor.Data_rate,
			stateClass:  sensor.StateMeasurement,
			diagnostic:  true,
		},
	},
	"Frequency": {
		linuxSensor: linuxSensor{
			sensorType:  wifiFrequency,
			units:       "MHz",
			deviceClass: sensor.Frequency,
			stateClass:  sensor.StateMeasurement,
			diagnostic:  true,
		},
	},
	"Strength": {
		linuxSensor: linuxSensor{
			sensorType: wifiStrength,
			units:      "%",
			stateClass: sensor.StateMeasurement,
			diagnostic: true,
		},
	},
}

type wifiSensor struct {
	linuxSensor
}

func (w *wifiSensor) State() interface{} {
	switch w.sensorType {
	case wifiSSID:
		return string(w.value.([]uint8))
	case wifiHWAddress:
		return w.value.(string)
	case wifiFrequency, wifiSpeed:
		return w.value.(uint32)
	case wifiStrength:
		return w.value.(uint8)
	default:
		return sensor.StateUnknown
	}
}

func (w *wifiSensor) Icon() string {
	switch w.sensorType {
	case wifiSSID, wifiHWAddress, wifiFrequency, wifiSpeed:
		return "mdi:wifi"
	case wifiStrength:
		switch s := w.value.(uint8); {
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

// getWifiProperties will initially fetch and then monitor for changes of
// relevant WiFi properties that are to be represented as sensors.
func getWifiProperties(ctx context.Context, updateCh chan interface{}, p dbus.ObjectPath) {
	// get the devices associated with this connection
	v, _ := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Path(p).
		Destination(dBusNMObj).
		GetProp(dBusNMObj + ".Connection.Active.Devices")
	if !v.Signature().Empty() {
		for _, d := range dbushelpers.VariantToValue[[]dbus.ObjectPath](v) {
			// for each device, get the access point it is currently associated with
			v, _ := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
				Path(d).
				Destination(dBusNMObj).
				GetProp(dBusNMObj + ".Device.Wireless.ActiveAccessPoint")
			ap := dbushelpers.VariantToValue[dbus.ObjectPath](v)
			if !v.Signature().Empty() {
				r := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
					Path(ap).
					Destination(dBusNMObj)
				propBase := dBusNMObj + ".AccessPoint"
				for k, p := range wifiProps {
					// for the associated access point, get the wifi properties as sensors
					v, _ = r.GetProp(propBase + "." + k)
					if !v.Signature().Empty() {
						p.value = v.Value()
						wifiProps[k] = p
						updateCh <- p
					}
				}
				monitorWifiProperties(ctx, updateCh, ap)
			}
		}
	}
}

func monitorWifiProperties(ctx context.Context, updateCh chan interface{}, p dbus.ObjectPath) {
	err := dbushelpers.NewBusRequest(ctx, dbushelpers.SystemBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(p),
		}).
		Handler(func(s *dbus.Signal) {
			if len(s.Body) <= 1 {
				log.Debug().Caller().Interface("body", s.Body).Msg("Unexpected body length.")
				return
			}
			props, ok := s.Body[1].(map[string]dbus.Variant)
			if ok {
				for k, v := range props {
					prop, ok := wifiProps[k]
					if ok {
						prop.value = v.Value()
						updateCh <- prop
					}
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create WiFi property D-Bus watch.")
	}
}
