// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"log/slog"

	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux/media"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
)

// linuxMQTTWorker represents the Linux-specific OS MQTTWorker.
type linuxMQTTWorker struct {
	*mqttEntities
	logger *slog.Logger
}

// Subscriptions returns the any subscription request messaages for any workers
// with subscriptions.
func (c *linuxMQTTWorker) Subscriptions() []*mqttapi.Subscription {
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

// Configs returns the configuration messages for all workers.
func (c *linuxMQTTWorker) Configs() []*mqttapi.Msg {
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

// Msgs returns the messages channel through which workers pass any messages
// they generate.
func (c *linuxMQTTWorker) Msgs() chan *mqttapi.Msg {
	return c.msgs
}

// generateConfig is a helper function to avoid duplicate code around generating
// an entity subscription.
func (c *linuxMQTTWorker) generateSubscription(e commandEntity) *mqttapi.Subscription {
	sub, err := e.MarshalSubscription()
	if err != nil {
		c.logger.Warn("Could not create subscription.", slog.Any("error", err))

		return nil
	}

	return sub
}

// generateConfig is a helper function to avoid duplicate code around generating
// an entity config.
func (c *linuxMQTTWorker) generateConfig(e stateEntity) *mqttapi.Msg {
	cfg, err := e.MarshalConfig()
	if err != nil {
		c.logger.Warn("Could not create config.", slog.Any("error", err.Error()))

		return nil
	}

	return cfg
}

// setupOSMQTTWorker initializes the list of MQTT workers for sensors and
// returns those that are supported on this device.
func setupOSMQTTWorker(ctx context.Context) MQTTWorker {
	mqttController := &linuxMQTTWorker{
		mqttEntities: &mqttEntities{
			msgs: make(chan *mqttapi.Msg),
		},
	}
	mqttDevice := preferences.MQTTDevice()

	// Add the power controls (suspend, resume, poweroff, etc.).
	powerEntities, err := power.NewPowerControl(ctx, mqttDevice)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not create power controls.",
			slog.Any("error", err))
	} else {
		mqttController.buttons = append(mqttController.buttons, powerEntities...)
	}

	// Add inhibit controls.
	inhibitEntities, err := power.NewInhibitControl(ctx, mqttController.Msgs(), mqttDevice)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not create inhibit control.",
			slog.Any("error", err))
	} else {
		mqttController.switches = append(mqttController.switches, inhibitEntities)
	}

	// Add the screen lock controls.
	screenControls, err := power.NewScreenLockControl(ctx, mqttDevice)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not create screen lock controls.",
			slog.Any("error", err))
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
		logging.FromContext(ctx).Warn("Could not activate MPRIS controller.",
			slog.Any("error", err))
	} else {
		mqttController.sensors = append(mqttController.sensors, mprisEntity)
	}
	// Add camera control.
	cameraEntities, err := media.NewCameraControl(ctx, mqttController.Msgs(), mqttDevice)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not activate Camera controller.",
			slog.Any("error", err))
	} else {
		if cameraEntities != nil {
			mqttController.buttons = append(mqttController.buttons, cameraEntities.StartButton, cameraEntities.StopButton)
			mqttController.cameras = append(mqttController.cameras, cameraEntities.Images)
			mqttController.sensors = append(mqttController.sensors, cameraEntities.Status)
		}
	}

	// Add the D-Bus command action.
	dbusCmdController, err := system.NewDBusCommandSubscription(ctx, mqttDevice)
	if err != nil {
		logging.FromContext(ctx).Warn("Could not activate D-Bus commands controller.",
			slog.Any("error", err))
	} else {
		mqttController.controls = append(mqttController.controls, dbusCmdController)
	}

	go func() {
		defer close(mqttController.msgs)
		<-ctx.Done()
	}()

	return mqttController
}
