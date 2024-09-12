// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver
//
//go:generate moq -out agent_mocks_test.go . UI SensorController MQTTController Worker
package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	fyneui "github.com/joshuar/go-hass-agent/internal/agent/ui/fyneUI"

	"github.com/joshuar/go-hass-agent/internal/agent/ui"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	defaultTimeout = 30 * time.Second
)

var ErrInvalidPrefernces = errors.New("invalid agent preferences")

// UI are the methods required for the agent to display its windows, tray
// and notifications.
type UI interface {
	DisplayNotification(n ui.Notification)
	DisplayTrayIcon(ctx context.Context, agent ui.Agent, client ui.HassClient, doneCh chan struct{})
	DisplayRegistrationWindow(prefs *preferences.Preferences, doneCh chan struct{}) chan struct{}
	Run(agent ui.Agent, doneCh chan struct{})
}

// Agent holds the options of the running agent, the UI object and a channel for
// closing the agent down.
type Agent struct {
	ui            UI
	hass          HassClient
	done          chan struct{}
	prefs         *preferences.Preferences
	logger        *slog.Logger
	headless      bool
	forceRegister bool
}

// Option is a functional parameter that will configure a feature of the agent.
type Option func(*Agent)

// newDefaultAgent returns an agent with default options.
func newDefaultAgent(ctx context.Context) *Agent {
	return &Agent{
		done:   make(chan struct{}),
		logger: logging.FromContext(ctx),
	}
}

// NewAgent creates a new agent with the options specified.
func NewAgent(ctx context.Context, options ...Option) (*Agent, error) {
	agent := newDefaultAgent(ctx)

	// Load the agent preferences.
	prefs, err := preferences.Load(ctx)
	if err != nil && !errors.Is(err, preferences.ErrNoPreferences) {
		return nil, fmt.Errorf("could not create agent: %w", err)
	}

	agent.prefs = prefs

	for _, option := range options {
		option(agent)
	}

	agent.ui = fyneui.NewFyneUI(ctx)

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
//revive:disable:function-length
func (agent *Agent) Run(ctx context.Context) error {
	var (
		wg      sync.WaitGroup
		regWait sync.WaitGroup
		err     error
	)

	ctx, cancelFunc := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancelFunc()
	}()

	agent.hass, err = hass.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("could not start hass client: %w", err)
	}

	regWait.Add(1)

	go func() {
		defer regWait.Done()

		if err := agent.checkRegistration(ctx); err != nil {
			agent.logger.Log(ctx, logging.LevelFatal, "Error checking registration status.", slog.Any("error", err))
			close(agent.done)
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		regWait.Wait()

		agent.hass.Endpoint(agent.prefs.RestAPIURL(), defaultTimeout)

		// Create a context for runners
		controllerCtx, cancelFunc := context.WithCancel(ctx)

		// Cancel the runner context when the agent is done.
		go func() {
			<-agent.done
			cancelFunc()
			agent.logger.Debug("Agent done.")
		}()

		var (
			sensorControllers []SensorController
			mqttControllers   []MQTTController
		)
		// Setup and sort all controllers by type.
		for _, c := range agent.setupControllers(controllerCtx) {
			switch controller := c.(type) {
			case SensorController:
				sensorControllers = append(sensorControllers, controller)
			case MQTTController:
				mqttControllers = append(mqttControllers, controller)
			}
		}

		wg.Add(1)
		// Run workers for any sensor controllers.
		go func() {
			defer wg.Done()
			agent.runSensorWorkers(controllerCtx, sensorControllers...)
		}()

		if len(mqttControllers) > 0 {
			wg.Add(1)
			// Run workers for any MQTT controllers.
			go func() {
				defer wg.Done()
				agent.runMQTTWorkers(controllerCtx, mqttControllers...)
			}()
		}

		wg.Add(1)
		// Listen for notifications from Home Assistant.
		go func() {
			defer wg.Done()
			agent.runNotificationsWorker(controllerCtx)
		}()
	}()

	agent.ui.DisplayTrayIcon(ctx, agent, agent.hass, agent.done)
	agent.ui.Run(agent, agent.done)

	wg.Wait()

	return nil
}

func (agent *Agent) Register(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := agent.checkRegistration(ctx); err != nil {
			agent.logger.Log(ctx, logging.LevelFatal, "Error checking registration status", slog.Any("error", err))
		}

		close(agent.done)
	}()

	agent.ui.Run(agent, agent.done)
	wg.Wait()
}

// Stop will close the agent's done channel which indicates to any goroutines it
// is time to clean up and exit.
func (agent *Agent) Stop() {
	defer close(agent.done)

	agent.logger.Debug("Stopping Agent.")

	if err := agent.prefs.Save(); err != nil {
		agent.logger.Warn("Could not save agent preferences", slog.Any("error", err))
	}
}

// Reset will remove any agent related files and configuration.
func (agent *Agent) Reset(ctx context.Context) error {
	prefs := agent.GetMQTTPreferences()
	if prefs != nil && prefs.IsMQTTEnabled() {
		if err := agent.resetMQTTControllers(ctx); err != nil {
			agent.logger.Warn("Problems occurred resetting MQTT configuration.", slog.Any("error", err))
		}
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
	if agent.prefs != nil {
		agent.prefs.MQTT = prefs

		err := agent.prefs.Save()
		if err != nil {
			return fmt.Errorf("failed to save mqtt preferences: %w", err)
		}

		return nil
	}

	return ErrInvalidPrefernces
}

func (agent *Agent) GetRestAPIURL() string {
	return agent.prefs.RestAPIURL()
}
