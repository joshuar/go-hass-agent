// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package desktop

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/mandykoh/prism/srgb"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/models/sensor"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

var _ workers.EntityWorker = (*settingsWorker)(nil)

const (
	portalInterface         = "org.freedesktop.portal"
	desktopPortalPath       = "/org/freedesktop/portal/desktop"
	desktopPortalInterface  = portalInterface + ".Desktop"
	settingsPortalInterface = portalInterface + ".Settings"
	settingsChangedSignal   = "SettingChanged"
	colorSchemeProp         = "color-scheme"
	accentColorProp         = "accent-color"

	unknownValue = "Unknown"
)

type settingsWorker struct {
	*models.WorkerMetadata

	bus   *dbusx.Bus
	prefs *WorkerPrefs
}

func (w *settingsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

func (w *settingsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(desktopPortalPath),
		dbusx.MatchInterface(settingsPortalInterface),
		dbusx.MatchMembers(settingsChangedSignal),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, fmt.Errorf("watch desktop settings: %w", err)
	}
	sensorCh := make(chan models.Entity)

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

				prop, value := extractProp(event.Content)

				switch prop {
				case colorSchemeProp:
					scheme, icon := parseColorScheme(value)
					sensorCh <- newColorSchemeSensor(ctx, scheme, icon)
				case accentColorProp:
					sensorCh <- newAccentColorSensor(ctx, parseAccentColor(value))
				}
			}
		}
	}()
	// Send an initial update.
	go func() {
		for _, s := range w.generateSensors(ctx) {
			sensorCh <- s
		}
	}()

	return sensorCh, nil
}

func (w *settingsWorker) generateSensors(ctx context.Context) []models.Entity {
	sensors := make([]models.Entity, 0, 2)

	// Accent Color Sensor.
	if value, err := w.getProp(accentColorProp); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not retrieve accent color property", slog.Any("error", err))
	} else {
		sensors = append(sensors, newAccentColorSensor(ctx, parseAccentColor(value)))
	}

	// Color Theme Sensor.
	if value, err := w.getProp(colorSchemeProp); err != nil {
		slogctx.FromCtx(ctx).Warn("Could not retrieve color scheme property", slog.Any("error", err))
	} else {
		scheme, icon := parseColorScheme(value)
		sensors = append(sensors, newColorSchemeSensor(ctx, scheme, icon))
	}

	return sensors
}

func (w *settingsWorker) getProp(prop string) (dbus.Variant, error) {
	value, err := dbusx.GetData[dbus.Variant](w.bus,
		desktopPortalPath,
		desktopPortalInterface,
		settingsPortalInterface+".Read",
		"org.freedesktop.appearance",
		prop)
	if err != nil {
		return dbus.Variant{}, fmt.Errorf("could not retrieve desktop property %s from D-Bus: %w", prop, err)
	}

	return value, nil
}

// NewDesktopWorker creates a worker to track desktop settings changes.
func NewDesktopWorker(ctx context.Context) (workers.EntityWorker, error) {
	worker := &settingsWorker{
		WorkerMetadata: models.SetWorkerMetadata("desktop_settings", "Desktop settings"),
	}

	var ok bool

	_, ok = linux.CtxGetDesktopPortal(ctx)
	if !ok {
		return worker, fmt.Errorf("get desktop portal: %w", linux.ErrNoDesktopPortal)
	}

	worker.bus, ok = linux.CtxGetSessionBus(ctx)
	if !ok {
		return worker, fmt.Errorf("get session bus: %w", linux.ErrNoSessionBus)
	}

	defaultPrefs := &WorkerPrefs{}
	var err error
	worker.prefs, err = workers.LoadWorkerPreferences(prefPrefix+"desktop_settings_sensors", defaultPrefs)
	if err != nil {
		return worker, fmt.Errorf("load preferences: %w", err)
	}

	return worker, nil
}

func parseColorScheme(value dbus.Variant) (string, string) {
	scheme, err := dbusx.VariantToValue[uint32](value)
	if err != nil {
		return unknownValue, "mdi:theme-light-dark"
	}

	switch scheme {
	case 1:
		return "dark", "mdi:weather-night"
	case 2:
		return "light", "mdi:weather-sunny"
	default:
		return unknownValue, "mdi:theme-light-dark"
	}
}

func parseAccentColor(value dbus.Variant) string {
	values, err := dbusx.VariantToValue[[]any](value)
	if err != nil {
		return "Unknown"
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

func extractProp(event []any) (string, dbus.Variant) {
	var ok bool

	prop, ok := event[1].(string)
	if !ok {
		return "", dbus.Variant{}
	}

	value, ok := event[2].(dbus.Variant)
	if !ok {
		return "", dbus.Variant{}
	}

	return prop, value
}

func newColorSchemeSensor(ctx context.Context, scheme, icon string) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName("Desktop Color Scheme"),
		sensor.WithID("desktop_color_scheme"),
		sensor.AsDiagnostic(),
		sensor.WithIcon(icon),
		sensor.WithState(scheme),
		sensor.WithDataSourceAttribute(linux.DataSrcDBus),
	)
}

func newAccentColorSensor(ctx context.Context, value string) models.Entity {
	return sensor.NewSensor(ctx,
		sensor.WithName("Desktop Accent Color"),
		sensor.WithID("desktop_accent_color"),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:palette"),
		sensor.WithState(value),
		sensor.WithDataSourceAttribute(linux.DataSrcDBus),
	)
}
