// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/linux/media"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

type mqttWorker struct {
	msgs          chan *mqttapi.Msg
	sensors       []*mqtthass.SensorEntity
	buttons       []*mqtthass.ButtonEntity
	numbers       []*mqtthass.NumberEntity[int]
	switches      []*mqtthass.SwitchEntity
	controls      []*mqttapi.Subscription
	binarySensors []*mqtthass.SensorEntity
	cameras       []*mqtthass.CameraEntity
}

type linuxMQTTController struct {
	*mqttWorker
	logger *slog.Logger
}

// stateEntity is a convienience interface to avoid duplicating a lot of loop content
// when configuring the controller.
type stateEntity interface {
	MarshalConfig() (*mqttapi.Msg, error)
}

// commandEntity is a convienience interface to avoid duplicating a lot of loop content
// when configuring the controller.
type commandEntity interface {
	stateEntity
	MarshalSubscription() (*mqttapi.Subscription, error)
}

func (c *linuxMQTTController) Subscriptions() []*mqttapi.Subscription {
	totalLength := len(c.buttons) + len(c.numbers) + len(c.switches) + len(c.cameras)
	subs := make([]*mqttapi.Subscription, 0, totalLength)

	// Create subscriptions for buttons.
	for _, button := range c.buttons {
		subs = append(subs, c.generateSubscription(button))
	}
	// Create subscriptions for numbers.
	for _, number := range c.numbers {
		subs = append(subs, c.generateSubscription(number))
	}
	// Create subscriptions for switches.
	for _, sw := range c.switches {
		subs = append(subs, c.generateSubscription(sw))
	}
	// Add subscriptions for any additional controls.
	subs = append(subs, c.controls...)

	return subs
}

func (c *linuxMQTTController) Configs() []*mqttapi.Msg {
	totalLength := len(c.sensors) + len(c.binarySensors) + len(c.buttons) + len(c.switches) + len(c.numbers) + len(c.cameras)
	configs := make([]*mqttapi.Msg, 0, totalLength)

	// Create sensor configs.
	for _, sensorEntity := range c.sensors {
		configs = append(configs, c.generateConfig(sensorEntity))
	}
	// Create binary sensor configs.
	for _, binarySensorEntity := range c.binarySensors {
		configs = append(configs, c.generateConfig(binarySensorEntity))
	}
	// Create button configs.
	for _, buttonEntity := range c.buttons {
		configs = append(configs, c.generateConfig(buttonEntity))
	}
	// Create number configs.
	for _, numberEntity := range c.numbers {
		configs = append(configs, c.generateConfig(numberEntity))
	}
	// Create switch configs.
	for _, switchEntity := range c.switches {
		configs = append(configs, c.generateConfig(switchEntity))
	}
	// Create camera configs.
	for _, cameraEntity := range c.cameras {
		configs = append(configs, c.generateConfig(cameraEntity))
	}

	return configs
}

func (c *linuxMQTTController) Msgs() chan *mqttapi.Msg {
	return c.msgs
}

// generateConfig is a helper function to avoid duplicate code around generating
// an entity subscription.
func (c *linuxMQTTController) generateSubscription(e commandEntity) *mqttapi.Subscription {
	sub, err := e.MarshalSubscription()
	if err != nil {
		c.logger.Warn("Could not create subscription.", slog.Any("error", err))

		return nil
	}

	return sub
}

// generateConfig is a helper function to avoid duplicate code around generating
// an entity config.
func (c *linuxMQTTController) generateConfig(e stateEntity) *mqttapi.Msg {
	cfg, err := e.MarshalConfig()
	if err != nil {
		c.logger.Warn("Could not create config.", slog.Any("error", err.Error()))

		return nil
	}

	return cfg
}

// newOSMQTTController initializes the list of MQTT workers for sensors and
// returns those that are supported on this device.
func newOSMQTTController(ctx context.Context, mqttDevice *mqtthass.Device) MQTTController {
	ctx = linux.NewContext(ctx)
	logger := logging.FromContext(ctx).With(slog.Group("linux", slog.String("controller", "mqtt")))

	mqttController := &linuxMQTTController{
		mqttWorker: &mqttWorker{
			msgs: make(chan *mqttapi.Msg),
		},
	}

	// Add the power controls (suspend, resume, poweroff, etc.).
	powerEntities, err := power.NewPowerControl(ctx, mqttDevice)
	if err != nil {
		logger.Warn("Could not create power controls.", slog.Any("error", err))
	} else {
		mqttController.buttons = append(mqttController.buttons, powerEntities...)
	}
	// Add the screen lock controls.
	screenControls, err := power.NewScreenLockControl(ctx, mqttDevice)
	if err != nil {
		logger.Warn("Could not create screen lock controls.", slog.Any("error", err))
	} else {
		mqttController.buttons = append(mqttController.buttons, screenControls...)
	}
	// Add the volume controls.
	volEntity, muteEntity := media.VolumeControl(ctx, mqttController.Msgs(), mqttDevice)
	if volEntity != nil && muteEntity != nil {
		mqttController.numbers = append(mqttController.numbers, volEntity)
		mqttController.switches = append(mqttController.switches, muteEntity)
	}
	// Add media control.
	mprisEntity, err := media.MPRISControl(ctx, mqttDevice, mqttController.Msgs())
	if err != nil {
		logger.Warn("Could not activate MPRIS controller.", slog.Any("error", err))
	} else {
		mqttController.sensors = append(mqttController.sensors, mprisEntity)
	}
	// Add camera control.
	cameraEntities := media.NewCameraControl(ctx, mqttController.Msgs(), mqttDevice)
	if cameraEntities != nil {
		mqttController.buttons = append(mqttController.buttons, cameraEntities.StartButton, cameraEntities.StopButton)
		mqttController.cameras = append(mqttController.cameras, cameraEntities.Images)
		mqttController.sensors = append(mqttController.sensors, cameraEntities.Status)
	}

	// Add the D-Bus command action.
	dbusCmdController, err := system.NewDBusCommandSubscription(ctx, mqttDevice)
	if err != nil {
		logger.Warn("Could not activate D-Bus commands controller.", slog.Any("error", err))
	} else {
		mqttController.controls = append(mqttController.controls, dbusCmdController)
	}

	go func() {
		defer close(mqttController.msgs)
		<-ctx.Done()
	}()

	return mqttController
}
