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
	"slices"

	"github.com/eclipse/paho.golang/paho"
	slogctx "github.com/veqryn/slog-context"

	"github.com/blackjack/webcam"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/config"
)

const (
	startIcon  = "mdi:play"
	stopIcon   = "mdi:stop"
	statusIcon = "mdi:webcam"

	startedState = "Recording"
	stoppedState = "Not Recording"

	defaultCameraDevice        = "/dev/video0"
	defaultHeight       uint32 = 640
	defaultWidth        uint32 = 480

	cameraPreferencesID = "sensors.media.video"
)

var ErrInitCameraControls = errors.New("could not init camera controls")

var defaultPreferredFmts = []string{"Motion-JPEG"}

// CameraWorker represents all of the entities that make up a camera. This
// includes the entity for showing images, as well as button entities for
// start/stop commands and a sensor entity showing the recording status.
type CameraWorker struct {
	Images      *mqtthass.CameraEntity
	StartButton *mqtthass.ButtonEntity
	StopButton  *mqtthass.ButtonEntity
	Status      *mqtthass.SensorEntity
	camera      *webcam.Webcam
	prefs       *CameraWorkerPrefs
	state       string
	MsgCh       chan mqttapi.Msg
}

// CameraWorkerPrefs are the preferences a user can set for the CameraWorker.
type CameraWorkerPrefs struct {
	*workers.CommonWorkerPrefs

	CameraDevice  string   `toml:"camera_device"`
	CameraFormats []string `toml:"camera_formats"`
	Height        uint32   `toml:"camera_video_height"`
	Width         uint32   `toml:"camera_video_width"`
}

// NewCameraWorker is called by the OS controller to provide the entities for a camera.
func NewCameraWorker(ctx context.Context, mqttDevice *mqtthass.Device) (*CameraWorker, error) {
	var err error
	worker := &CameraWorker{
		MsgCh: make(chan mqttapi.Msg),
	}

	defaultPrefs := &CameraWorkerPrefs{
		CameraDevice:  defaultCameraDevice,
		Width:         defaultWidth,
		Height:        defaultHeight,
		CameraFormats: defaultPreferredFmts,
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
			var err error
			// Open the camera device.
			worker.camera, err = worker.openCamera()
			if err != nil {
				slogctx.FromCtx(ctx).Error("Could not open camera device.",
					slog.Any("error", err))
				return
			}

			slogctx.FromCtx(ctx).Info("Start recording webcam.")

			worker.state = startedState

			go publishImages(ctx, worker.camera, worker.Images.Topic, worker.MsgCh)
			worker.MsgCh <- *mqttapi.NewMsg(worker.Status.StateTopic, []byte(worker.state))
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
			if err := worker.camera.StopStreaming(); err != nil {
				slogctx.FromCtx(ctx).Error("Stop streaming failed.", slog.Any("error", err))
			}

			if err := worker.camera.Close(); err != nil {
				slogctx.FromCtx(ctx).Error("Close camera failed.", slog.Any("error", err))
			}

			worker.state = stoppedState
			worker.MsgCh <- *mqttapi.NewMsg(worker.Status.StateTopic, []byte(worker.state))

			slogctx.FromCtx(ctx).Info("Stop recording webcam.")
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

// openCamera opens the camera device and ensures that it has a preferred image
// format, framerate and dimensions.
func (w *CameraWorker) openCamera() (*webcam.Webcam, error) {
	cam, err := webcam.Open(w.prefs.CameraDevice)
	if err != nil {
		return nil, fmt.Errorf("could not open camera %s: %w", w.prefs.CameraDevice, err)
	}

	// select pixel format
	var preferredFormat webcam.PixelFormat

	for format, desc := range cam.GetSupportedFormats() {
		if slices.Contains(w.prefs.CameraFormats, desc) {
			preferredFormat = format
			break
		}
	}

	if preferredFormat == 0 {
		return nil, errors.New("could not determine an appropriate format")
	}

	_, _, _, err = cam.SetImageFormat(preferredFormat, w.prefs.Width, w.prefs.Height)
	if err != nil {
		return nil, fmt.Errorf("could not set camera parameters: %w", err)
	}

	return cam, nil
}

// publishImages loops over the received frames from the camera and wraps them
// as a MQTT message to be sent back on the bus.
func publishImages(ctx context.Context, cam *webcam.Webcam, topic string, msgCh chan mqttapi.Msg) {
	if err := cam.StartStreaming(); err != nil {
		slogctx.FromCtx(ctx).Error("Could not start recording", slog.Any("error", err))

		return
	}

	for {
		if err := cam.WaitForFrame(uint32(5)); err != nil && errors.Is(err, &webcam.Timeout{}) {
			continue
		}

		frame, err := cam.ReadFrame()
		if len(frame) == 0 || err != nil {
			break
		}

		msgCh <- *mqttapi.NewMsg(topic, frame)
	}
}
