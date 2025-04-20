// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"
	slogctx "github.com/veqryn/slog-context"

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
)

var ErrInitMPRISWorker = errors.New("could not init MPRIS worker")

type MPRISWorker struct {
	MPRISStatus *mqtthass.SensorEntity
	MsgCh       chan mqttapi.Msg
	mediaState  string
	prefs       *preferences.CommonWorkerPrefs
}

func (m *MPRISWorker) PreferencesID() string {
	return mprisPrefID
}

func (m *MPRISWorker) DefaultPreferences() preferences.CommonWorkerPrefs {
	return preferences.CommonWorkerPrefs{}
}

func (m *MPRISWorker) mprisStateCallback(_ ...any) (json.RawMessage, error) {
	return json.RawMessage(`{ "value": ` + m.mediaState + ` }`), nil
}

func (m *MPRISWorker) publishPlaybackState(ctx context.Context, state string) {
	m.mediaState = state

	switch m.mediaState {
	case "Playing":
		m.MPRISStatus.Icon = mediaPlayIcon
	case "Paused":
		m.MPRISStatus.Icon = mediaPauseIcon
	case "Stopped":
		m.MPRISStatus.Icon = mediaStopIcon
	default:
		m.MPRISStatus.Icon = mediaOffIcon
	}

	msg, err := m.MPRISStatus.MarshalState()
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not publish MPRIS state.", slog.Any("error", err))

		return
	}
	m.MsgCh <- *msg
}

func NewMPRISWorker(ctx context.Context, device *mqtthass.Device) (*MPRISWorker, error) {
	var err error

	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return nil, errors.Join(ErrInitMPRISWorker, linux.ErrNoSessionBus)
	}

	worker := &MPRISWorker{
		MsgCh: make(chan mqttapi.Msg),
	}

	worker.prefs, err = preferences.LoadWorker(worker)
	if err != nil {
		return nil, errors.Join(ErrInitMPRISWorker, err)
	}

	//nolint:nilnil
	if worker.prefs.IsDisabled() {
		return nil, nil
	}

	worker.MPRISStatus = mqtthass.NewSensorEntity().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Media State"),
			mqtthass.ID(device.Name+"_media_state"),
			mqtthass.OriginInfo(preferences.MQTTOrigin()),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon(mediaOffIcon),
		).
		WithState(
			mqtthass.StateCallback(worker.mprisStateCallback),
			mqtthass.ValueTemplate("{{ value_json.value }}"),
		)

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(mprisDBusPath),
		dbusx.MatchPropChanged(),
		dbusx.MatchArgNameSpace(mprisDBusNamespace),
	).Start(ctx, bus)
	if err != nil {
		return nil, errors.Join(ErrInitMPRISWorker,
			fmt.Errorf("could not watch D-Bus for MPRIS signals: %w", err))
	}

	// Watch for power profile changes.
	go func() {
		slogctx.FromCtx(ctx).Debug("Monitoring for MPRIS signals.")
		for {
			select {
			case <-ctx.Done():
				slogctx.FromCtx(ctx).Debug("Stopped monitoring for MPRIS signals.")
				return
			case event := <-triggerCh:
				changed, status, err := dbusx.HasPropertyChanged[string](event.Content, "PlaybackStatus")
				if err != nil {
					slogctx.FromCtx(ctx).Warn("Could not parse received D-Bus signal.", slog.Any("error", err))
				} else if changed {
					worker.publishPlaybackState(ctx, status)
				}
			}
		}
	}()

	return worker, nil
}
