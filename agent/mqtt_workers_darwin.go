package agent

import (
	"context"
	"fmt"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v12/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v12/pkg/mqtt"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/platform/darwin/media"
	"github.com/joshuar/go-hass-agent/platform/darwin/power"
)

// stateEntity is satisfied by any entity whose config can be marshalled.
type stateEntity interface {
	MarshalConfig() (*mqttapi.Msg, error)
}

// commandEntity is satisfied by any entity that also has a command
// subscription.
type commandEntity interface {
	stateEntity
	MarshalSubscription() (*mqttapi.Subscription, error)
}

// darwinMQTTWorker represents the darwin-specific (macOS) MQTTWorker.
type darwinMQTTWorker struct {
	msgs     chan mqttapi.Msg
	buttons  []*mqtthass.ButtonEntity
	switches []*mqtthass.SwitchEntity
	numbers  []*mqtthass.NumberEntity[int]
	logger   *slog.Logger
}

func (c *darwinMQTTWorker) IsDisabled() bool {
	return false
}

// Subscriptions returns the subscription requests for all controls.
func (c *darwinMQTTWorker) Subscriptions() []*mqttapi.Subscription {
	total := len(c.buttons) + len(c.switches) + len(c.numbers)
	subs := make([]*mqttapi.Subscription, 0, total)

	for _, button := range c.buttons {
		subs = append(subs, c.generateSubscription(button))
	}
	for _, sw := range c.switches {
		subs = append(subs, c.generateSubscription(sw))
	}
	for _, number := range c.numbers {
		subs = append(subs, c.generateSubscription(number))
	}

	return subs
}

// Configs returns the configuration messages for all controls.
func (c *darwinMQTTWorker) Configs() []*mqttapi.Msg {
	total := len(c.buttons) + len(c.switches) + len(c.numbers)
	configs := make([]*mqttapi.Msg, 0, total)

	for _, button := range c.buttons {
		configs = append(configs, c.generateConfig(button))
	}
	for _, sw := range c.switches {
		configs = append(configs, c.generateConfig(sw))
	}
	for _, number := range c.numbers {
		configs = append(configs, c.generateConfig(number))
	}

	return configs
}

// Msgs returns the channel through which controls publish state updates.
func (c *darwinMQTTWorker) Msgs() chan mqttapi.Msg {
	return c.msgs
}

func (c *darwinMQTTWorker) Start(_ context.Context) (*mqtt.WorkerData, error) {
	return &mqtt.WorkerData{
		Configs:       c.Configs(),
		Subscriptions: c.Subscriptions(),
		Msgs:          c.Msgs(),
	}, nil
}

func (c *darwinMQTTWorker) generateSubscription(e commandEntity) *mqttapi.Subscription {
	sub, err := e.MarshalSubscription()
	if err != nil {
		c.logger.Warn("Could not create subscription.", slog.Any("error", err))
		return nil
	}

	return sub
}

func (c *darwinMQTTWorker) generateConfig(e stateEntity) *mqttapi.Msg {
	cfg, err := e.MarshalConfig()
	if err != nil {
		c.logger.Warn("Could not create config.", slog.Any("error", err))
		return nil
	}

	return cfg
}

// CreateOSMQTTWorkers initializes the list of MQTT workers for controls and
// returns those that are supported on this device.
func CreateOSMQTTWorkers(ctx context.Context) (workers.MQTTWorker, error) {
	mqttController := &darwinMQTTWorker{
		logger: slogctx.FromCtx(ctx).WithGroup("mqtt"),
	}

	mqttDevice, err := mqtt.Device()
	if err != nil {
		return nil, fmt.Errorf("create mqtt device: %w", err)
	}

	var workerMsgs []<-chan mqttapi.Msg

	// Add the power controls (sleep, restart, shut down, lock screen).
	powerEntities, err := power.NewPowerControl(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not create MQTT power controls.",
			slog.Any("error", err))
	} else if powerEntities != nil {
		mqttController.buttons = append(mqttController.buttons, powerEntities...)
	}

	// Add the inhibit sleep control.
	inhibitWorker, err := power.NewInhibitWorker(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not create MQTT inhibit control.",
			slog.Any("error", err))
	} else if inhibitWorker != nil {
		mqttController.switches = append(mqttController.switches, inhibitWorker.InhibitControl)
		workerMsgs = append(workerMsgs, inhibitWorker.MsgCh)
	}

	// Add the volume/mute controls.
	volumeWorker, err := media.NewVolumeWorker(ctx, mqttDevice)
	if err != nil {
		slogctx.FromCtx(ctx).Warn("Could not create MQTT volume controls.",
			slog.Any("error", err))
	} else if volumeWorker != nil {
		mqttController.numbers = append(mqttController.numbers, volumeWorker.VolumeControl)
		mqttController.switches = append(mqttController.switches, volumeWorker.MuteControl)
		workerMsgs = append(workerMsgs, volumeWorker.MsgCh)
	}

	mqttController.msgs = workers.MergeCh(ctx, workerMsgs...)

	return mqttController, nil
}
