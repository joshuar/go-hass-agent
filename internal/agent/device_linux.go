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

	"github.com/joshuar/go-hass-agent/internal/device"
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
	"github.com/joshuar/go-hass-agent/internal/preferences"
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

type linuxController struct {
	deviceController
	dbusAPI    *dbusx.DBusAPI
	mqttDevice *mqtthass.Device
	*mqttWorker
}

// linuxController implements MQTTController

func (w *linuxController) Subscriptions() []*mqttapi.Subscription {
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

func (w *linuxController) Configs() []*mqttapi.Msg {
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

func (w *linuxController) Msgs() chan *mqttapi.Msg {
	return w.msgs
}

// newOSController initialises the list of workers for sensors and returns those
// that are supported on this device.
//
//nolint:exhaustruct
func (agent *Agent) newOSController(ctx context.Context) Controller {
	mqttDevice, err := device.MQTTDevice(preferences.AppName, preferences.AppID, preferences.AppURL, preferences.AppVersion)
	if err != nil {
		agent.logger.Error("Unable to create MQTT device controller.", "error", err.Error())

		return nil
	}

	controller := &linuxController{
		deviceController: deviceController{
			sensorWorkers: make(map[string]*sensorWorker),
			logger:        agent.logger.With(slog.Group("linux")),
		},
		dbusAPI: dbusx.NewDBusAPI(ctx, agent.logger.With(slog.Group("dbus"))),
		mqttWorker: &mqttWorker{
			msgs: make(chan *mqttapi.Msg),
		},
		mqttDevice: mqttDevice,
	}

	// Set up sensor workers.
	for _, startWorkerFunc := range allworkers {
		worker, err := startWorkerFunc(ctx, controller.dbusAPI)
		if err != nil {
			controller.logger.Warn("Could not start a sensor worker.", "error", err.Error())

			continue
		}

		controller.sensorWorkers[worker.ID()] = &sensorWorker{object: worker, started: false}
	}

	// Only set up MQTT if MQTT is enabled.
	if !agent.prefs.MQTT.IsMQTTEnabled() {
		return controller
	}

	// Add the power controls (suspend, resume, poweroff, etc.).
	powerButtons := power.NewPowerControl(ctx, controller.dbusAPI, controller.logger, controller.mqttDevice)
	if powerButtons != nil {
		controller.buttons = append(controller.buttons, powerButtons...)
	}
	// Add the screen lock controls.
	screenLock := power.NewScreenLockControl(ctx, controller.dbusAPI, controller.logger, controller.mqttDevice)
	if screenLock != nil {
		controller.buttons = append(controller.buttons, screenLock)
	}
	// Add the volume controls.
	volEntity, muteEntity := media.VolumeControl(ctx, controller.Msgs(), controller.logger, controller.mqttDevice)
	if volEntity != nil && muteEntity != nil {
		controller.numbers = append(controller.numbers, volEntity)
		controller.switches = append(controller.switches, muteEntity)
	}
	// Add the D-Bus command action.
	controller.controls = append(controller.controls, system.NewDBusCommandSubscription(ctx, controller.dbusAPI, controller.logger))

	go func() {
		defer close(controller.msgs)
		<-ctx.Done()
	}()

	return controller
}
