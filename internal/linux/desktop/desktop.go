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

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
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

type settingsWorker struct {
	triggerCh chan dbusx.Trigger
	getProp   func(prop string) (dbus.Variant, error)
	prefs     *WorkerPrefs
}

func (w *settingsWorker) PreferencesID() string {
	return preferencesID
}

func (w *settingsWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{}
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
					scheme, icon := parseColorScheme(value)

					sensorCh <- sensor.NewSensor(
						sensor.WithName("Desktop Color Scheme"),
						sensor.WithID("desktop_color_scheme"),
						sensor.AsDiagnostic(),
						sensor.WithState(
							sensor.WithIcon(icon),
							sensor.WithValue(scheme),
							sensor.WithDataSourceAttribute(linux.DataSrcDbus),
						),
					)
				case accentColorProp:
					sensorCh <- sensor.NewSensor(
						sensor.WithName("Desktop Accent Color"),
						sensor.WithID("desktop_accent_color"),
						sensor.AsDiagnostic(),
						sensor.WithState(
							sensor.WithIcon("mdi:palette"),
							sensor.WithValue(parseAccentColor(value)),
							sensor.WithDataSourceAttribute(linux.DataSrcDbus),
						),
					)
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
func (w *settingsWorker) Sensors(ctx context.Context) ([]sensor.Entity, error) {
	sensors := make([]sensor.Entity, 0, 2)

	var errs error

	if value, err := w.getProp(accentColorProp); err != nil {
		logging.FromContext(ctx).Warn("Could not retrieve accent color property", slog.Any("error", err))
	} else {
		sensors = append(sensors, sensor.NewSensor(
			sensor.WithName("Desktop Accent Color"),
			sensor.WithID("desktop_accent_color"),
			sensor.AsDiagnostic(),
			sensor.WithState(
				sensor.WithIcon("mdi:palette"),
				sensor.WithValue(parseAccentColor(value)),
				sensor.WithDataSourceAttribute(linux.DataSrcDbus),
			),
		))
	}

	if value, err := w.getProp(colorSchemeProp); err != nil {
		logging.FromContext(ctx).Warn("Could not retrieve color scheme property", slog.Any("error", err))
	} else {
		scheme, icon := parseColorScheme(value)

		sensors = append(sensors, sensor.NewSensor(
			sensor.WithName("Desktop Color Scheme"),
			sensor.WithID("desktop_color_scheme"),
			sensor.AsDiagnostic(),
			sensor.WithState(
				sensor.WithIcon(icon),
				sensor.WithValue(scheme),
				sensor.WithDataSourceAttribute(linux.DataSrcDbus),
			),
		))
	}

	return sensors, errs
}

func NewDesktopWorker(ctx context.Context) (*linux.EventSensorWorker, error) {
	var err error

	worker := linux.NewEventSensorWorker(workerID)

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

	settingsWorker := &settingsWorker{
		triggerCh: triggerCh,
		getProp: func(prop string) (dbus.Variant, error) {
			var value dbus.Variant
			value, err = dbusx.GetData[dbus.Variant](bus,
				desktopPortalPath,
				desktopPortalInterface,
				settingsPortalInterface+".Read",
				"org.freedesktop.appearance",
				prop)
			if err != nil {
				return dbus.Variant{}, fmt.Errorf("could not retrieve desktop property %s from D-Bus: %w", prop, err)
			}

			return value, nil
		},
	}

	settingsWorker.prefs, err = preferences.LoadWorker(ctx, settingsWorker)
	if err != nil {
		return worker, fmt.Errorf("could not load preferences: %w", err)
	}

	// If disabled, don't use the addressWorker.
	if settingsWorker.prefs.Disabled {
		slog.Info("disabled")
		return worker, nil
	}

	return worker, nil
}

//nolint:mnd
func parseColorScheme(value dbus.Variant) (string, string) {
	scheme, err := dbusx.VariantToValue[uint32](value)
	if err != nil {
		return sensor.StateUnknown, "mdi:theme-light-dark"
	}

	switch scheme {
	case 1:
		return "dark", "mdi:weather-night"
	case 2:
		return "light", "mdi:weather-sunny"
	default:
		return sensor.StateUnknown, "mdi:theme-light-dark"
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
