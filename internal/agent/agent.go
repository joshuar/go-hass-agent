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
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/adrg/xdg"

	fyneui "github.com/joshuar/go-hass-agent/internal/agent/ui/fyneUI"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrCtxFailed = errors.New("unable to create a context")

// Agent holds the options of the running agent, the UI object and a channel for
// closing the agent down.
type Agent struct {
	ui               UI
	done             chan struct{}
	registrationInfo *hass.RegistrationInput
	prefs            *preferences.Preferences
	id               string
	headless         bool
	forceRegister    bool
}

// Option is a functional parameter that will configure a feature of the agent.
type Option func(*Agent)

// newDefaultAgent returns an agent with default options.
//
//nolint:exhaustruct
func newDefaultAgent() *Agent {
	return &Agent{
		done:             make(chan struct{}),
		id:               preferences.AppID,
		registrationInfo: &hass.RegistrationInput{},
	}
}

// NewAgent creates a new agent with the options specified.
func NewAgent(ctx context.Context, options ...Option) (*Agent, error) {
	agent := newDefaultAgent()

	for _, option := range options {
		option(agent)
	}

	prefs, err := preferences.Load(agent.id)
	if err != nil && !errors.Is(err, preferences.ErrNoPreferences) {
		return nil, fmt.Errorf("could not create agent: %w", err)
	}

	agent.prefs = prefs

	if !agent.headless {
		agent.ui = fyneui.NewFyneUI(ctx, agent.id)
	}

	return agent, nil
}

// WithID will set the agent ID to the value given.
func WithID(id string) Option {
	return func(a *Agent) {
		a.id = id
	}
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
		a.registrationInfo = &hass.RegistrationInput{
			Server:           server,
			Token:            token,
			IgnoreOutputURLs: ignoreURLs,
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
//revive:disable:function-length
func (agent *Agent) Run(ctx context.Context, trk SensorTracker, reg sensor.Registry) error {
	var wg sync.WaitGroup

	// Pre-flight: check if agent is registered. If not, run registration flow.
	var regWait sync.WaitGroup

	regWait.Add(1)

	go func() {
		defer regWait.Done()

		if err := agent.checkRegistration(ctx, trk); err != nil {
			logging.FromContext(ctx).Log(ctx, logging.LevelFatal, "Error checking registration status.", "error", err.Error())
		}
	}()

	wg.Add(1)

	go func() {
		var err error

		defer wg.Done()
		regWait.Wait()

		// Embed the agent preferences in the context.
		ctx = preferences.ContextSetPrefs(ctx, agent.prefs)

		// Embed required settings for Home Assistant in the context.
		ctx, err = hass.SetupContext(ctx)
		if err != nil {
			logging.FromContext(ctx).Error("Could not add hass details to context.", "error", err.Error())

			return
		}

		runnerCtx, cancelFunc := context.WithCancel(ctx)

		// Create a new OS controller. The controller will have all the
		// necessary configuration for any OS-specific sensors and MQTT
		// configuration.
		osController := newOSController(runnerCtx)

		go func() {
			<-agent.done
			cancelFunc()
			logging.FromContext(ctx).Debug("Agent done.")
		}()

		// Start worker funcs for sensors.
		wg.Add(1)

		go func() {
			defer wg.Done()
			runWorkers(runnerCtx, osController, trk, reg)
		}()
		// Start any scripts.
		wg.Add(1)

		go func() {
			defer wg.Done()

			scriptPath := filepath.Join(xdg.ConfigHome, agent.AppID(), "scripts")
			runScripts(runnerCtx, scriptPath, trk, reg)
		}()
		// Start the mqtt client if MQTT is enabled.
		if agent.prefs.MQTTEnabled {
			wg.Add(1)

			go func() {
				defer wg.Done()

				commandsFile := filepath.Join(xdg.ConfigHome, agent.AppID(), "commands.toml")
				agent.runMQTTWorker(runnerCtx, osController, commandsFile)
			}()
		}
		// Listen for notifications from Home Assistant.
		if !agent.headless {
			wg.Add(1)

			go func() {
				defer wg.Done()
				agent.runNotificationsWorker(runnerCtx)
			}()
		}
	}()

	agent.handleSignals(ctx)

	if !agent.headless {
		agent.ui.DisplayTrayIcon(ctx, agent, trk)
		agent.ui.Run(ctx, agent.done)
	}

	wg.Wait()

	return nil
}

func (agent *Agent) Register(ctx context.Context, trk SensorTracker) {
	go func() {
		if err := agent.checkRegistration(ctx, trk); err != nil {
			logging.FromContext(ctx).Log(ctx, logging.LevelFatal, "Error checking registration status", "error", err.Error())
		}

		agent.Stop(ctx)
	}()

	if !agent.headless {
		agent.ui.Run(ctx, agent.done)
	}
}

// handleSignals will handle Ctrl-C of the agent.
func (agent *Agent) handleSignals(ctx context.Context) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer close(agent.done)
		<-c
		logging.FromContext(ctx).Debug("Ctrl-C pressed.")
	}()
}

// AppID returns the "application ID". Currently, this ID is just used to
// indicate whether the agent is running in debug mode or not.
func (agent *Agent) AppID() string {
	return agent.id
}

// Stop will close the agent's done channel which indicates to any goroutines it
// is time to clean up and exit.
func (agent *Agent) Stop(ctx context.Context) {
	defer close(agent.done)

	logging.FromContext(ctx).Debug("Stopping Agent.")

	if err := agent.prefs.Save(); err != nil {
		logging.FromContext(ctx).Warn("Could not save agent preferences", "error", err.Error())
	}
}

func (agent *Agent) Reset(ctx context.Context) error {
	// Embed the agent preferences in the context.
	ctx = preferences.ContextSetPrefs(ctx, agent.prefs)

	osController := newOSController(ctx)

	if err := agent.resetMQTTWorker(ctx, osController); err != nil {
		return fmt.Errorf("problem resetting agent: %w", err)
	}

	return nil
}
