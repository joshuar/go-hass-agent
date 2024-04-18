// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package desktop

import (
	"context"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"

	"github.com/mandykoh/prism/srgb"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	desktopPortalPath              = "/org/freedesktop/portal/desktop"
	desktopPortalSettingsInterface = "org.freedesktop.impl.portal.Settings"
	desktopPortalSettingsSignal    = "org.freedesktop.impl.portal.Settings.SettingChanged"
)

type desktopSettingSensor struct {
	linux.Sensor
}

func Updater(ctx context.Context) chan sensor.Details {
	sensorCh := make(chan sensor.Details)
	portalDest := linux.FindPortal()
	if portalDest == "" {
		log.Warn().
			Msg("Unable to monitor for active applications. No app tracking available.")
		close(sensorCh)
		return sensorCh
	}

	reqCtx, cancelReq := context.WithTimeout(ctx, 15*time.Second)
	defer cancelReq()
	settingsReq := dbusx.NewBusRequest(reqCtx, dbusx.SessionBus).
		Path(desktopPortalPath).
		Destination("org.freedesktop.portal.Desktop")
	go func() {
		if v, err := dbusx.GetData[dbus.Variant](settingsReq,
			"org.freedesktop.portal.Settings.Read",
			"org.freedesktop.appearance",
			"accent-color"); err != nil {
			log.Warn().Err(err).Msg("Could not retrieve accent color from D-Bus.")
		} else {
			s := getAccentColor(v)
			sensorCh <- newAccentColorSensor(s)
		}
	}()
	go func() {
		if v, err := dbusx.GetData[dbus.Variant](settingsReq,
			"org.freedesktop.portal.Settings.Read",
			"org.freedesktop.appearance",
			"color-scheme"); err != nil {
			log.Warn().Err(err).Msg("Could not retrieve color scheme type from D-Bus.")
		} else {
			s := getColorSchemePref(v)
			sensorCh <- newColorSchemeSensor(s)
		}
	}()

	err := dbusx.NewBusRequest(ctx, dbusx.SessionBus).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(desktopPortalPath),
			dbus.WithMatchInterface(desktopPortalSettingsInterface),
		}).
		Handler(func(s *dbus.Signal) {
			if s.Name != desktopPortalSettingsSignal {
				return
			}
			prop, ok := s.Body[1].(string)
			if !ok {
				return
			}
			switch prop {
			case "color-scheme":
				s := getColorSchemePref(s.Body[2])
				sensorCh <- newColorSchemeSensor(s)
			case "accent-color":
				s := getAccentColor(s.Body[2])
				sensorCh <- newAccentColorSensor(s)
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Failed to create desktop settings D-Bus watch.")
		close(sensorCh)
	}
	go func() {
		defer close(sensorCh)
		<-ctx.Done()
		log.Debug().Msg("Stopped desktop settings sensors.")
	}()
	return sensorCh
}

func getColorSchemePref(value any) string {
	v, ok := value.(dbus.Variant)
	if !ok {
		return sensor.StateUnknown
	}
	scheme := dbusx.VariantToValue[uint32](v)
	switch scheme {
	case 1:
		return "dark"
	case 2:
		return "light"
	default:
		return sensor.StateUnknown
	}
}

func getAccentColor(value any) string {
	v, ok := value.(dbus.Variant)
	if !ok {
		return sensor.StateUnknown
	}
	values := dbusx.VariantToValue[[]any](v)

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

func newAccentColorSensor(accent string) *desktopSettingSensor {
	s := &desktopSettingSensor{}
	s.SensorSrc = linux.DataSrcDbus
	s.SensorTypeValue = linux.SensorAccentColor
	s.Value = accent
	s.IconString = "mdi:palette"
	return s
}

func newColorSchemeSensor(scheme string) *desktopSettingSensor {
	s := &desktopSettingSensor{}
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
