// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
//nolint:misspell
package desktop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/mandykoh/prism/srgb"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
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

type desktopSettingSensor struct {
	linux.Sensor
}

type worker struct{}

//nolint:exhaustruct
func (w *worker) Setup(_ context.Context) *dbusx.Watch {
	return &dbusx.Watch{
		Bus:       dbusx.SessionBus,
		Names:     []string{settingsChangedSignal},
		Interface: settingsPortalInterface,
		Path:      desktopPortalPath,
	}
}

func (w *worker) Watch(ctx context.Context, triggerCh chan dbusx.Trigger) chan sensor.Details {
	sensorCh := make(chan sensor.Details)

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopped desktop settings sensors.")

				return
			case event := <-triggerCh:
				if !strings.Contains(event.Signal, settingsChangedSignal) {
					continue
				}

				prop, value := extractProp(event.Content)

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
			log.Warn().Err(err).Msg("Could not get initial sensor updates.")
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}()

	return sensorCh
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

func NewDesktopWorker(_ context.Context) (*linux.SensorWorker, error) {
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
	scheme := dbusx.VariantToValue[uint32](value)
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
	values := dbusx.VariantToValue[[]any](value)
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
	if accent == "" {
		accent = getProp(ctx, accentColorProp)
	}

	newSensor := &desktopSettingSensor{}
	newSensor.IsDiagnostic = true
	newSensor.IconString = "mdi:palette"
	newSensor.SensorSrc = linux.DataSrcDbus
	newSensor.SensorTypeValue = linux.SensorAccentColor
	newSensor.Value = accent

	return newSensor
}

//nolint:exhaustruct
func newColorSchemeSensor(ctx context.Context, scheme string) *desktopSettingSensor {
	if scheme == "" {
		scheme = getProp(ctx, colorSchemeProp)
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

func getProp(ctx context.Context, prop string) string {
	value, err := dbusx.GetData[dbus.Variant](ctx,
		dbusx.SessionBus,
		desktopPortalPath,
		desktopPortalInterface,
		settingsPortalInterface+".Read",
		"org.freedesktop.appearance",
		prop)
	if err != nil {
		log.Warn().Err(err).Msg("Could not retrieve accent color from D-Bus.")

		return sensor.StateUnknown
	}

	switch prop {
	case accentColorProp:
		return parseAccentColor(value)
	case colorSchemeProp:
		return parseColorScheme(value)
	}

	return sensor.StateUnknown
}

func extractProp(event []any) (prop string, value dbus.Variant) {
	var ok bool

	prop, ok = event[1].(string)
	if !ok {
		log.Warn().Msg("Didn't understand changed property.")
	}

	value, ok = event[2].(dbus.Variant)
	if !ok {
		log.Warn().Msg("Didn't understand changed property value.")
	}

	return prop, value
}
