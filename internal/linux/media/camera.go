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
	"time"

	"github.com/eclipse/paho.golang/paho"
	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"
	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"

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
	defaultDevice = "/dev/video0"
	preferredFmts = []v4l2.FourCCType{v4l2.PixelFmtMPEG, v4l2.PixelFmtMJPEG, v4l2.PixelFmtJPEG, v4l2.PixelFmtYUYV}
	defaultHeight = 640
	defaultWidth  = 480
)

// CameraEntities represents all of the entities that make up a camera. This
// includes the entity for showing images, as well as button entities for
// start/stop commands and a sensor entity showing the recording status.
type CameraEntities struct {
	Images      *mqtthass.ImageEntity
	StartButton *mqtthass.ButtonEntity
	StopButton  *mqtthass.ButtonEntity
	Status      *mqtthass.SensorEntity
}

// cameraControl is an internal struct that contains the data used to control
// the camera and populate the entities.
type cameraControl struct {
	device     *device.Device
	cancelFunc context.CancelFunc
	logger     *slog.Logger
	state      string
	fps        time.Duration
}

// NewCameraControl is called by the OS controller to provide the entities for a camera.
//
//nolint:lll
func NewCameraControl(ctx context.Context, msgCh chan *mqttapi.Msg, parentLogger *slog.Logger, mqttDevice *mqtthass.Device) *CameraEntities {
	camera := newCamera(parentLogger)
	entities := &CameraEntities{}

	entities.Images = mqtthass.AsImage(mqtthass.NewEntity(preferences.AppName, "Webcam", mqttDevice.Name+"_camera").
		WithDeviceInfo(mqttDevice).
		WithDefaultOriginInfo(), mqtthass.ModeImage)

	entities.StartButton = mqtthass.AsButton(mqtthass.NewEntity(preferences.AppName, "Start Webcam", mqttDevice.Name+"_start_camera").
		WithDeviceInfo(mqttDevice).
		WithDefaultOriginInfo().
		WithIcon(startIcon).
		WithCommandCallback(func(_ *paho.Publish) {
			err := camera.openCamera(defaultDevice)
			if err != nil {
				camera.logger.Error("Could not open camera device.", slog.Any("error", err))

				return
			}

			camCtx, cancelFunc := context.WithCancel(ctx)
			camera.cancelFunc = cancelFunc
			camera.state = startedState

			camera.logger.Debug("Start recording webcam.")

			go camera.publishImages(camCtx, entities.Images.GetImageTopic(), msgCh)
			msgCh <- mqttapi.NewMsg(entities.Status.StateTopic, []byte(camera.state))
		}))

	entities.StopButton = mqtthass.AsButton(mqtthass.NewEntity(preferences.AppName, "Stop Webcam", mqttDevice.Name+"_stop_camera").
		WithDeviceInfo(mqttDevice).
		WithDefaultOriginInfo().
		WithIcon(stopIcon).
		WithCommandCallback(func(_ *paho.Publish) {
			camera.state = stoppedState
			if camera.cancelFunc != nil {
				camera.cancelFunc()
				camera.logger.Debug("Stop recording webcam.")

				if err := camera.closeCamera(); err != nil {
					camera.logger.Error("Close camera failed.", slog.Any("error", err))
				}
			}
			msgCh <- mqttapi.NewMsg(entities.Status.StateTopic, []byte(camera.state))
		}))

	entities.Status = mqtthass.AsSensor(mqtthass.NewEntity(preferences.AppName, "Webcam Status", mqttDevice.Name+"_camera_status").
		WithDeviceInfo(mqttDevice).
		WithDefaultOriginInfo().
		WithIcon(statusIcon).
		WithValueTemplate("{{ value }}").
		WithStateCallback(func(_ ...any) (json.RawMessage, error) {
			return json.RawMessage(camera.state), nil
		}))

	go func() {
		msgCh <- mqttapi.NewMsg(entities.Status.StateTopic, []byte(camera.state))
	}()

	return entities
}

func newCamera(logger *slog.Logger) *cameraControl {
	return &cameraControl{
		logger: logger.WithGroup("camera"),
		state:  stoppedState,
	}
}

// openCamera opens the camera device and ensures that it has a preferred image
// format, framerate and dimensions.
func (c *cameraControl) openCamera(cameraDevice string) error {
	camDev, err := device.Open(cameraDevice)
	if err != nil {
		return fmt.Errorf("could not open camera %s: %w", cameraDevice, err)
	}

	fps, err := camDev.GetFrameRate()
	if err != nil {
		return fmt.Errorf("could not determine camera frame rate: %w", err)
	}

	fmtDescs, err := camDev.GetFormatDescriptions()
	if err != nil {
		return fmt.Errorf("could not determine camera formats: %w", err)
	}

	var fmtDesc *v4l2.FormatDescription
	for _, preferredFmt := range preferredFmts {
		fmtDesc = getFormats(fmtDescs, preferredFmt)
		if fmtDesc != nil {
			break
		}
	}

	if fmtDesc == nil {
		return fmt.Errorf("camera does not support any preferred formats: %w", err)
	}

	if err = camDev.SetPixFormat(v4l2.PixFormat{
		Width:       uint32(defaultWidth),
		Height:      uint32(defaultHeight),
		PixelFormat: fmtDesc.PixelFormat,
		Field:       v4l2.FieldNone,
	}); err != nil {
		return fmt.Errorf("could not configure camera: %w", err)
	}

	pixFmt, err := camDev.GetPixFormat()
	if err == nil {
		c.logger.Debug("Camera configured.", slog.Any("format", pixFmt), slog.Any("fps", fps))
	}

	c.device = camDev
	c.fps = time.Second / time.Duration(fps)

	return nil
}

// publishImages loops over the received frames from the camera and wraps them
// as a MQTT message to be sent back on the bus.
func (c *cameraControl) publishImages(ctx context.Context, topic string, msgCh chan *mqttapi.Msg) {
	if err := c.device.Start(ctx); err != nil {
		c.logger.Error("Could not start recording", slog.Any("error", err))

		return
	}

	for frame := range c.device.GetOutput() {
		c.logger.Log(ctx, mqttapi.LevelTrace, "Sending frame.")
		msgCh <- mqttapi.NewMsg(topic, frame)

		time.Sleep(c.fps)
	}
}

// closeCamera wraps the v4l2 camera close method.
func (c *cameraControl) closeCamera() error {
	if err := c.device.Close(); err != nil {
		return fmt.Errorf("could not close camera device: %w", err)
	}

	return nil
}

// getFormats finds an appropriate image format to use for the camera.
func getFormats(fmts []v4l2.FormatDescription, pixEncoding v4l2.FourCCType) *v4l2.FormatDescription {
	for _, desc := range fmts {
		if desc.PixelFormat == pixEncoding {
			return &desc
		}
	}

	return nil
}
