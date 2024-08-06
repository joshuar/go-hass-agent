// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/linux/apps"
	"github.com/joshuar/go-hass-agent/internal/linux/battery"
	"github.com/joshuar/go-hass-agent/internal/linux/cpu"
	"github.com/joshuar/go-hass-agent/internal/linux/desktop"
	"github.com/joshuar/go-hass-agent/internal/linux/disk"
	"github.com/joshuar/go-hass-agent/internal/linux/location"
	"github.com/joshuar/go-hass-agent/internal/linux/media"
	"github.com/joshuar/go-hass-agent/internal/linux/mem"
	"github.com/joshuar/go-hass-agent/internal/linux/net"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/problems"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
	"github.com/joshuar/go-hass-agent/internal/linux/user"
	"github.com/joshuar/go-hass-agent/pkg/linux/dbusx"
)

// allworkers is the list of sensor allworkers supported on Linux.
var allworkers = []func(context.Context, *dbusx.DBusAPI) (*linux.SensorWorker, error){
	apps.NewAppWorker,
	battery.NewBatteryWorker,
	cpu.NewCPUFreqWorker,
	cpu.NewLoadAvgWorker,
	cpu.NewUsageWorker,
	desktop.NewDesktopWorker,
	disk.NewIOWorker,
	disk.NewUsageWorker,
	location.NewLocationWorker,
	mem.NewUsageWorker,
	net.NewConnectionWorker,
	net.NewRatesWorker,
	power.NewLaptopWorker,
	power.NewProfileWorker,
	power.NewStateWorker,
	power.NewScreenLockWorker,
	problems.NewProblemsWorker,
	// power.IdleUpdater,
	system.NewHWMonWorker,
	system.NewInfoWorker,
	system.NewTimeWorker,
	user.NewUserWorker,
}

var (
	ErrWorkerAlreadyStarted = errors.New("worker already started")
	ErrUnknownWorker        = errors.New("unknown worker")
)

type mqttWorker struct {
	msgs     chan *mqttapi.Msg
	sensors  []*mqtthass.SensorEntity
	buttons  []*mqtthass.ButtonEntity
	numbers  []*mqtthass.NumberEntity[int]
	switches []*mqtthass.SwitchEntity
	controls []*mqttapi.Subscription
}

type linuxSensorController struct {
	deviceController
}

type linuxMQTTController struct {
	*mqttWorker
	logger *slog.Logger
}

func (w *linuxMQTTController) Subscriptions() []*mqttapi.Subscription {
	var subs []*mqttapi.Subscription

	// Create subscriptions for buttons.
	for _, button := range w.buttons {
		if sub, err := button.MarshalSubscription(); err != nil {
			w.logger.Warn("Could not create subscription.", "entity", button.Name, "error", err.Error())
		} else {
			subs = append(subs, sub)
		}
	}
	// Create subscriptions for numbers.
	for _, number := range w.numbers {
		if sub, err := number.MarshalSubscription(); err != nil {
			w.logger.Warn("Could not create subscription.", "entity", number.Name, "error", err.Error())
		} else {
			subs = append(subs, sub)
		}
	}
	// Create subscriptions for switches.
	for _, sw := range w.switches {
		if sub, err := sw.MarshalSubscription(); err != nil {
			w.logger.Warn("Could not create subscription.", "entity", sw.Name, "error", err.Error())
		} else {
			subs = append(subs, sub)
		}
	}
	// Add subscriptions for any additional controls.
	subs = append(subs, w.controls...)

	return subs
}

