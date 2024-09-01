// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	accessPointInterface  = dBusNMObj + ".AccessPoint"
	activeAccessPointProp = dBusNMObj + ".Device.Wireless.ActiveAccessPoint"

	ssidPropName       = "Ssid"
	hwAddrPropName     = "HwAddress"
	maxBitRatePropName = "MaxBitrate"
	freqPropName       = "Frequency"
	strPropName        = "Strength"
	bandwidthPropName  = "Bandwidth"
)

var apPropList = []string{ssidPropName, hwAddrPropName, maxBitRatePropName, freqPropName, strPropName, bandwidthPropName}

type wifiSensor struct {
	prop string
	linux.Sensor
}

func (w *wifiSensor) State() any {
	switch w.prop {
	case ssidPropName:
		if value, ok := w.Value.([]uint8); ok {
			return string(value)
		} else {
			return sensor.StateUnknown
		}
	case hwAddrPropName:
		if value, ok := w.Value.(string); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	case freqPropName, maxBitRatePropName, bandwidthPropName:
		if value, ok := w.Value.(uint32); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	case strPropName:
		if value, ok := w.Value.(uint8); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	default:
		return sensor.StateUnknown
	}
}

//nolint:mnd
func (w *wifiSensor) Icon() string {
	switch w.prop {
	case ssidPropName, hwAddrPropName, freqPropName, maxBitRatePropName, bandwidthPropName:
		return "mdi:wifi"
	case strPropName:
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

func newWifiSensor(prop string, value any) *wifiSensor {
	wifiSensor := &wifiSensor{
		prop: prop,
		Sensor: linux.Sensor{
			IsDiagnostic: true,
			Value:        value,
		},
	}

	switch prop {
	case ssidPropName:
		wifiSensor.DisplayName = "Wi-Fi SSID"
	case hwAddrPropName:
		wifiSensor.DisplayName = "Wi-Fi BSSID"
	case maxBitRatePropName:
		wifiSensor.DisplayName = "Wi-Fi Link Speed"
		wifiSensor.UnitsString = "kB/s"
		wifiSensor.DeviceClassValue = types.DeviceClassDataRate
		wifiSensor.StateClassValue = types.StateClassMeasurement
	case freqPropName:
		wifiSensor.DisplayName = "Wi-Fi Frequency"
		wifiSensor.UnitsString = "MHz"
		wifiSensor.DeviceClassValue = types.DeviceClassFrequency
		wifiSensor.StateClassValue = types.StateClassMeasurement
	case bandwidthPropName:
		wifiSensor.DisplayName = "Wi-Fi Bandwidth"
		wifiSensor.UnitsString = "MHz"
		wifiSensor.DeviceClassValue = types.DeviceClassFrequency
		wifiSensor.StateClassValue = types.StateClassMeasurement
	case strPropName:
		wifiSensor.DisplayName = "Wi-Fi Signal Strength"
		wifiSensor.UnitsString = "%"
		wifiSensor.StateClassValue = types.StateClassMeasurement
	}

	return wifiSensor
}

func getWifiSensors(bus *dbusx.Bus, apPath string) []*wifiSensor {
	sensors := make([]*wifiSensor, 0, len(apPropList))

	for _, prop := range apPropList {
		value, err := dbusx.NewProperty[any](bus, apPath, dBusNMObj, accessPointInterface+"."+prop).Get()
		if err != nil {
			slog.Debug("Could not retrieve access point property.", slog.String("prop", prop), slog.Any("error", err))

			continue
		}

		sensors = append(sensors, newWifiSensor(prop, value))
	}

	return sensors
}
