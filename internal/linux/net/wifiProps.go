// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package net

import (
	"context"
	"slices"

	"github.com/godbus/dbus/v5"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	accessPointInterface = dBusNMObj + ".AccessPoint"
	accessPointProp      = dBusNMObj + ".Device.Wireless.ActiveAccessPoint"

	ssidProp    = "Ssid"
	hwAddrProp  = "HwAddress"
	bitRateProp = "MaxBitrate"
	freqProp    = "Frequency"
	strProp     = "Strength"
)

var wifiPropList = []string{ssidProp, hwAddrProp, bitRateProp, freqProp, strProp}

type wifiSensor struct {
	linux.Sensor
}

//nolint:exhaustive
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

//nolint:exhaustive,mnd
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

func newWifiSensor(sensorType string) *wifiSensor {
	switch sensorType {
	case ssidProp:
		return &wifiSensor{
			Sensor: linux.Sensor{
				SensorTypeValue: linux.SensorWifiSSID,
				IsDiagnostic:    true,
			},
		}
	case hwAddrProp:
		return &wifiSensor{
			Sensor: linux.Sensor{
				SensorTypeValue: linux.SensorWifiHWAddress,
				IsDiagnostic:    true,
			},
		}
	case bitRateProp:
		return &wifiSensor{
			Sensor: linux.Sensor{
				SensorTypeValue:  linux.SensorWifiSpeed,
				UnitsString:      "kB/s",
				DeviceClassValue: types.DeviceClassDataRate,
				StateClassValue:  types.StateClassMeasurement,
				IsDiagnostic:     true,
			},
		}
	case freqProp:
		return &wifiSensor{
			Sensor: linux.Sensor{
				SensorTypeValue:  linux.SensorWifiFrequency,
				UnitsString:      "MHz",
				DeviceClassValue: types.DeviceClassFrequency,
				StateClassValue:  types.StateClassMeasurement,
				IsDiagnostic:     true,
			},
		}
	case strProp:
		return &wifiSensor{
			Sensor: linux.Sensor{
				SensorTypeValue: linux.SensorWifiStrength,
				UnitsString:     "%",
				StateClassValue: types.StateClassMeasurement,
				IsDiagnostic:    true,
			},
		}
	}

	return nil
}

// monitorWifi will initially fetch and then monitor for changes of
// relevant WiFi properties that are to be represented as sensors.
func (c *connection) monitorWifi(ctx context.Context) <-chan sensor.Details {
	outCh := make(chan sensor.Details)
	// get the devices associated with this connection
	wifiDevices, err := dbusx.GetProp[[]dbus.ObjectPath](ctx, c.bus, string(c.path), dBusNMObj, dbusNMActiveConnIntr+".Devices")
	if err != nil {
		c.logger.Warn("Could not retrieve active wireless device from D-Bus.", "error", err.Error())

		return nil
	}

	for _, d := range wifiDevices {
		// for each device, get the access point it is currently associated with
		accessPointPath, err := dbusx.GetProp[dbus.ObjectPath](ctx, c.bus, string(d), dBusNMObj, accessPointProp)
		if err != nil {
			c.logger.Warn("Could not retrieve access point object from D-Bus.", "error", err.Error())

			continue
		}

		for _, prop := range wifiPropList {
			// for the associated access point, get the wifi properties as sensors
			value, err := dbusx.GetProp[any](ctx, c.bus, string(accessPointPath), dBusNMObj, accessPointInterface+"."+prop)
			if err != nil {
				c.logger.Warn("Could not retrieve wifi property from D-Bus.", "property", prop, "error", err.Error())

				continue
			}

			wifiSensor := newWifiSensor(prop)
			if wifiSensor == nil {
				c.logger.Warn("Unhandled wifi property.", "prop", prop)

				continue
			}

			wifiSensor.Value = value

			// send the wifi property as a sensor
			go func() {
				outCh <- wifiSensor
			}()
		}
		// monitor for changes in the wifi properties for this device
		go func() {
			for s := range c.monitorWifiProps(ctx, accessPointPath) {
				outCh <- s
			}
		}()
	}

	return outCh
}

func (c *connection) monitorWifiProps(ctx context.Context, propPath dbus.ObjectPath) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	events, err := c.bus.WatchBus(ctx, &dbusx.Watch{
		Names: wifiPropList,
		Path:  string(propPath),
	})
	if err != nil {
		c.logger.Debug("Failed to watch D-Bus for wifi property changes.", "error", err.Error())
		close(sensorCh)

		return sensorCh
	}

	go func() {
		c.logger.Debug("Monitoring D-Bus for wifi property changes.")

		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				c.logger.Debug("Unmonitoring D-Bus for wifi property changes.")

				return
			case event := <-events:
				props, err := dbusx.ParsePropertiesChanged(event.Content)
				if err != nil {
					c.logger.Warn("Received an unknown event from D-Bus.", "error", err.Error())

					continue
				}

				for prop, value := range props.Changed {
					if slices.Contains(wifiPropList, prop) {
						wifiSensor := newWifiSensor(prop)
						wifiSensor.Value = value.Value()
						sensorCh <- wifiSensor
					}
				}
			}
		}
	}()

	return sensorCh
}
