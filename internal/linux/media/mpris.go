// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	mediaStopIcon  = "mdi:stop"
	mediaPauseIcon = "mdi:pause"
	mediaPlayIcon  = "mdi:play"
	mediaOffIcon   = "mdi:music-note-off"

	mprisDBusPath      = "/org/mpris/MediaPlayer2"
	mprisDBusNamespace = "org.mpris.MediaPlayer2.Player"

	mprisPreferencesID = "mpris"
)

type mprisMonitor struct {
	logger           *slog.Logger
	mediaStateEntity *mqtthass.SensorEntity
	msgCh            chan *mqttapi.Msg
	mediaState       string
	prefs            *preferences.CommonWorkerPrefs
}

func (m *mprisMonitor) PreferencesID() string {
	return mprisPreferencesID
}

func (m *mprisMonitor) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func MPRISControl(ctx context.Context, device *mqtthass.Device, msgCh chan *mqttapi.Msg) (*mqtthass.SensorEntity, error) {
	var err error

	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return nil, linux.ErrNoSessionBus
	}

	mprisMonitor := &mprisMonitor{
		logger: logging.FromContext(ctx).With(slog.String("controller", "mpris")),
		msgCh:  msgCh,
	}

	mprisMonitor.prefs, err = preferences.LoadWorker(ctx, mprisMonitor)
	if err != nil {
		return nil, fmt.Errorf("could not load preferences: %w", err)
	}

	//nolint:nilnil
	if mprisMonitor.prefs.IsDisabled() {
		return nil, nil
	}

	mprisMonitor.mediaStateEntity = mqtthass.NewSensorEntity().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Media State"),
			mqtthass.ID(device.Name+"_media_state"),
			mqtthass.OriginInfo(preferences.MQTTOrigin()),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon(mediaOffIcon),
		).
		WithState(
			mqtthass.StateCallback(mprisMonitor.mprisStateCallback),
			mqtthass.ValueTemplate("{{ value_json.value }}"),
		)

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(mprisDBusPath),
		dbusx.MatchPropChanged(),
		dbusx.MatchArgNameSpace(mprisDBusNamespace),
	).Start(ctx, bus)
	if err != nil {
		return nil, fmt.Errorf("could not watch D-Bus for MPRIS signals: %w", err)
	}

	// Watch for power profile changes.
	go func() {
		mprisMonitor.logger.Debug("Monitoring for MPRIS signals.")

		for {
			select {
			case <-ctx.Done():
				mprisMonitor.logger.Debug("Stopped monitoring for MPRIS signals.")

				return
			case event := <-triggerCh:
				changed, status, err := dbusx.HasPropertyChanged[string](event.Content, "PlaybackStatus")
				if err != nil {
					mprisMonitor.logger.Warn("Could not parse received D-Bus signal.", slog.Any("error", err))
				} else {
					if changed {
						mprisMonitor.publishPlaybackState(status)
					}
				}
			}
		}
	}()

	return mprisMonitor.mediaStateEntity, nil
}

func (m *mprisMonitor) mprisStateCallback(_ ...any) (json.RawMessage, error) {
	return json.RawMessage(`{ "value": ` + m.mediaState + ` }`), nil
}

func (m *mprisMonitor) publishPlaybackState(state string) {
	m.mediaState = state

	switch m.mediaState {
	case "Playing":
		m.mediaStateEntity.Icon = mediaPlayIcon
	case "Paused":
		m.mediaStateEntity.Icon = mediaPauseIcon
	case "Stopped":
		m.mediaStateEntity.Icon = mediaStopIcon
	default:
		m.mediaStateEntity.Icon = mediaOffIcon
	}

	msg, err := m.mediaStateEntity.MarshalState()
	if err != nil {
		m.logger.Warn("Could not publish MPRIS state.", slog.Any("error", err))

		return
	}
	m.msgCh <- msg
}
