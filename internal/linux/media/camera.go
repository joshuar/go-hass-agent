// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
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

	"github.com/blackjack/webcam"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
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

	cameraPreferencesID = "camera_controls"
)

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
}

// CameraWorkerPrefs are the preferences a user can set for the CameraWorker.
type CameraWorkerPrefs struct {
	CameraDevice  string   `toml:"camera_device" comment:"The camera device to control. Defaults to /dev/video0."`
	CameraFormats []string `toml:"camera_formats" comment:"Preferred camera video formats. Defaults to Motion-JPEG."`
	Height        uint32   `toml:"camera_video_height" comment:"Height (in pixels) of the camera image. Defaults to 640px."`
	Width         uint32   `toml:"camera_video_width" comment:"With (in pixels) of the camera image. Defaults to 480px."`
	preferences.CommonWorkerPrefs
}

func (w *CameraWorker) PreferencesID() string {
	return cameraPreferencesID
}

func (w *CameraWorker) DefaultPreferences() CameraWorkerPrefs {
	return CameraWorkerPrefs{
		CameraDevice:  defaultCameraDevice,
		Width:         defaultWidth,
		Height:        defaultHeight,
		CameraFormats: defaultPreferredFmts,
	}
}

func (w *CameraWorker) Disabled() bool {
	return w.prefs.Disabled
}

// NewCameraControl is called by the OS controller to provide the entities for a camera.
func NewCameraControl(ctx context.Context, msgCh chan *mqttapi.Msg, mqttDevice *mqtthass.Device) (*CameraWorker, error) {
	var err error

	worker := &CameraWorker{}

	worker.prefs, err = preferences.LoadWorker(ctx, worker)
	if err != nil {
		return worker, fmt.Errorf("could not load preferences: %w", err)
	}

	worker.Images = mqtthass.NewCameraEntity().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Webcam"),
			mqtthass.ID(mqttDevice.Name+"_camera"),
			mqtthass.OriginInfo(preferences.MQTTOrigin()),
			mqtthass.DeviceInfo(mqttDevice),
		)

	worker.StartButton = mqtthass.NewButtonEntity().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Start Webcam"),
			mqtthass.ID(mqttDevice.Name+"_start_camera"),
			mqtthass.OriginInfo(preferences.MQTTOrigin()),
			mqtthass.DeviceInfo(mqttDevice),
			mqtthass.Icon(startIcon),
		).WithCommand(
		mqtthass.CommandCallback(func(_ *paho.Publish) {
			var err error
			// Open the camera device.
			worker.camera, err = worker.openCamera()
			if err != nil {
				slog.Error("Could not open camera device.",
					slog.Any("error", err))
				return
			}

			slog.Info("Start recording webcam.")

			worker.state = startedState

			go publishImages(worker.camera, worker.Images.Topic, msgCh)
			msgCh <- mqttapi.NewMsg(worker.Status.StateTopic, []byte(worker.state))
		}))

	worker.StopButton = mqtthass.NewButtonEntity().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Stop Webcam"),
			mqtthass.ID(mqttDevice.Name+"_stop_camera"),
			mqtthass.OriginInfo(preferences.MQTTOrigin()),
			mqtthass.DeviceInfo(mqttDevice),
			mqtthass.Icon(stopIcon),
		).WithCommand(
		mqtthass.CommandCallback(func(_ *paho.Publish) {
			if err := worker.camera.StopStreaming(); err != nil {
				slog.Error("Stop streaming failed.", slog.Any("error", err))
			}

			if err := worker.camera.Close(); err != nil {
				slog.Error("Close camera failed.", slog.Any("error", err))
			}

			worker.state = stoppedState
			msgCh <- mqttapi.NewMsg(worker.Status.StateTopic, []byte(worker.state))

			slog.Info("Stop recording webcam.")
		}),
	)

	worker.Status = mqtthass.NewSensorEntity().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Webcam Status"),
			mqtthass.ID(mqttDevice.Name+"_camera_status"),
			mqtthass.OriginInfo(preferences.MQTTOrigin()),
			mqtthass.DeviceInfo(mqttDevice),
			mqtthass.Icon(statusIcon),
		).
		WithState(
			mqtthass.StateCallback(func(_ ...any) (json.RawMessage, error) {
				return json.RawMessage(worker.state), nil
			}),
		)

	go func() {
		msgCh <- mqttapi.NewMsg(worker.Status.StateTopic, []byte(stoppedState))
	}()

	return worker, nil
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
func publishImages(cam *webcam.Webcam, topic string, msgCh chan *mqttapi.Msg) {
	if err := cam.StartStreaming(); err != nil {
		slog.Error("Could not start recording", slog.Any("error", err))

		return
	}

	for {
		err := cam.WaitForFrame(uint32(5))
		if err != nil && errors.Is(err, &webcam.Timeout{}) {
			continue
		}

		frame, err := cam.ReadFrame()
		if len(frame) == 0 || err != nil {
			break
		}

		msgCh <- mqttapi.NewMsg(topic, frame)
	}
}
