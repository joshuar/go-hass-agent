// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/pkg/dbushelpers"
	"github.com/rs/zerolog/log"
)

var wifiProps = map[string]*wifiSensor{
	"Ssid": {
		linuxSensor: linuxSensor{
			sensorType:   wifiSSID,
			isDiagnostic: true,
		},
	},
	"HwAddress": {
		linuxSensor: linuxSensor{
			sensorType:   wifiHWAddress,
			isDiagnostic: true,
		},
	},
	"MaxBitrate": {
		linuxSensor: linuxSensor{
			sensorType:   wifiSpeed,
			units:        "kB/s",
			deviceClass:  sensor.Data_rate,
			stateClass:   sensor.StateMeasurement,
			isDiagnostic: true,
		},
	},
	"Frequency": {
		linuxSensor: linuxSensor{
			sensorType:   wifiFrequency,
			units:        "MHz",
			deviceClass:  sensor.Frequency,
			stateClass:   sensor.StateMeasurement,
			isDiagnostic: true,
		},
	},
	"Strength": {
		linuxSensor: linuxSensor{
			sensorType:   wifiStrength,
			units:        "%",
			stateClass:   sensor.StateMeasurement,
			isDiagnostic: true,
		},
	},
}

type wifiSensor struct {
	linuxSensor
}

func (w *wifiSensor) State() any {
	switch w.sensorType {
	case wifiSSID:
		if value, ok := w.value.([]uint8); ok {
			return string(value)
		} else {
			return sensor.StateUnknown
		}
	case wifiHWAddress:
		if value, ok := w.value.(string); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	case wifiFrequency, wifiSpeed:
		if value, ok := w.value.(uint32); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	case wifiStrength:
		if value, ok := w.value.(uint8); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	default:
		return sensor.StateUnknown
	}
}

func (w *wifiSensor) Icon() string {
	switch w.sensorType {
	case wifiSSID, wifiHWAddress, wifiFrequency, wifiSpeed:
		return "mdi:wifi"
	case wifiStrength:
		value, ok := w.value.(uint8)
		if !ok {
			return "mdi:wifi-strength-alert-outline"
		}
		switch {
		case value <= 25:
			return "mdi:wifi-strength-1"
		case value > 25 && value <= 50:
			return "mdi:wifi-strength-2"
		case value > 50 && value <= 75:
			return "mdi:wifi-strength-3"
		case value > 75:
			return "mdi:wifi-strength-4"
		}
	}
	return "mdi:network"
}

// getWifiProperties will initially fetch and then monitor for changes of
// relevant WiFi properties that are to be represented as sensors.
func getWifiProperties(ctx context.Context, p dbus.ObjectPath) <-chan tracker.Sensor {
	var outCh []<-chan tracker.Sensor
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
				go func() {
					sensorCh := make(chan tracker.Sensor, 1)
					defer close(sensorCh)
					outCh = append(outCh, sensorCh)
					for k, p := range wifiProps {
						// for the associated access point, get the wifi properties as sensors
						v, _ = r.GetProp(propBase + "." + k)
						if !v.Signature().Empty() {
							p.value = v.Value()
							wifiProps[k] = p
							sensorCh <- p
						}
					}
				}()
				outCh = append(outCh, monitorWifiProperties(ctx, ap))
			}
		}
	}
	return tracker.MergeSensorCh(ctx, outCh...)
}

func monitorWifiProperties(ctx context.Context, p dbus.ObjectPath) chan tracker.Sensor {
	sensorCh := make(chan tracker.Sensor, 1)
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
				go func() {
					for k, v := range props {
						prop, ok := wifiProps[k]
						if ok {
							prop.value = v.Value()
							sensorCh <- prop
						}
					}
				}()
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to create WiFi property D-Bus watch.")
		close(sensorCh)
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
	}()
	return sensorCh
}
