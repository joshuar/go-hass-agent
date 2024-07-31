// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver
package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/adrg/xdg"

	"github.com/joshuar/go-hass-agent/internal/agent/ui"
	fyneui "github.com/joshuar/go-hass-agent/internal/agent/ui/fyneUI"
	"github.com/joshuar/go-hass-agent/internal/commands"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrCtxFailed = errors.New("unable to create a context")

// UI are the methods required for the agent to display its windows, tray
// and notifications.
type UI interface {
	DisplayNotification(n ui.Notification)
	DisplayTrayIcon(ctx context.Context, agent ui.Agent, trk ui.SensorTracker)
	DisplayRegistrationWindow(ctx context.Context, prefs *preferences.Preferences, doneCh chan struct{})
	Run(ctx context.Context, agent ui.Agent, doneCh chan struct{})
}

type Registry interface {
	SetDisabled(id string, state bool) error
	SetRegistered(id string, state bool) error
	IsDisabled(id string) bool
	IsRegistered(id string) bool
}

type SensorTracker interface {
	SensorList() []string
	// Process(ctx context.Context, reg sensor.Registry, sensorUpdates ...<-chan sensor.Details) error
	Add(details sensor.Details) error
	Get(key string) (sensor.Details, error)
	Reset()
}

// Agent holds the options of the running agent, the UI object and a channel for
// closing the agent down.
type Agent struct {
	ui            UI
	done          chan struct{}
	prefs         *preferences.Preferences
	logger        *slog.Logger
	id            string
	headless      bool
	forceRegister bool
}

// Option is a functional parameter that will configure a feature of the agent.
type Option func(*Agent)

// newDefaultAgent returns an agent with default options.
func newDefaultAgent(ctx context.Context, id string) *Agent {
	return &Agent{
		done:   make(chan struct{}),
		id:     id,
		logger: logging.FromContext(ctx).With(slog.Group("agent")),
	}
}

// NewAgent creates a new agent with the options specified.
func NewAgent(ctx context.Context, id string, options ...Option) (*Agent, error) {
	agent := newDefaultAgent(ctx, id)

	// If we are using a custom agent ID, adjust the path to the preferences
	// file.
	if agent.id != preferences.AppID {
		preferences.SetPath(filepath.Join(xdg.ConfigHome, agent.id))
	}

	// Load the agent preferences.
	prefs, err := preferences.Load()
	if err != nil && !errors.Is(err, preferences.ErrNoPreferences) {
		return nil, fmt.Errorf("could not create agent: %w", err)
	}

	agent.prefs = prefs

	for _, option := range options {
		option(agent)
	}

	agent.ui = fyneui.NewFyneUI(ctx, agent.id)

	return agent, nil
}

// Headless sets whether the agent should run in a headless mode, without any
// GUI.
func Headless(value bool) Option {
	return func(a *Agent) {
		a.headless = value
	}
}

// WithRegistrationInfo will set the info required for registering the agent.
// Only used when the Register command is run.
func WithRegistrationInfo(server, token string, ignoreURLs bool) Option {
	return func(a *Agent) {
		a.prefs.Registration = &preferences.Registration{
			Server: server,
			Token:  token,
		}
		a.prefs.Hass = &preferences.Hass{
			IgnoreHassURLs: ignoreURLs,
		}
	}
}

// ForceRegister will force the agent to register against Home Assistant,
// regardless of whether it is already registered. Only used when the Register
// command is run.
func ForceRegister(value bool) Option {
	return func(a *Agent) {
		a.forceRegister = value
	}
}

