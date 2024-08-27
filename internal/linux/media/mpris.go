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

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

const (
	mediaStopIcon  = "mdi:stop"
	mediaPauseIcon = "mdi:pause"
	mediaPlayIcon  = "mdi:play"
	mediaOffIcon   = "mdi:music-note-off"

	mprisDBusPath      = "/org/mpris/MediaPlayer2"
	mprisDBusNamespace = "org.mpris.MediaPlayer2.Player"
)

type mprisMonitor struct {
	logger           *slog.Logger
	mediaStateEntity *mqtthass.SensorEntity
	msgCh            chan *mqttapi.Msg
	mediaState       string
}

//nolint:lll
func MPRISControl(ctx context.Context, api *dbusx.DBusAPI, parentLogger *slog.Logger, device *mqtthass.Device, msgCh chan *mqttapi.Msg) (*mqtthass.SensorEntity, error) {
	mprisMonitor := &mprisMonitor{
		logger: parentLogger,
		msgCh:  msgCh,
	}

	mprisMonitor.mediaStateEntity = mqtthass.AsSensor(
		mqtthass.NewEntity(preferences.AppName, "Media State", device.Name+"_media_state").
			WithOriginInfo(preferences.MQTTOrigin()).
			WithDeviceInfo(device).
			WithIcon(mediaOffIcon).
			WithValueTemplate("{{ value_json.value }}").
			WithStateCallback(mprisMonitor.mprisStateCallback),
	)

	dbusAPI, err := api.GetBus(ctx, dbusx.SessionBus)
	if err != nil {
		return nil, fmt.Errorf("could not connect to D-Bus: %w", err)
	}

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(mprisDBusPath),
		dbusx.MatchPropChanged(),
		dbusx.MatchArgNameSpace(mprisDBusNamespace),
	).Start(ctx, dbusAPI)
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
