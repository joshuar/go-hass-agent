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
	"slices"

	"github.com/eclipse/paho.golang/paho"

	"github.com/blackjack/webcam"
	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	startIcon  = "mdi:play"
	stopIcon   = "mdi:stop"
	statusIcon = "mdi:webcam"

	startedState = "Recording"
	stoppedState = "Not Recording"
)

// Some defaults for the device file, formats and image size.
var (
	defaultDevice        = "/dev/video0"
	preferredFmts        = []string{"Motion-JPEG"}
	defaultHeight uint32 = 640
	defaultWidth  uint32 = 480
)

// CameraEntities represents all of the entities that make up a camera. This
// includes the entity for showing images, as well as button entities for
// start/stop commands and a sensor entity showing the recording status.
type CameraEntities struct {
	Images      *mqtthass.CameraEntity
	StartButton *mqtthass.ButtonEntity
	StopButton  *mqtthass.ButtonEntity
	Status      *mqtthass.SensorEntity
	camera      *webcam.Webcam
	state       string
}

// NewCameraControl is called by the OS controller to provide the entities for a camera.
func NewCameraControl(_ context.Context, msgCh chan *mqttapi.Msg, mqttDevice *mqtthass.Device) *CameraEntities {
	entities := &CameraEntities{}

	entities.Images = mqtthass.NewCameraEntity().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Webcam"),
			mqtthass.ID(mqttDevice.Name+"_camera"),
			mqtthass.OriginInfo(preferences.MQTTOrigin()),
			mqtthass.DeviceInfo(mqttDevice),
		)

	entities.StartButton = mqtthass.NewButtonEntity().
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
			entities.camera, err = openCamera(defaultDevice)
			if err != nil {
				slog.Error("Could not open camera device.",
					slog.Any("error", err))
				return
			}

			slog.Info("Start recording webcam.")

			entities.state = startedState

			go publishImages(entities.camera, entities.Images.Topic, msgCh)
			msgCh <- mqttapi.NewMsg(entities.Status.StateTopic, []byte(entities.state))
		}))

	entities.StopButton = mqtthass.NewButtonEntity().
		WithDetails(
			mqtthass.App(preferences.AppName),
			mqtthass.Name("Stop Webcam"),
			mqtthass.ID(mqttDevice.Name+"_stop_camera"),
			mqtthass.OriginInfo(preferences.MQTTOrigin()),
			mqtthass.DeviceInfo(mqttDevice),
			mqtthass.Icon(stopIcon),
		).WithCommand(
		mqtthass.CommandCallback(func(_ *paho.Publish) {
			if err := entities.camera.StopStreaming(); err != nil {
				slog.Error("Stop streaming failed.", slog.Any("error", err))
			}

			if err := entities.camera.Close(); err != nil {
				slog.Error("Close camera failed.", slog.Any("error", err))
			}

			entities.state = stoppedState
			msgCh <- mqttapi.NewMsg(entities.Status.StateTopic, []byte(entities.state))

			slog.Info("Stop recording webcam.")
		}),
	)

	entities.Status = mqtthass.NewSensorEntity().
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
				return json.RawMessage(entities.state), nil
			}),
		)

	go func() {
		msgCh <- mqttapi.NewMsg(entities.Status.StateTopic, []byte(stoppedState))
	}()

	return entities
}

// openCamera opens the camera device and ensures that it has a preferred image
// format, framerate and dimensions.
func openCamera(cameraDevice string) (*webcam.Webcam, error) {
	cam, err := webcam.Open(cameraDevice)
	if err != nil {
		return nil, fmt.Errorf("could not open camera %s: %w", cameraDevice, err)
	}

	// select pixel format
	var preferredFormat webcam.PixelFormat

	for format, desc := range cam.GetSupportedFormats() {
		if slices.Contains(preferredFmts, desc) {
			preferredFormat = format
			break
		}
	}

	if preferredFormat == 0 {
		return nil, errors.New("could not determine an appropriate format")
	}

	_, _, _, err = cam.SetImageFormat(preferredFormat, defaultWidth, defaultHeight)
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
