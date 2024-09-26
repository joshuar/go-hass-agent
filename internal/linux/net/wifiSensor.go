// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
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

func newWifiSensor(prop string, value any) sensor.Entity {
	wifiSensor := sensor.Entity{
		Category: types.CategoryDiagnostic,
		EntityState: &sensor.EntityState{
			State: generateState(prop, value),
			Icon:  "mdi:wifi",
		},
	}

	switch prop {
	case ssidPropName:
		wifiSensor.Name = "Wi-Fi SSID"
		wifiSensor.ID = "wi_fi_ssid"
	case hwAddrPropName:
		wifiSensor.Name = "Wi-Fi BSSID"
		wifiSensor.ID = "wi_fi_bssid"
	case maxBitRatePropName:
		wifiSensor.Name = "Wi-Fi Link Speed"
		wifiSensor.ID = "wi_fi_link_speed"
		wifiSensor.Units = "kB/s"
		wifiSensor.DeviceClass = types.DeviceClassDataRate
		wifiSensor.StateClass = types.StateClassMeasurement
	case freqPropName:
		wifiSensor.Name = "Wi-Fi Frequency"
		wifiSensor.ID = "wi_fi_frequency"
		wifiSensor.Units = "MHz"
		wifiSensor.DeviceClass = types.DeviceClassFrequency
		wifiSensor.StateClass = types.StateClassMeasurement
	case bandwidthPropName:
		wifiSensor.Name = "Wi-Fi Bandwidth"
		wifiSensor.ID = "wi_fi_bandwidth"
		wifiSensor.Units = "MHz"
		wifiSensor.DeviceClass = types.DeviceClassFrequency
		wifiSensor.StateClass = types.StateClassMeasurement
	case strPropName:
		wifiSensor.Name = "Wi-Fi Signal Strength"
		wifiSensor.ID = "wi_fi_signal_strength"
		wifiSensor.Units = "%"
		wifiSensor.StateClass = types.StateClassMeasurement
		wifiSensor.Icon = generateStrIcon(value)
	}

	return wifiSensor
}

func getWifiSensors(bus *dbusx.Bus, apPath string) []sensor.Entity {
	sensors := make([]sensor.Entity, 0, len(apPropList))

	for _, prop := range apPropList {
		value, err := dbusx.NewProperty[any](bus, apPath, dBusNMObj, accessPointInterface+"."+prop).Get()
		if err != nil {
			slog.Debug("Could not retrieve access point property.",
				slog.String("prop", prop),
				slog.Any("error", err))

			continue
		}

		sensors = append(sensors, newWifiSensor(prop, value))
	}

	return sensors
}

func generateState(prop string, value any) any {
	switch prop {
	case ssidPropName:
		if value, ok := value.([]uint8); ok {
			return string(value)
		} else {
			return sensor.StateUnknown
		}
	case hwAddrPropName:
		if value, ok := value.(string); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	case freqPropName, maxBitRatePropName, bandwidthPropName:
		if value, ok := value.(uint32); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	case strPropName:
		if value, ok := value.(uint8); ok {
			return value
		} else {
			return sensor.StateUnknown
		}
	default:
		return sensor.StateUnknown
	}
}

func generateStrIcon(value any) string {
	str, ok := value.(uint8)

	switch {
	case !ok:
		return "mdi:wifi-strength-alert-outline"
	case str <= 25:
		return "mdi:wifi-strength-1"
	case str > 25 && str <= 50:
		return "mdi:wifi-strength-2"
	case str > 50 && str <= 75:
		return "mdi:wifi-strength-3"
	case str > 75:
		return "mdi:wifi-strength-4"
	default:
		return "mdi:wifi-strength-alert-outline"
	}
}
