// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:unused-receiver
package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/eclipse/paho.golang/paho"
	slogctx "github.com/veqryn/slog-context"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/pkg/linux/webcam"
)

const (
	startIcon  = "mdi:play"
	stopIcon   = "mdi:stop"
	statusIcon = "mdi:webcam"

	startedState = "Recording"
	stoppedState = "Not Recording"

	defaultCameraDevice     = "/dev/video0"
	defaultHeight       int = 640
	defaultWidth        int = 480
	defaultFps          int = 30

	cameraPreferencesID = "sensors.media.video"
)

var ErrInitCameraControls = errors.New("could not init camera controls")

// CameraWorker represents all of the entities that make up a camera. This
// includes the entity for showing images, as well as button entities for
// start/stop commands and a sensor entity showing the recording status.
type CameraWorker struct {
	Images      *mqtthass.CameraEntity
	StartButton *mqtthass.ButtonEntity
	StopButton  *mqtthass.ButtonEntity
	Status      *mqtthass.SensorEntity
	prefs       *CameraWorkerPrefs
	state       string
	MsgCh       chan mqttapi.Msg
	cancelFunc  context.CancelFunc
	ffmpegPath  string
}

// CameraWorkerPrefs are the preferences a user can set for the CameraWorker.
type CameraWorkerPrefs struct {
	*workers.CommonWorkerPrefs

	CameraDevice string `toml:"camera_device"`
	Height       int    `toml:"camera_video_height"`
	Width        int    `toml:"camera_video_width"`
	Fps          int    `toml:"camera_fps"`
}

// NewCameraWorker is called by the OS controller to provide the entities for a camera.
func NewCameraWorker(ctx context.Context, mqttDevice *mqtthass.Device) (*CameraWorker, error) {
	var err error
	worker := &CameraWorker{
		MsgCh: make(chan mqttapi.Msg),
	}

	worker.ffmpegPath, err = exec.LookPath("ffmpeg")
	if err != nil {
		return worker, fmt.Errorf("find ffmpeg executable: %w", err)
	}

	defaultPrefs := &CameraWorkerPrefs{
		CameraDevice: defaultCameraDevice,
		Width:        defaultWidth,
		Height:       defaultHeight,
		Fps:          defaultFps,
	}
	worker.prefs, err = workers.LoadWorkerPreferences(cameraPreferencesID, defaultPrefs)
	if err != nil {
		return worker, errors.Join(ErrInitCameraControls, err)
	}

	worker.Images = mqtthass.NewCameraEntity().
		WithDetails(
			mqtthass.App(config.AppName+"_"+mqttDevice.Name),
			mqtthass.Name("Webcam"),
			mqtthass.ID(mqttDevice.Name+"_camera"),
			mqtthass.OriginInfo(mqtt.Origin()),
			mqtthass.DeviceInfo(mqttDevice),
		)

	worker.StartButton = mqtthass.NewButtonEntity().
		WithDetails(
			mqtthass.App(config.AppName+"_"+mqttDevice.Name),
			mqtthass.Name("Start Webcam"),
			mqtthass.ID(mqttDevice.Name+"_start_camera"),
			mqtthass.OriginInfo(mqtt.Origin()),
			mqtthass.DeviceInfo(mqttDevice),
			mqtthass.Icon(startIcon),
		).WithCommand(
		mqtthass.CommandCallback(func(_ *paho.Publish) {
			// Ignore if we are already streaming.
			if worker.cancelFunc != nil {
				slogctx.FromCtx(ctx).Warn("Already streaming webcam. Ignoring.")
				return
			}

			// Create a child context and channel for streaming.
			streamCtx, streamCancel := context.WithCancel(ctx)
			worker.cancelFunc = streamCancel
			framesCh := make(chan []byte)

			// Stream webcam.
			go webcam.Capture(
				streamCtx,
				worker.ffmpegPath,
				worker.prefs.CameraDevice,
				worker.prefs.Fps,
				worker.prefs.Width,
				worker.prefs.Height,
				framesCh,
			)

			go func() {
				for frame := range framesCh {
					worker.MsgCh <- *mqttapi.NewMsg(worker.Images.Topic, frame)
				}
			}()

			worker.state = startedState
			worker.MsgCh <- *mqttapi.NewMsg(worker.Status.StateTopic, []byte(worker.state))
			slogctx.FromCtx(ctx).Info("Started webcam recording.",
				slog.Time("timestamp", time.Now()))
		}))

	worker.StopButton = mqtthass.NewButtonEntity().
		WithDetails(
			mqtthass.App(config.AppName+"_"+mqttDevice.Name),
			mqtthass.Name("Stop Webcam"),
			mqtthass.ID(mqttDevice.Name+"_stop_camera"),
			mqtthass.OriginInfo(mqtt.Origin()),
			mqtthass.DeviceInfo(mqttDevice),
			mqtthass.Icon(stopIcon),
		).WithCommand(
		mqtthass.CommandCallback(func(_ *paho.Publish) {
			if worker.cancelFunc == nil {
				slogctx.FromCtx(ctx).Warn("Not streaming. Ignoring.")
				return
			}
			worker.cancelFunc()
			worker.state = stoppedState
			worker.MsgCh <- *mqttapi.NewMsg(worker.Status.StateTopic, []byte(worker.state))

			slogctx.FromCtx(ctx).Info("Stopped webcam recording.",
				slog.Time("timestamp", time.Now()))
		}),
	)

	worker.Status = mqtthass.NewSensorEntity().
		WithDetails(
			mqtthass.App(config.AppName+"_"+mqttDevice.Name),
			mqtthass.Name("Webcam Status"),
			mqtthass.ID(mqttDevice.Name+"_camera_status"),
			mqtthass.OriginInfo(mqtt.Origin()),
			mqtthass.DeviceInfo(mqttDevice),
			mqtthass.Icon(statusIcon),
		).
		WithState(
			mqtthass.StateCallback(func(_ ...any) (json.RawMessage, error) {
				return json.RawMessage(fmt.Sprintf("%q", worker.state)), nil
			}),
		)

	go func() {
		defer close(worker.MsgCh)
		worker.MsgCh <- *mqttapi.NewMsg(worker.Status.StateTopic, []byte(stoppedState))
		<-ctx.Done()
	}()

	return worker, nil
}

func (w *CameraWorker) Disabled() bool {
	return w.prefs.Disabled
}
