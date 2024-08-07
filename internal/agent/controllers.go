// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// SensorController represents an object that manages one or more Workers.
type SensorController interface {
	// ActiveWorkers is a list of the names of all currently active Workers.
	ActiveWorkers() []string
	// InactiveWorkers is a list of the names of all currently inactive Workers.
	InactiveWorkers() []string
	// Start provides a way to start the named Worker.
	Start(ctx context.Context, name string) (<-chan sensor.Details, error)
	// Stop provides a way to stop the named Worker.
	Stop(name string) error
	// StartAll will start all Workers that this controller manages.
	StartAll(ctx context.Context) (<-chan sensor.Details, error)
	// StopAll will stop all Workers that this controller manages.
	StopAll() error
}

// Worker represents an object that is responsible for controlling the
// publishing of one or more sensors.
type Worker interface {
	ID() string
	// Sensors returns an array of the current value of all sensors, or a
	// non-nil error if this is not possible.
	Sensors(ctx context.Context) ([]sensor.Details, error)
	// Updates returns a channel on which updates to sensors will be published,
	// when they become available.
	Updates(ctx context.Context) (<-chan sensor.Details, error)
	// Stop is used to tell the worker to stop any background updates of
	// sensors.
	Stop() error
}

// MQTTController represents an object that is responsible for controlling the
// publishing of one or more commands over MQTT.
type MQTTController interface {
	// Subscriptions is a list of MQTT subscriptions this object wants to
	// establish on the MQTT broker.
	Subscriptions() []*mqttapi.Subscription
	// Configs are MQTT messages sent to the broker that Home Assistant will use
	// to set up entities.
	Configs() []*mqttapi.Msg
	// Msgs returns a channel on which this object will send MQTT messages on
	// certain events.
	Msgs() chan *mqttapi.Msg
}

func (agent *Agent) setupControllers(ctx context.Context) []any {
	var (
		mqttDevice  *mqtthass.Device
		err         error
		controllers []any
	)

	// If MQTT functionality is enabled create an MQTT device, used to configure
	// MQTT functionality for some controllers.
	if agent.prefs.MQTT.IsMQTTEnabled() {
		mqttDevice, err = device.MQTTDevice(preferences.AppName, preferences.AppID, preferences.AppURL, preferences.AppVersion)
		if err != nil {
			agent.logger.Warn("Could not create MQTT device, MQTT functionality will not be available.", slog.Any("error", err))
		}
	}

	scriptsController := agent.newScriptsController(ctx)
	if scriptsController != nil {
		controllers = append(controllers, scriptsController)
	}

	// Create a new device controller. The controller will have all the
	// necessary configuration for device-specific sensors and MQTT
	// configuration.
	devController := agent.newDeviceController(ctx)
	if devController != nil {
		controllers = append(controllers, devController)
	}
	// Create a new OS controller. The controller will have all the
	// necessary configuration for any OS-specific sensors and MQTT
	// configuration.
	osSensorController, osMQTTController := agent.newOSController(ctx, mqttDevice)
	controllers = append(controllers, osSensorController, osMQTTController)

	// Create an MQTT commands controller.
	mqttCmdController := agent.newMQTTController(ctx, mqttDevice)
	if mqttCmdController != nil {
		controllers = append(controllers, mqttCmdController)
	}

	return controllers
}

// runSensorWorkers will start all the sensor worker functions for all sensor
// controllers passed in. It returns a single merged channel of sensor updates.
func (agent *Agent) runSensorWorkers(ctx context.Context, controllers ...SensorController) []<-chan sensor.Details {
	var sensorCh []<-chan sensor.Details

	for _, controller := range controllers {
		ch, err := controller.StartAll(ctx)
		if err != nil {
			agent.logger.Warn("Start controller had errors.", "errors", err.Error())
		} else {
			sensorCh = append(sensorCh, ch)
		}
	}

	if len(sensorCh) == 0 {
		agent.logger.Warn("No workers were started by any controllers.")

		return sensorCh
	}

	go func() {
		<-ctx.Done()

		for _, controller := range controllers {
			if err := controller.StopAll(); err != nil {
				agent.logger.Warn("Stop controller had errors.", "error", err.Error())
			}
		}
	}()

	return sensorCh
}
