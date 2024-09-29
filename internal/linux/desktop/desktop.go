// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package desktop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/mandykoh/prism/srgb"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor/types"
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

	workerID = "desktop_settings_sensors"
)

var ErrUnknownProp = errors.New("unknown desktop property")

type settingsWorker struct {
	triggerCh chan dbusx.Trigger
	getProp   func(prop string) (string, error)
}

func (w *settingsWorker) newAccentColorSensor(accent string) (sensor.Entity, error) {
	var err error

	if accent == "" {
		accent, err = w.getProp(accentColorProp)
		if err != nil {
			return sensor.Entity{}, fmt.Errorf("invalid accent color: %w", err)
		}
	}

	return sensor.Entity{
			Category: types.CategoryDiagnostic,
			Name:     "Desktop Accent Color",
			State: &sensor.State{
				ID:    "desktop_accent_color",
				Value: accent,
				Icon:  "mdi:palette",
				Attributes: map[string]any{
					"data_source": linux.DataSrcDbus,
				},
			},
		},
		nil
}

func (w *settingsWorker) newColorSchemeSensor(scheme string) (sensor.Entity, error) {
	var err error

	if scheme == "" {
		scheme, err = w.getProp(colorSchemeProp)
		if err != nil {
			return sensor.Entity{}, fmt.Errorf("invalid color scheme: %w", err)
		}
	}

	newSensor := sensor.Entity{
		Category: types.CategoryDiagnostic,
		Name:     "Desktop Color Scheme",
		State: &sensor.State{
			ID:    "desktop_color_scheme",
			Value: scheme,
			Attributes: map[string]any{
				"data_source": linux.DataSrcDbus,
			},
		},
	}

	switch scheme {
	case "dark":
		newSensor.Icon = "mdi:weather-night"
	case "light":
		newSensor.Icon = "mdi:weather-sunny"
	default:
		newSensor.Icon = "mdi:theme-light-dark"
	}

	return newSensor, nil
}

//nolint:cyclop,gocognit
func (w *settingsWorker) Events(ctx context.Context) (<-chan sensor.Entity, error) {
	sensorCh := make(chan sensor.Entity)
	logger := logging.FromContext(ctx).With(slog.String("worker", workerID))

	go func() {
		defer close(sensorCh)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-w.triggerCh:
				if !strings.Contains(event.Signal, settingsChangedSignal) {
					continue
				}

				prop, value, err := extractProp(event.Content)
				if err != nil {
					logger.Debug("Error processing received signal.", slog.Any("error", err))
				}

				switch prop {
				case colorSchemeProp:
					if colourSchemeSensor, err := w.newColorSchemeSensor(parseColorScheme(value)); err != nil {
						logger.Debug("Error generating color scheme sensor.", slog.Any("error", err))
					} else {
						sensorCh <- colourSchemeSensor
					}
				case accentColorProp:
					if accentColourSensor, err := w.newAccentColorSensor(parseAccentColor(value)); err != nil {
						logger.Debug("Error generating accent color sensor.", slog.Any("error", err))
					} else {
						sensorCh <- accentColourSensor
					}
				}
			}
		}
	}()
	// Send an initial update.
	go func() {
		sensors, err := w.Sensors(ctx)
		if err != nil {
			logger.Debug("Could not get desktop settings from D-Bus.", slog.Any("error", err))
		}

		for _, s := range sensors {
			sensorCh <- s
		}
	}()

	return sensorCh, nil
}

//nolint:mnd
func (w *settingsWorker) Sensors(_ context.Context) ([]sensor.Entity, error) {
	sensors := make([]sensor.Entity, 0, 2)

	var errs error

	if colourSchemeSensor, err := w.newColorSchemeSensor(""); err != nil {
		errs = errors.Join(errs, err)
	} else {
		sensors = append(sensors, colourSchemeSensor)
	}

	if accentColourSensor, err := w.newAccentColorSensor(""); err != nil {
		errs = errors.Join(errs, err)
	} else {
		sensors = append(sensors, accentColourSensor)
	}

	return sensors, errs
}

func NewDesktopWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	worker := linux.NewEventWorker(workerID)

	_, ok := linux.CtxGetDesktopPortal(ctx)
	if !ok {
		return worker, linux.ErrNoDesktopPortal
	}

	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return worker, linux.ErrNoSessionBus
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(desktopPortalPath),
		dbusx.MatchInterface(settingsPortalInterface),
		dbusx.MatchMembers(settingsChangedSignal),
	).Start(ctx, bus)
	if err != nil {
		return worker, fmt.Errorf("could not watch D-Bus for desktop settings updates: %w", err)
	}

	worker.EventType = &settingsWorker{
		triggerCh: triggerCh,
		getProp: func(prop string) (string, error) {
			value, err := dbusx.GetData[dbus.Variant](bus,
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
		},
	}

	return worker, nil
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

	for color, v := range values {
		val, ok := v.(float64)
		if !ok {
			continue
		}

		rgb[color] = srgb.To8Bit(float32(val))
	}

	return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])
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
