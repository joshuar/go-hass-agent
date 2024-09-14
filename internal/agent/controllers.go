// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
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

func (agent *Agent) setupControllers(ctx context.Context, prefs agentPreferences) []any {
	var (
		mqttDevice  *mqtthass.Device
		controllers []any
	)

	// If MQTT functionality is enabled create an MQTT device, used to configure
	// MQTT functionality for some controllers.
	if prefs.IsMQTTEnabled() {
		mqttDevice = prefs.GenerateMQTTDevice(ctx)
		// Create an MQTT commands controller.
		mqttCmdController := newMQTTController(ctx, mqttDevice)
		if mqttCmdController != nil {
			controllers = append(controllers, mqttCmdController)
		}
		// Add the OS MQTT controller.
		controllers = append(controllers, newOSMQTTController(ctx, mqttDevice))
	}

	scriptsController := newScriptsController(ctx)
	if scriptsController != nil {
		controllers = append(controllers, scriptsController)
	}

	// Create a new device controller. The controller will have all the
	// necessary configuration for device-specific sensors.
	devController := agent.newDeviceController(ctx, prefs)
	if devController != nil {
		controllers = append(controllers, devController)
	}
	// Create a new OS controller. The controller will have all the necessary
	// configuration for any OS-specific sensors.
	controllers = append(controllers, newOSSensorController(ctx))

	return controllers
}
