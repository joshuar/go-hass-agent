// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
//nolint:misspell
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

	workerID = "desktop_settings_sensors"
)

var ErrUnknownProp = errors.New("unknown desktop property")

type desktopSettingSensor struct {
	linux.Sensor
}

type worker struct {
	triggerCh chan dbusx.Trigger
	getProp   func(prop string) (string, error)
}

func (w *worker) newAccentColorSensor(accent string) (*desktopSettingSensor, error) {
	var err error

	if accent == "" {
		accent, err = w.getProp(accentColorProp)
		if err != nil {
			return nil, fmt.Errorf("invalid accent colour: %w", err)
		}
	}

	return &desktopSettingSensor{
		Sensor: linux.Sensor{
			IsDiagnostic: true,
			IconString:   "mdi:palette",
			DataSource:   linux.DataSrcDbus,
			DisplayName:  "Desktop Accent Color",
			Value:        accent,
		},
	}, nil
}

func (w *worker) newColorSchemeSensor(scheme string) (*desktopSettingSensor, error) {
	var err error

	if scheme == "" {
		scheme, err = w.getProp(colorSchemeProp)
		if err != nil {
			return nil, fmt.Errorf("invalid colour scheme: %w", err)
		}
	}

	newSensor := &desktopSettingSensor{
		Sensor: linux.Sensor{
			IsDiagnostic: true,
			DataSource:   linux.DataSrcDbus,
			DisplayName:  "Desktop Colour Scheme",
			Value:        scheme,
		},
	}

	switch scheme {
	case "dark":
		newSensor.IconString = "mdi:weather-night"
	case "light":
		newSensor.IconString = "mdi:weather-sunny"
	default:
		newSensor.IconString = "mdi:theme-light-dark"
	}

	return newSensor, nil
}

//nolint:cyclop,gocognit
func (w *worker) Events(ctx context.Context) (chan sensor.Details, error) {
	sensorCh := make(chan sensor.Details)
	logger := slog.Default().With(slog.String("worker", workerID))

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
						logger.Debug("Error generating colour scheme sensor.", slog.Any("error", err))
					} else {
						sensorCh <- colourSchemeSensor
					}
				case accentColorProp:
					if accentColourSensor, err := w.newAccentColorSensor(parseAccentColor(value)); err != nil {
						logger.Debug("Error generating accent colour sensor.", slog.Any("error", err))
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
func (w *worker) Sensors(_ context.Context) ([]sensor.Details, error) {
	sensors := make([]sensor.Details, 0, 2)

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

func NewDesktopWorker(ctx context.Context) (*linux.SensorWorker, error) {
	_, ok := linux.CtxGetDesktopPortal(ctx)
	if !ok {
		return nil, linux.ErrNoDesktopPortal
	}

	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return nil, linux.ErrNoSessionBus
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(desktopPortalPath),
		dbusx.MatchInterface(settingsPortalInterface),
		dbusx.MatchMembers(settingsChangedSignal),
	).Start(ctx, bus)
	if err != nil {
		return nil, fmt.Errorf("could not watch D-Bus for desktop settings updates: %w", err)
	}

	return &linux.SensorWorker{
			Value: &worker{
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
				triggerCh: triggerCh,
			},
			WorkerID: workerID,
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
