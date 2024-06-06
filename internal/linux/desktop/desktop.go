// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package desktop

import (
	"context"
	"errors"
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
)

type desktopSettingSensor struct {
	linux.Sensor
}

type worker struct{}

func (w *worker) Setup(ctx context.Context) *dbusx.Watch {
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
				prop, ok := event.Content[1].(string)
				if !ok {
					log.Warn().Msg("Didn't understand changed property.")
					continue
				}
				value, ok := event.Content[2].(dbus.Variant)
				if !ok {
					log.Warn().Msg("Didn't understand changed property value.")
					continue
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
			log.Warn().Err(err).Msg("Could not get initial sensor updates.")
		}
		for _, s := range sensors {
			sensorCh <- s
		}
	}()
	return sensorCh
}

func (w *worker) Sensors(ctx context.Context) ([]sensor.Details, error) {
	var sensors []sensor.Details
	reqCtx, cancelReq := context.WithTimeout(ctx, 15*time.Second)
	defer cancelReq()
	sensors = append(sensors,
		newAccentColorSensor(reqCtx, ""),
		newColorSchemeSensor(reqCtx, ""))
	return sensors, nil
}

func NewDesktopWorker() (*linux.SensorWorker, error) {
	// If we cannot find a portal interface, we cannot monitor desktop settings.
	portalDest := linux.FindPortal()
	if portalDest == "" {
		return nil, errors.New("unable to monitor for desktop settings: no portal present")
	}

	return &linux.SensorWorker{
			WorkerName: "Desktop Preferences Sensors",
			WorkerDesc: "The desktop theme type (light/dark) and accent color.",
			Value:      &worker{},
		},
		nil
}

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

func parseAccentColor(value dbus.Variant) string {
	values := dbusx.VariantToValue[[]any](value)
	rgb := make([]uint8, 3)
	for i, v := range values {
		if val, ok := v.(float64); !ok {
			continue
		} else {
			rgb[i] = srgb.To8Bit(float32(val))
		}
	}
	return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])
}

func newAccentColorSensor(ctx context.Context, accent string) *desktopSettingSensor {
	if accent == "" {
		accent = getProp(ctx, accentColorProp)
	}
	s := &desktopSettingSensor{}
	s.IsDiagnostic = true
	s.IconString = "mdi:palette"
	s.SensorSrc = linux.DataSrcDbus
	s.SensorTypeValue = linux.SensorAccentColor
	s.Value = accent
	return s
}

func newColorSchemeSensor(ctx context.Context, scheme string) *desktopSettingSensor {
	if scheme == "" {
		scheme = getProp(ctx, colorSchemeProp)
	}
	s := &desktopSettingSensor{}
	s.IsDiagnostic = true
	s.SensorSrc = linux.DataSrcDbus
	s.SensorTypeValue = linux.SensorColorScheme
	s.Value = scheme
	switch scheme {
	case "dark":
		s.IconString = "mdi:weather-night"
	case "light":
		s.IconString = "mdi:weather-sunny"
	default:
		s.IconString = "mdi:theme-light-dark"
	}
	return s
}

func getProp(ctx context.Context, prop string) string {
	var value dbus.Variant
	var err error
	if value, err = dbusx.GetData[dbus.Variant](ctx,
		dbusx.SessionBus,
		desktopPortalPath,
		desktopPortalInterface,
		settingsPortalInterface+".Read",
		"org.freedesktop.appearance",
		prop); err != nil {
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