// Run is the "main loop" of the agent. It sets up the agent, loads the config
// then spawns a sensor tracker and the workers to gather sensor data and
// publish it to Home Assistant.
//
//nolint:funlen
//revive:disable:function-length
func (agent *Agent) Run(ctx context.Context, trk SensorTracker, reg Registry) error {
	var wg sync.WaitGroup

	// Pre-flight: check if agent is registered. If not, run registration flow.
	var regWait sync.WaitGroup

	regWait.Add(1)

	go func() {
		defer regWait.Done()

		if err := agent.checkRegistration(ctx, trk); err != nil {
			agent.logger.Log(ctx, logging.LevelFatal, "Error checking registration status.", "error", err.Error())
			close(agent.done)
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		regWait.Wait()

		// Embed some preferences into the context which are used by other parts
		// of the code.
		runnerCtx, cancelFunc := context.WithCancel(ctx)
		runnerCtx = preferences.ContextSetRestAPIURL(runnerCtx, agent.prefs.Hass.RestAPIURL)
		runnerCtx = preferences.ContextSetWebsocketURL(runnerCtx, agent.prefs.Hass.WebsocketURL)
		runnerCtx = preferences.ContextSetWebhookID(runnerCtx, agent.prefs.Hass.WebhookID)
		runnerCtx = preferences.ContextSetToken(runnerCtx, agent.prefs.Registration.Token)

		// Cancel the runner context when the agent is done.
		go func() {
			<-agent.done
			cancelFunc()
			agent.logger.Debug("Agent done.")
		}()

		// Create a new OS controller. The controller will have all the
		// necessary configuration for any OS-specific sensors and MQTT
		// configuration.
		osController := agent.newOSController(runnerCtx)
		// Create a new device controller. The controller will have all the
		// necessary configuration for device-specific sensors and MQTT
		// configuration.
		deviceController := agent.newDeviceController(runnerCtx)

		var sensorCh []<-chan sensor.Details

		// Run workers and get all channels for sensor updates.
		sensorCh = append(sensorCh, agent.runWorkers(runnerCtx, osController, deviceController)...)
		// Run scripts and get channel for sensor updates.
		sensorCh = append(sensorCh, agent.runScripts(runnerCtx))

		wg.Add(1)
		// Process the sensor updates.
		go func() {
			defer wg.Done()
			agent.processSensors(runnerCtx, trk, reg, sensorCh...)
		}()

		// Start the mqtt client if MQTT is enabled.
		if agent.prefs.GetMQTTPreferences().IsMQTTEnabled() {
			wg.Add(1)

			var commandController *commands.Controller

			commandsFile := filepath.Join(xdg.ConfigHome, agent.id, "commands.toml")

			mqttDeviceInfo, err := device.MQTTDevice(preferences.AppName, agent.id, preferences.AppURL, preferences.AppVersion)
			if err != nil {
				agent.logger.Warn("Could not set up MQTT commands controller.", "error", err.Error())
			} else {
				// Create an MQTT device for this operating system and run its Setup.
				commandController, err = commands.NewCommandsController(ctx, commandsFile, mqttDeviceInfo)
				if err != nil {
					agent.logger.Warn("Could not set up MQTT commands controller.", "error", err.Error())

					return
				}
			}

			go func() {
				defer wg.Done()

				agent.runMQTTWorker(runnerCtx, osController, commandController)
			}()
		}

		wg.Add(1)
		// Listen for notifications from Home Assistant.
		go func() {
			defer wg.Done()
			agent.runNotificationsWorker(runnerCtx)
		}()
	}()

	agent.handleSignals()

	agent.ui.DisplayTrayIcon(ctx, agent, trk)
	agent.ui.Run(ctx, agent, agent.done)

	wg.Wait()

	return nil
}

func (agent *Agent) Register(ctx context.Context, trk SensorTracker) {
	defer agent.Stop()

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := agent.checkRegistration(ctx, trk); err != nil {
			agent.logger.Log(ctx, logging.LevelFatal, "Error checking registration status", "error", err.Error())
		}
	}()

	agent.ui.Run(ctx, agent, agent.done)
	wg.Wait()
}

// handleSignals will handle Ctrl-C of the agent.
func (agent *Agent) handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer close(agent.done)
		<-c
		agent.logger.Debug("Ctrl-C pressed.")
	}()
}

// Stop will close the agent's done channel which indicates to any goroutines it
// is time to clean up and exit.
func (agent *Agent) Stop() {
	defer close(agent.done)

	agent.logger.Debug("Stopping Agent.")

	if err := agent.prefs.Save(); err != nil {
		agent.logger.Warn("Could not save agent preferences", "error", err.Error())
	}
}

// Reset will remove any agent related files and configuration.
func (agent *Agent) Reset(ctx context.Context) error {
	osController := agent.newOSController(ctx)

	if err := agent.resetMQTTWorker(ctx, osController); err != nil {
		return fmt.Errorf("problem resetting agent: %w", err)
	}

	return nil
}

// Headless returns a boolean indicating whether the agent is running in
// headless mode or not.
func (agent *Agent) Headless() bool {
	return agent.headless
}

// GetMQTTPreferences returns the subset of agent preferences to do with MQTT.
func (agent *Agent) GetMQTTPreferences() *preferences.MQTT {
	return agent.prefs.GetMQTTPreferences()
}

// SaveMQTTPreferences takes the given preferences and saves them to disk as
// part of all agent preferences.
func (agent *Agent) SaveMQTTPreferences(prefs *preferences.MQTT) error {
	agent.prefs.MQTT = prefs

	err := agent.prefs.Save()
	if err != nil {
		return fmt.Errorf("failed to save mqtt preferences: %w", err)
	}

	return nil
}

func (agent *Agent) GetRestAPIURL() string {
	return agent.prefs.Hass.RestAPIURL
}
