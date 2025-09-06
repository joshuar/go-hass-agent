// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"context"
	"errors"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/class"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
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

	unknownState = "Unknown"
)

var apPropList = []string{ssidPropName, hwAddrPropName, maxBitRatePropName, freqPropName, strPropName, bandwidthPropName}

var ErrNewWifiPropSensor = errors.New("could not create wifi property sensor")

func newWifiSensor(ctx context.Context, prop string, value any) models.Entity {
	var (
		name, id, units string
		deviceClass     class.SensorDeviceClass
		stateClass      class.SensorStateClass
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
		deviceClass = class.SensorClassDataRate
		stateClass = class.StateMeasurement
	case freqPropName:
		name = "Wi-Fi Frequency"
		id = "wi_fi_frequency"
		units = "MHz"
		deviceClass = class.SensorClassFrequency
		stateClass = class.StateMeasurement
	case bandwidthPropName:
		name = "Wi-Fi Bandwidth"
		id = "wi_fi_bandwidth"
		units = "MHz"
		deviceClass = class.SensorClassFrequency
		stateClass = class.StateMeasurement
	case strPropName:
		name = "Wi-Fi Signal Strength"
		id = "wi_fi_signal_strength"
		units = "%"
		stateClass = class.StateMeasurement
		icon = generateStrIcon(value)
	}

	return sensor.NewSensor(ctx,
		sensor.WithName(name),
		sensor.WithID(id),
		sensor.AsDiagnostic(),
		sensor.WithIcon(icon),
		sensor.WithState(generateState(prop, value)),
		sensor.WithDeviceClass(deviceClass),
		sensor.WithStateClass(stateClass),
		sensor.WithUnits(units),
	)
}

func getWifiSensors(ctx context.Context, bus *dbusx.Bus, apPath string) []models.Entity {
	sensors := make([]models.Entity, 0, len(apPropList))

	for _, prop := range apPropList {
		value, err := dbusx.NewProperty[any](bus, apPath, dBusNMObj, accessPointInterface+"."+prop).Get()
		if err != nil {
			slog.Debug("Could not retrieve access point property.",
				slog.String("prop", prop),
				slog.Any("error", err))

			continue
		}
		sensors = append(sensors, newWifiSensor(ctx, prop, value))
	}

	return sensors
}

func generateState(prop string, value any) any {
	switch prop {
	case ssidPropName:
		if value, ok := value.([]uint8); ok {
			return string(value)
		}

		return unknownState
	case hwAddrPropName:
		if value, ok := value.(string); ok {
			return value
		}

		return unknownState
	case freqPropName, maxBitRatePropName, bandwidthPropName:
		if value, ok := value.(uint32); ok {
			return value
		}

		return unknownState
	case strPropName:
		if value, ok := value.(uint8); ok {
			return value
		}

		return unknownState
	default:
		return unknownState
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
