// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package device

import (
	"context"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/linux/media"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
	"github.com/joshuar/go-hass-agent/internal/mqtt"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

const (
	workerID = "linux_mqtt"
)

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

// linuxMQTTWorker represents the Linux-specific OS MQTTWorker.
type linuxMQTTWorker struct {
	msgs          chan mqttapi.Msg
	sensors       []*mqtthass.SensorEntity
	buttons       []*mqtthass.ButtonEntity
	numbers       []*mqtthass.NumberEntity[int]
	switches      []*mqtthass.SwitchEntity
	controls      []*mqttapi.Subscription
	binarySensors []*mqtthass.SensorEntity
	cameras       []*mqtthass.CameraEntity
	logger        *slog.Logger
}

func (c *linuxMQTTWorker) ID() string {
	return workerID
}

func (c *linuxMQTTWorker) IsDisabled() bool {
	return !preferences.MQTTEnabled()
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
	totalLength := len(c.sensors) +
		len(c.binarySensors) +
		len(c.buttons) +
		len(c.switches) +
		len(c.numbers) +
		len(c.cameras)
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
func (c *linuxMQTTWorker) Msgs() chan mqttapi.Msg {
	return c.msgs
}

func (c *linuxMQTTWorker) Start(ctx context.Context) (*mqtt.WorkerData, error) {
	return &mqtt.WorkerData{
		Configs:       c.Configs(),
		Subscriptions: c.Subscriptions(),
		Msgs:          c.Msgs(),
	}, nil
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

// CreateOSMQTTWorkers initializes the list of MQTT workers for sensors and
// returns those that are supported on this device.
//
//nolint:gocyclo,cyclop,funlen
func CreateOSMQTTWorkers(ctx context.Context) workers.MQTTWorker {
	mqttController := &linuxMQTTWorker{}
	// Don't continue if MQTT functionality is disabled.
	if !preferences.MQTTEnabled() {
		return mqttController
	}

	mqttDevice := preferences.MQTTDevice()

	var workerMsgs []<-chan mqttapi.Msg

	// Add the power controls (suspend, resume, poweroff, etc.).
	powerEntities, err := power.NewPowerControl(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not create power controls.",
			slog.Any("error", err))
	} else if powerEntities != nil {
		mqttController.buttons = append(mqttController.buttons, powerEntities...)
	}

	// Add inhibit controls.
	inhibitWorker, err := power.NewInhibitWorker(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not create inhibit control.",
			slog.Any("error", err))
	} else if inhibitWorker != nil {
		mqttController.switches = append(mqttController.switches, inhibitWorker.InhibitControl)
		workerMsgs = append(workerMsgs, inhibitWorker.MsgCh)
	}

	// Add the screen lock controls.
	screenControls, err := power.NewScreenLockControl(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not create screen lock controls.",
			slog.Any("error", err))
	} else if screenControls != nil {
		mqttController.buttons = append(mqttController.buttons, screenControls...)
	}
	// Add the volume controls.
	volumeWorker, err := media.NewVolumeWorker(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could init volume worker.",
			slog.Any("error", err))
	} else if volumeWorker != nil {
		mqttController.numbers = append(mqttController.numbers, volumeWorker.VolumeControl)
		mqttController.switches = append(mqttController.switches, volumeWorker.MuteControl)
		workerMsgs = append(workerMsgs, volumeWorker.MsgCh)
	}
	// Add media control.
	mprisWorker, err := media.NewMPRISWorker(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not activate MPRIS controller.",
			slog.Any("error", err))
	} else if mprisWorker != nil {
		mqttController.sensors = append(mqttController.sensors, mprisWorker.MPRISStatus)
		workerMsgs = append(workerMsgs, mprisWorker.MsgCh)
	}
	// Add camera control.
	cameraWorker, err := media.NewCameraWorker(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not activate Camera controller.",
			slog.Any("error", err))
	} else if cameraWorker != nil {
		mqttController.buttons = append(mqttController.buttons, cameraWorker.StartButton, cameraWorker.StopButton)
		mqttController.cameras = append(mqttController.cameras, cameraWorker.Images)
		mqttController.sensors = append(mqttController.sensors, cameraWorker.Status)
		workerMsgs = append(workerMsgs, cameraWorker.MsgCh)
	}

	// Add the D-Bus command action.
	dbusCmdController, err := system.NewDBusCommandSubscription(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not activate D-Bus commands controller.",
			slog.Any("error", err))
	} else if dbusCmdController != nil {
		mqttController.controls = append(mqttController.controls, dbusCmdController)
	}

	go func() {
		defer close(mqttController.msgs)
		<-ctx.Done()
	}()

	mqttController.msgs = workers.MergeCh(ctx, workerMsgs...)

	return mqttController
}
