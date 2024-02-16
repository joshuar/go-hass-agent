// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

var wifiProps = map[string]*wifiSensor{
	"Ssid": {
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorWifiSSID,
			IsDiagnostic:    true,
		},
	},
	"HwAddress": {
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorWifiHWAddress,
			IsDiagnostic:    true,
		},
	},
	"MaxBitrate": {
		Sensor: linux.Sensor{
			SensorTypeValue:  linux.SensorWifiSpeed,
			UnitsString:      "kB/s",
			DeviceClassValue: types.DeviceClassDataRate,
			StateClassValue:  types.StateClassMeasurement,
			IsDiagnostic:     true,
		},
	},
	"Frequency": {
		Sensor: linux.Sensor{
			SensorTypeValue:  linux.SensorWifiFrequency,
			UnitsString:      "MHz",
			DeviceClassValue: types.DeviceClassFrequency,
			StateClassValue:  types.StateClassMeasurement,
			IsDiagnostic:     true,
		},
	},
	"Strength": {
		Sensor: linux.Sensor{
			SensorTypeValue: linux.SensorWifiStrength,
			UnitsString:     "%",
			StateClassValue: types.StateClassMeasurement,
			IsDiagnostic:    true,
		},
	},
}

type wifiSensor struct {
	linux.Sensor
}

func (w *wifiSensor) State() any {
	switch w.SensorTypeValue {
	case linux.SensorWifiSSID:
		if value, ok := w.Value.([]uint8); ok {
			return string(value)
		} else {
			return sensor.StateUnknown
		}
	case linux.SensorWifiHWAddress:
		if value, ok := w.Value.(string); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	case linux.SensorWifiFrequency, linux.SensorWifiSpeed:
		if value, ok := w.Value.(uint32); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	case linux.SensorWifiStrength:
		if value, ok := w.Value.(uint8); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	default:
		return sensor.StateUnknown
	}
}

func (w *wifiSensor) Icon() string {
	switch w.SensorTypeValue {
	case linux.SensorWifiSSID, linux.SensorWifiHWAddress, linux.SensorWifiFrequency, linux.SensorWifiSpeed:
		return "mdi:wifi"
	case linux.SensorWifiStrength:
		value, ok := w.Value.(uint8)
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
func getWifiProperties(ctx context.Context, p dbus.ObjectPath) <-chan sensor.Details {
	var outCh []<-chan sensor.Details
	req := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
		Path(p).
		Destination(dBusNMObj)
	// get the devices associated with this connection
	wifiDevices, err := dbusx.GetProp[[]dbus.ObjectPath](req, dBusNMObj+".Connection.Active.Devices")
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve active wireless devices.")
		return nil
	}
	for _, d := range wifiDevices {
		// for each device, get the access point it is currently associated with
		ap, err := dbusx.GetProp[dbus.ObjectPath](req.Path(d), dBusNMObj+".Device.Wireless.ActiveAccessPoint")
		if err != nil {
			log.Warn().Err(err).Msg("Could not ascertain access point.")
			continue
		}
		propBase := dBusNMObj + ".AccessPoint"
		go func() {
			sensorCh := make(chan sensor.Details, 1)
			defer close(sensorCh)
			outCh = append(outCh, sensorCh)
			for k, p := range wifiProps {
				// for the associated access point, get the wifi properties as sensors
				value, err := dbusx.GetProp[any](req.Path(ap), propBase+"."+k)
				if err != nil {
					log.Warn().Err(err).Msgf("Could not get wifi property %s.", k)
					continue
				}
				p.Value = value
				wifiProps[k] = p
				sensorCh <- p
			}
		}()
		outCh = append(outCh, monitorWifiProperties(ctx, ap))
	}
	return sensor.MergeSensorCh(ctx, outCh...)
}

func monitorWifiProperties(ctx context.Context, p dbus.ObjectPath) chan sensor.Details {
	sensorCh := make(chan sensor.Details, 1)
	err := dbusx.NewBusRequest(ctx, dbusx.SystemBus).
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
							prop.Value = v.Value()
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
