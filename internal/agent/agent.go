// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver

package agent

import (
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"

	fyneui "github.com/joshuar/go-hass-agent/internal/agent/ui/fyneUI"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrCtxFailed = errors.New("unable to create a context")

// Agent holds the options of the running agent, the UI object and a channel for
// closing the agent down.
type Agent struct {
	ui               UI
	done             chan struct{}
	registrationInfo *hass.RegistrationInput
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
		done: make(chan struct{}),
		id:   preferences.AppID,
	}
}

// NewAgent creates a new agent with the options specified.
func NewAgent(options ...Option) *Agent {
	agent := newDefaultAgent()

	for _, option := range options {
		option(agent)
	}

	if !agent.headless {
		agent.ui = fyneui.NewFyneUI(agent.id)
	}

	return agent
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
func (agent *Agent) Run(trk SensorTracker, reg sensor.Registry) {
	var wg sync.WaitGroup

	// Pre-flight: check if agent is registered. If not, run registration flow.
	var regWait sync.WaitGroup

	regWait.Add(1)

	go func() {
		defer regWait.Done()

		if err := agent.checkRegistration(trk); err != nil {
			log.Fatal().Err(err).Msg("Error checking registration status.")
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		regWait.Wait()

		ctx, cancelFunc := hass.NewContext()
		if ctx == nil {
			log.Error().Msg("Unable to create context.")

			return
		}

		runnerCtx := setupDeviceContext(ctx)

		go func() {
			<-agent.done
			log.Debug().Msg("Agent done.")
			cancelFunc()
		}()

		// Start worker funcs for sensors.
		wg.Add(1)

		go func() {
			defer wg.Done()
			runWorkers(runnerCtx, trk, reg)
		}()
		// Start any scripts.
		wg.Add(1)

		go func() {
			defer wg.Done()

			scriptPath := filepath.Join(xdg.ConfigHome, agent.AppID(), "scripts")
			runScripts(runnerCtx, scriptPath, trk, reg)
		}()
		// Start the mqtt client
		wg.Add(1)

		go func() {
			defer wg.Done()

			commandsFile := filepath.Join(xdg.ConfigHome, agent.AppID(), "commands.toml")
			runMQTTWorker(runnerCtx, commandsFile)
		}()
		// Listen for notifications from Home Assistant.
		if !agent.headless {
			wg.Add(1)

			go func() {
				defer wg.Done()
				agent.runNotificationsWorker(runnerCtx)
			}()
		}
	}()

	agent.handleSignals()

	if !agent.headless {
		agent.ui.DisplayTrayIcon(agent, trk)
		agent.ui.Run(agent.done)
	}

	wg.Wait()
}

func (agent *Agent) Register(trk SensorTracker) {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := agent.checkRegistration(trk); err != nil {
			log.Fatal().Err(err).Msg("Error checking registration status.")
		}
	}()

	if !agent.headless {
		agent.ui.Run(agent.done)
	}

	wg.Wait()
	agent.Stop()
}

// handleSignals will handle Ctrl-C of the agent.
func (agent *Agent) handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer close(agent.done)
		<-c
		log.Debug().Msg("Ctrl-C pressed.")
	}()
}

// AppID returns the "application ID". Currently, this ID is just used to
// indicate whether the agent is running in debug mode or not.
func (agent *Agent) AppID() string {
	return agent.id
}

// Stop will close the agent's done channel which indicates to any goroutines it
// is time to clean up and exit.
func (agent *Agent) Stop() {
	log.Debug().Msg("Stopping agent.")
	close(agent.done)
}

func (agent *Agent) Reset() error {
	ctx, _ := hass.NewContext()
	if ctx == nil {
		return ErrCtxFailed
	}

	runnerCtx := setupDeviceContext(ctx)

	resetMQTTWorker(runnerCtx)

	return nil
}
