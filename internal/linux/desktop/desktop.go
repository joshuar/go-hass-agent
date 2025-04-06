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
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/models"
	"github.com/joshuar/go-hass-agent/internal/models/sensor"
	"github.com/joshuar/go-hass-agent/internal/workers"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

var _ workers.EntityWorker = (*settingsWorker)(nil)

var (
	ErrInitDesktopWorker = errors.New("could not init desktop worker")
	ErrUnknownProp       = errors.New("unknown desktop property")
)

const (
	portalInterface         = "org.freedesktop.portal"
	desktopPortalPath       = "/org/freedesktop/portal/desktop"
	desktopPortalInterface  = portalInterface + ".Desktop"
	settingsPortalInterface = portalInterface + ".Settings"
	settingsChangedSignal   = "SettingChanged"
	colorSchemeProp         = "color-scheme"
	accentColorProp         = "accent-color"

	desktopWorkerID     = "desktop_settings_sensors"
	desktopWorkerDesc   = "Desktop settings"
	desktopWorkerPrefID = prefPrefix + "preferences"

	unknownValue = "Unknown"
)

type settingsWorker struct {
	bus   *dbusx.Bus
	prefs *WorkerPrefs
	*models.WorkerMetadata
}

func (w *settingsWorker) PreferencesID() string {
	return desktopWorkerPrefID
}

func (w *settingsWorker) DefaultPreferences() WorkerPrefs {
	return WorkerPrefs{}
}

func (w *settingsWorker) IsDisabled() bool {
	return w.prefs.IsDisabled()
}

//nolint:cyclop,gocognit
func (w *settingsWorker) Start(ctx context.Context) (<-chan models.Entity, error) {
	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(desktopPortalPath),
		dbusx.MatchInterface(settingsPortalInterface),
		dbusx.MatchMembers(settingsChangedSignal),
	).Start(ctx, w.bus)
	if err != nil {
		return nil, errors.Join(ErrInitDesktopWorker,
			fmt.Errorf("could not watch D-Bus for desktop settings updates: %w", err))
	}
	sensorCh := make(chan models.Entity)
	logger := logging.FromContext(ctx).With(slog.String("worker", desktopWorkerID))

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
					logger.Debug("Error processing received signal.", slog.Any("error", err))
				}

				switch prop {
				case colorSchemeProp:
					scheme, icon := parseColorScheme(value)

					entity, err := newColorSchemeSensor(ctx, scheme, icon)
					if err != nil {
						logger.Warn("Could not generate color scheme sensor.", slog.Any("error", err))
					} else {
						sensorCh <- entity
					}
				case accentColorProp:
					entity, err := newAccentColorSensor(ctx, parseAccentColor(value))
					if err != nil {
						logger.Warn("Could not generate accent color sensor.", slog.Any("error", err))
					} else {
						sensorCh <- entity
					}
				}
			}
		}
	}()
	// Send an initial update.
	go func() {
		sensors, err := w.generateSensors(ctx)
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
func (w *settingsWorker) generateSensors(ctx context.Context) ([]models.Entity, error) {
	sensors := make([]models.Entity, 0, 2)

	var errs error

	// Accent Color Sensor.
	if value, err := w.getProp(accentColorProp); err != nil {
		logging.FromContext(ctx).Warn("Could not retrieve accent color property", slog.Any("error", err))
	} else {
		entity, err := newAccentColorSensor(ctx, parseAccentColor(value))
		if err != nil {
			logging.FromContext(ctx).Warn("Could not generate accent color sensor.", slog.Any("error", err))
		} else {
			sensors = append(sensors, entity)
		}
	}

	// Color Theme Sensor.
	if value, err := w.getProp(colorSchemeProp); err != nil {
		logging.FromContext(ctx).Warn("Could not retrieve color scheme property", slog.Any("error", err))
	} else {
		scheme, icon := parseColorScheme(value)

		entity, err := newColorSchemeSensor(ctx, scheme, icon)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not generate color scheme sensor.", slog.Any("error", err))
		} else {
			sensors = append(sensors, entity)
		}
	}

	return sensors, errs
}

func (w *settingsWorker) getProp(prop string) (dbus.Variant, error) {
	value, err := dbusx.GetData[dbus.Variant](w.bus,
		desktopPortalPath,
		desktopPortalInterface,
		settingsPortalInterface+".Read",
		"org.freedesktop.appearance",
		prop)
	if err != nil {
		return dbus.Variant{}, errors.Join(ErrInitDesktopWorker,
			fmt.Errorf("could not retrieve desktop property %s from D-Bus: %w", prop, err))
	}

	return value, nil
}

func NewDesktopWorker(ctx context.Context) (workers.EntityWorker, error) {
	_, ok := linux.CtxGetDesktopPortal(ctx)
	if !ok {
		return nil, errors.Join(ErrInitDesktopWorker, linux.ErrNoDesktopPortal)
	}

	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitDesktopWorker, linux.ErrNoSessionBus)
	}

	worker := &settingsWorker{
		WorkerMetadata: models.SetWorkerMetadata(desktopWorkerID, desktopWorkerDesc),
		bus:            bus,
	}

	prefs, err := preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitDesktopWorker, err)
	}
	worker.prefs = prefs

	return worker, nil
}

//nolint:mnd
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

//nolint:mnd
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

func newColorSchemeSensor(ctx context.Context, scheme, icon string) (models.Entity, error) {
	entity, err := sensor.NewSensor(ctx,
		sensor.WithName("Desktop Color Scheme"),
		sensor.WithID("desktop_color_scheme"),
		sensor.AsDiagnostic(),
		sensor.WithIcon(icon),
		sensor.WithState(scheme),
		sensor.WithDataSourceAttribute(linux.DataSrcDbus),
	)
	if err != nil {
		return entity, fmt.Errorf("could not create color scheme sensor: %w", err)
	}

	return entity, nil
}

func newAccentColorSensor(ctx context.Context, value string) (models.Entity, error) {
	entity, err := sensor.NewSensor(ctx,
		sensor.WithName("Desktop Accent Color"),
		sensor.WithID("desktop_accent_color"),
		sensor.AsDiagnostic(),
		sensor.WithIcon("mdi:palette"),
		sensor.WithState(value),
		sensor.WithDataSourceAttribute(linux.DataSrcDbus),
	)
	if err != nil {
		return entity, fmt.Errorf("could not create color scheme sensor: %w", err)
	}

	return entity, nil
}
