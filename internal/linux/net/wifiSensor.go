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
	var (
		name, id, units string
		deviceClass     types.DeviceClass
		stateClass      types.StateClass
	)

	icon := "mdi:wifi"

	switch prop {
	case ssidPropName:
		name = "Wi-Fi SSID"
		id = "wi_fi_ssid"
	case hwAddrPropName:
		name = "Wi-Fi BSSID"
		id = "wi_fi_bssid"
	case maxBitRatePropName:
		name = "Wi-Fi Link Speed"
		id = "wi_fi_link_speed"
		units = "kB/s"
		deviceClass = types.SensorDeviceClassDataRate
		stateClass = types.StateClassMeasurement
	case freqPropName:
		name = "Wi-Fi Frequency"
		id = "wi_fi_frequency"
		units = "MHz"
		deviceClass = types.SensorDeviceClassFrequency
		stateClass = types.StateClassMeasurement
	case bandwidthPropName:
		name = "Wi-Fi Bandwidth"
		id = "wi_fi_bandwidth"
		units = "MHz"
		deviceClass = types.SensorDeviceClassFrequency
		stateClass = types.StateClassMeasurement
	case strPropName:
		name = "Wi-Fi Signal Strength"
		id = "wi_fi_signal_strength"
		units = "%"
		stateClass = types.StateClassMeasurement
		icon = generateStrIcon(value)
	}

	wifiSensor := sensor.NewSensor(
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.AsDiagnostic(),
		sensor.WithState(
			sensor.WithIcon(icon),
			sensor.WithValue(generateState(prop, value)),
		),
	)

	if deviceClass != types.SensorDeviceClassNone {
		wifiSensor = sensor.WithDeviceClass(deviceClass)(wifiSensor)
	}

	if stateClass != types.StateClassNone {
		wifiSensor = sensor.WithStateClass(stateClass)(wifiSensor)
	}

	if units != "" {
		wifiSensor = sensor.WithUnits(units)(wifiSensor)
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