func (w *linuxMQTTController) Configs() []*mqttapi.Msg {
	var configs []*mqttapi.Msg

	// Create sensor configs.
	for _, sensorEntity := range w.sensors {
		if sub, err := sensorEntity.MarshalConfig(); err != nil {
			w.logger.Warn("Could not create config.", "entity", sensorEntity.Name, "error", err.Error())
		} else {
			configs = append(configs, sub)
		}
	}
	// Create button configs.
	for _, buttonEntity := range w.buttons {
		if sub, err := buttonEntity.MarshalConfig(); err != nil {
			w.logger.Warn("Could not create config.", "entity", buttonEntity.Name, "error", err.Error())
		} else {
			configs = append(configs, sub)
		}
	}
	// Create number configs.
	for _, numberEntity := range w.numbers {
		if sub, err := numberEntity.MarshalConfig(); err != nil {
			w.logger.Warn("Could not create config.", "entity", numberEntity.Name, "error", err.Error())
		} else {
			configs = append(configs, sub)
		}
	}
	// Create switch configs.
	for _, switchEntity := range w.switches {
		if sub, err := switchEntity.MarshalConfig(); err != nil {
			w.logger.Warn("Could not create config.", "entity", switchEntity.Name, "error", err.Error())
		} else {
			configs = append(configs, sub)
		}
	}

	return configs
}

func (w *linuxMQTTController) Msgs() chan *mqttapi.Msg {
	return w.msgs
}

// newOSController initialises the list of workers for sensors and returns those
// that are supported on this device.
func (agent *Agent) newOSController(ctx context.Context, mqttDevice *mqtthass.Device) (SensorController, MQTTController) {
	dbusAPI := dbusx.NewDBusAPI(ctx, agent.logger.With(slog.Group("dbus")))

	sensorController := &linuxSensorController{
		deviceController: deviceController{
			sensorWorkers: make(map[string]*sensorWorker),
			logger:        agent.logger.With(slog.Group("linux", slog.String("controller", "sensor"))),
		},
	}

	// Set up sensor workers.
	for _, startWorkerFunc := range allworkers {
		worker, err := startWorkerFunc(ctx, dbusAPI)
		if err != nil {
			sensorController.logger.Warn("Could not start a sensor worker.", "error", err.Error())

			continue
		}

		sensorController.sensorWorkers[worker.ID()] = &sensorWorker{object: worker, started: false}
	}

	// Stop setup if there is no mqttDevice.
	if mqttDevice == nil {
		return sensorController, nil
	}

	mqttController := &linuxMQTTController{
		mqttWorker: &mqttWorker{
			msgs: make(chan *mqttapi.Msg),
		},
		logger: agent.logger.With(slog.Group("linux", slog.String("controller", "mqtt"))),
	}

	// Add the power controls (suspend, resume, poweroff, etc.).
	powerButtons := power.NewPowerControl(ctx, dbusAPI, mqttController.logger, mqttDevice)
	if powerButtons != nil {
		mqttController.buttons = append(mqttController.buttons, powerButtons...)
	}
	// Add the screen lock controls.
	screenLock := power.NewScreenLockControl(ctx, dbusAPI, mqttController.logger, mqttDevice)
	if screenLock != nil {
		mqttController.buttons = append(mqttController.buttons, screenLock)
	}
	// Add the volume controls.
	volEntity, muteEntity := media.VolumeControl(ctx, mqttController.Msgs(), mqttController.logger, mqttDevice)
	if volEntity != nil && muteEntity != nil {
		mqttController.numbers = append(mqttController.numbers, volEntity)
		mqttController.switches = append(mqttController.switches, muteEntity)
	}
	// Add media control.
	mprisEntity, err := media.MPRISControl(ctx, dbusAPI, mqttController.logger, mqttDevice, mqttController.Msgs())
	if err != nil {
		mqttController.logger.Warn("could not activate MPRIS controller", slog.Any("error", err))
	} else {
		mqttController.sensors = append(mqttController.sensors, mprisEntity)
	}

	// Add the D-Bus command action.
	mqttController.controls = append(mqttController.controls, system.NewDBusCommandSubscription(ctx, dbusAPI, mqttController.logger))

	go func() {
		defer close(mqttController.msgs)
		<-ctx.Done()
	}()

	return sensorController, mqttController
}
