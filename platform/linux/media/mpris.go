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

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
	"github.com/joshuar/go-hass-agent/platform/linux"
)

const (
	mediaStopIcon  = "mdi:stop"
	mediaPauseIcon = "mdi:pause"
	mediaPlayIcon  = "mdi:play"
	mediaOffIcon   = "mdi:music-note-off"

	mprisDBusPath      = "/org/mpris/MediaPlayer2"
	mprisDBusNamespace = "org.mpris.MediaPlayer2"
)

var ErrInitMPRISWorker = errors.New("could not init MPRIS worker")

type MPRISWorker struct {
	MPRISStatus *mqtthass.SensorEntity
	MsgCh       chan mqttapi.Msg
	mediaState  string
	prefs       *workers.CommonWorkerPrefs
}

func NewMPRISWorker(ctx context.Context, device *mqtthass.Device) (*MPRISWorker, error) {
	worker := &MPRISWorker{
		MsgCh: make(chan mqttapi.Msg),
	}

	var err error

	bus, ok := linux.CtxGetSessionBus(ctx)
	if !ok {
		return worker, errors.Join(ErrInitMPRISWorker, linux.ErrNoSessionBus)
	}

	defaultPrefs := &workers.CommonWorkerPrefs{}
	worker.prefs, err = workers.LoadWorkerPreferences(mprisPrefID, defaultPrefs)
	if err != nil {
		return worker, errors.Join(ErrInitMPRISWorker, err)
	}

	if worker.prefs.IsDisabled() {
		return worker, nil
	}

	worker.MPRISStatus = mqtthass.NewSensorEntity().
		WithDetails(
			mqtthass.App(config.AppName+"_"+device.Name),
			mqtthass.Name("Media State"),
			mqtthass.ID(device.Name+"_media_state"),
			mqtthass.OriginInfo(mqtt.Origin()),
			mqtthass.DeviceInfo(device),
			mqtthass.Icon(mediaOffIcon),
		).
		WithState(
			mqtthass.StateCallback(worker.mprisStateCallback),
			mqtthass.ValueTemplate("{{ value_json.value }}"),
		)

	triggerCh, err := dbusx.NewWatch(
		dbusx.MatchPath(mprisDBusPath),
		// dbusx.MatchPropChanged(),
		dbusx.MatchArgNameSpace(mprisDBusNamespace),
	).Start(ctx, bus)
	if err != nil {
		return worker, errors.Join(ErrInitMPRISWorker,
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
				if changed, status, err := dbusx.HasPropertyChanged[string](
					event.Content,
					"PlaybackStatus",
				); err != nil {
					slogctx.FromCtx(ctx).Warn("Could not parse received D-Bus signal.", slog.Any("error", err))
				} else if changed {
					worker.publishPlaybackState(ctx, status)
				}
			}
		}
	}()

	// TODO: send the state on agent startup.

	return worker, nil
}

func (m *MPRISWorker) mprisStateCallback(_ ...any) (json.RawMessage, error) {
	return json.RawMessage(`{ "value": ` + m.mediaState + ` }`), nil
}

func (m *MPRISWorker) publishPlaybackState(ctx context.Context, state string) {
	m.mediaState = fmt.Sprintf("%q", state)

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
