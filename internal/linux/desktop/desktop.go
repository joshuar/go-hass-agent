// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
//nolint:exhaustruct,misspell
package desktop

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/mandykoh/prism/srgb"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	portalInterface         = "org.freedesktop.portal"
	desktopPortalPath       = "/org/freedesktop/portal/desktop"
	desktopPortalInterface  = portalInterface + ".Desktop"
	settingsPortalInterface = portalInterface + ".Settings"
	settingsChangedSignal   = "SettingChanged"
	colorSchemeProp         = "color-scheme"
	accentColorProp         = "accent-color"

	reqTimeout = 15 * time.Second
)

var ErrUnknownProp = errors.New("unknown desktop property")

type desktopSettingSensor struct {
	linux.Sensor
}

type worker struct{}

//nolint:cyclop
func (w *worker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)

	triggerCh, err := dbusx.WatchBus(ctx, &dbusx.Watch{
		Bus:       dbusx.SessionBus,
		Names:     []string{settingsChangedSignal},
		Interface: settingsPortalInterface,
		Path:      desktopPortalPath,
	})
	if err != nil {
		close(sensorCh)

		return sensorCh, fmt.Errorf("could not watch D-Bus for desktop settings updates: %w", err)
	}

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-triggerCh:
				if !strings.Contains(event.Signal, settingsChangedSignal) {
					continue
				}

				prop, value, err := extractProp(event.Content)
				if err != nil {
					logging.FromContext(ctx).Warn("Error processing received signal.", "error", err.Error())
				}

				switch prop {
				case colorSchemeProp:
					s := parseColorScheme(value)
					sensorCh <- newColorSchemeSensor(ctx, s)
				case accentColorProp:
					s := parseAccentColor(value)
					sensorCh <- newAccentColorSensor(ctx, s)
				}
			}
		}
	}()
	// Send an initial update.
	go func() {
		sensors, err := w.Sensors(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not get desktop settings from D-Bus.", "error", err.Error())
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}()

	return sensorCh, nil
}

//nolint:mnd
func (w *worker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	reqCtx, cancelReq := context.WithTimeout(ctx, reqTimeout)
	defer cancelReq()

	sensors := make([]sensor.Details, 0, 2)

	sensors = append(sensors,
		newAccentColorSensor(reqCtx, ""),
		newColorSchemeSensor(reqCtx, ""))

	return sensors, nil
}

func NewDesktopWorker() (*linux.SensorWorker, error) {
	// If we cannot find a portal interface, we cannot monitor desktop settings.
	_, err := linux.FindPortal()
	if err != nil {
		return nil, fmt.Errorf("unable to monitor for desktop settings: %w", err)
	}

	return &linux.SensorWorker{
			WorkerName: "Desktop Preferences Sensors",
			WorkerDesc: "The desktop theme type (light/dark) and accent color.",
			Value:      &worker{},
		},
		nil
}

//nolint:mnd
func parseColorScheme(value dbus.Variant) string {
	scheme, err := dbusx.VariantToValue[uint32](value)
	if err != nil {
		return sensor.StateUnknown
	}

	switch scheme {
	case 1:
		return "dark"
	case 2:
		return "light"
	default:
		return sensor.StateUnknown
	}
}

//nolint:mnd
func parseAccentColor(value dbus.Variant) string {
	values, err := dbusx.VariantToValue[[]any](value)
	if err != nil {
		return sensor.StateUnknown
	}

	rgb := make([]uint8, 3)

	for colour, v := range values {
		val, ok := v.(float64)
		if !ok {
			continue
		}

		rgb[colour] = srgb.To8Bit(float32(val))
	}

	return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])
}

//nolint:exhaustruct
func newAccentColorSensor(ctx context.Context, accent string) *desktopSettingSensor {
	var err error

	if accent == "" {
		accent, err = getProp(ctx, accentColorProp)
		if err != nil {
			logging.FromContext(ctx).Warn("Invalid accent colour.", "error", err.Error())
		}
	}

	newSensor := &desktopSettingSensor{}
	newSensor.IsDiagnostic = true
	newSensor.IconString = "mdi:palette"
	newSensor.SensorSrc = linux.DataSrcDbus
	newSensor.SensorTypeValue = linux.SensorAccentColor
	newSensor.Value = accent

	return newSensor
}

func newColorSchemeSensor(ctx context.Context, scheme string) *desktopSettingSensor {
	var err error

	if scheme == "" {
		scheme, err = getProp(ctx, colorSchemeProp)
		if err != nil {
			logging.FromContext(ctx).Warn("Invalid colour scheme.", "error", err.Error())
		}
	}

	newSensor := &desktopSettingSensor{}
	newSensor.IsDiagnostic = true
	newSensor.SensorSrc = linux.DataSrcDbus
	newSensor.SensorTypeValue = linux.SensorColorScheme
	newSensor.Value = scheme

	switch scheme {
	case "dark":
		newSensor.IconString = "mdi:weather-night"
	case "light":
		newSensor.IconString = "mdi:weather-sunny"
	default:
		newSensor.IconString = "mdi:theme-light-dark"
	}

	return newSensor
}

func getProp(ctx context.Context, prop string) (string, error) {
	value, err := dbusx.GetData[dbus.Variant](ctx,
		dbusx.SessionBus,
		desktopPortalPath,
		desktopPortalInterface,
		settingsPortalInterface+".Read",
		"org.freedesktop.appearance",
		prop)
	if err != nil {
		return sensor.StateUnknown, fmt.Errorf("could not retrieve desktop property %s from D-Bus: %w", prop, err)
	}

	switch prop {
	case accentColorProp:
		return parseAccentColor(value), nil
	case colorSchemeProp:
		return parseColorScheme(value), nil
	}

	return sensor.StateUnknown, fmt.Errorf("could not retrieve desktop property %s from D-Bus: %w", prop, ErrUnknownProp)
}

func extractProp(event []any) (prop string, value dbus.Variant, err error) {
	var ok bool

	prop, ok = event[1].(string)
	if !ok {
		return "", dbus.Variant{}, fmt.Errorf("error extracting property from D-Bus signal: %w", ErrUnknownProp)
	}

	value, ok = event[2].(dbus.Variant)
	if !ok {
		return "", dbus.Variant{}, fmt.Errorf("error extracting property from D-Bus signal: %w", ErrUnknownProp)
	}

	return prop, value, nil
}
