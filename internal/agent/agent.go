// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go run github.com/matryer/moq -out agent_mocks_test.go . ui
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

	"github.com/joshuar/go-hass-agent/internal/agent/agentsensor"
	fyneui "github.com/joshuar/go-hass-agent/internal/agent/ui/fyneUI"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/internal/scripts"
)

var ErrAgentStart = errors.New("cannot start agent")

// ui are the methods required for the agent to display its windows, tray
// and notifications.
type ui interface {
	DisplayNotification(n fyneui.Notification)
	DisplayTrayIcon(ctx context.Context, cancelFunc context.CancelFunc)
	DisplayRegistrationWindow(ctx context.Context, prefs *preferences.Registration) chan bool
	Run(ctx context.Context)
}

// Agent represents a running agent.
type Agent struct {
	ui ui
}

// agentOptions are the options that can be used to change the behavior of Go
// Hass Agent.
type agentOptions struct {
	headless           bool
	forceRegister      bool
	ignoreHassURLs     bool
	registrationServer string
	registrationToken  string
}

// options represents the set of run-time options for the current instance of Go
// Hass Agent.
var options agentOptions

// Option represents a run-time option for Go Hass Agent.
type Option func(agentOptions) agentOptions

// setup will establish the options for this run of Go Hass Agent.
func setup(givenOptions ...Option) {
	for _, option := range givenOptions {
		options = option(options)
	}
}

// SetHeadless indicates Go Hass Agent should run without a UI.
func SetHeadless(value bool) Option {
	return func(ao agentOptions) agentOptions {
		ao.headless = value
		return ao
	}
}

// SetRegistrationInfo contains the information Go Hass Agent should use to
// register with Home Assistant.
func SetRegistrationInfo(server, token string, ignoreURLs bool) Option {
	return func(ao agentOptions) agentOptions {
		ao.ignoreHassURLs = ignoreURLs
		ao.registrationServer = server
		ao.registrationToken = token

		return ao
	}
}

// ForceRegister indicates that Go Hass Agent should ignore its current
// registration status and re-register with Home Assistant.
func SetForceRegister(value bool) Option {
	return func(ao agentOptions) agentOptions {
		ao.forceRegister = value
		return ao
	}
}

// newCtx creates a new context for this run of Go Hass Agent.
func newCtx(logger *slog.Logger) (context.Context, context.CancelFunc) {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	ctx = logging.ToContext(ctx, logger)

	return ctx, cancelFunc
}

// newAgent creates the Agent struct.
func newAgent(ctx context.Context) *Agent {
	agent := &Agent{}

	// If not running headless, set up the UI.
	if !options.headless {
		agent.ui = fyneui.NewFyneUI(ctx)
	}

	return agent
}

// Run is invoked when Go Hass Agent is run with the `run` command-line option
// (i.e., `go-hass-agent run`).
//
//nolint:funlen
//revive:disable:function-length
func Run(logger *slog.Logger, givenOptions ...Option) error {
	var (
		wg      sync.WaitGroup
		regWait sync.WaitGroup
		err     error
	)

	// Establish run-time options.
	setup(givenOptions...)
	// Setup context.
	ctx, cancelFunc := newCtx(logger)
	defer cancelFunc()
	// Create struct.
	agent := newAgent(ctx)

	// Load the preferences from file. Ignore the case where there are no
	// existing preferences.
	if err = preferences.Load(); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return fmt.Errorf("%w: %w", ErrAgentStart, err)
	}

	regWait.Add(1)

	go func() {
		defer regWait.Done()
		// Check if the agent is registered. If not, start a registration flow.
		if err = checkRegistration(ctx, agent.ui); err != nil {
			logging.FromContext(ctx).Error("Error checking registration status.", slog.Any("error", err))
			cancelFunc()
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		regWait.Wait()

		// If the agent is not registered, bail.
		if !preferences.Registered() {
			return
		}

		client, err := hass.NewClient(ctx, preferences.RestAPIURL())
		if err != nil {
			logging.FromContext(ctx).Error("Cannot connect to Home Assistant.",
				slog.Any("error", err))
			return
		}

		// Initialize and gather OS sensor and event workers.
		sensorWorkers, eventWorkers := setupOSWorkers(ctx)
		// Initialize and add connection latency sensor worker.
		sensorWorkers = append(sensorWorkers, agentsensor.NewConnectionLatencySensorWorker())
		// Initialize and add external IP address sensor worker.
		sensorWorkers = append(sensorWorkers, agentsensor.NewExternalIPUpdaterWorker(ctx))
		// Initialize and add external version sensor worker.
		sensorWorkers = append(sensorWorkers, agentsensor.NewVersionWorker())

		// Initialize and add the script worker.
		scriptsWorkers, err := scripts.NewScriptsWorker(ctx)
		if err != nil {
			logging.FromContext(ctx).Warn("Could not init scripts workers.", slog.Any("error", err))
		} else {
			sensorWorkers = append(sensorWorkers, scriptsWorkers)
		}

		wg.Add(1)
		// Process sensor workers.
		go func() {
			defer wg.Done()
			processWorkers(ctx, client, sensorWorkers...)
		}()

		wg.Add(1)
		// Process event workers.
		go func() {
			defer wg.Done()
			processWorkers(ctx, client, eventWorkers...)
		}()

		// If MQTT is enabled, init MQTT workers and process them.
		if preferences.MQTTEnabled() {
			if mqttPrefs, err := preferences.GetMQTTPreferences(); err != nil {
				logging.FromContext(ctx).Warn("Could not init mqtt workers.",
					slog.Any("error", err))
			} else {
				mqttWorkers := setupMQTT(ctx)

				wg.Add(1)

				go func() {
					defer wg.Done()
					processMQTTWorkers(ctx, mqttPrefs, mqttWorkers...)
				}()
			}
		}

		wg.Add(1)
		// Listen for notifications from Home Assistant.
		go func() {
			defer wg.Done()
			runNotificationsWorker(ctx, agent.ui)
		}()
	}()

	// Do not run the UI loop if the agent is running in headless mode.
	if !options.headless {
		agent.ui.DisplayTrayIcon(ctx, cancelFunc)
		agent.ui.Run(ctx)
	}

	wg.Wait()

	return nil
}

// Register is run when Go Hass Agent is invoked with the `register`
// command-line option (i.e., `go-hass-agent register`). It will attempt to
// register Go Hass Agent with Home Assistant.
func Register(logger *slog.Logger, givenOptions ...Option) error {
	var (
		wg  sync.WaitGroup
		err error
	)

	// Establish run-time options.
	setup(givenOptions...)
	// Setup context.
	ctx, cancelFunc := newCtx(logger)
	defer cancelFunc()
	// Create struct.
	agent := newAgent(ctx)

	if err = preferences.Load(); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return fmt.Errorf("%w: %w", ErrAgentStart, err)
	}

	regCtx, cancelReg := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancelReg()
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := checkRegistration(regCtx, agent.ui); err != nil {
			logging.FromContext(ctx).Error("Error checking registration status", slog.Any("error", err))
		}

		cancelReg()
	}()

	if !options.headless {
		agent.ui.Run(regCtx)
	}

	wg.Wait()

	return nil
}

// Reset is invoked when Go Hass Agent is run with the `reset` command-line
// option (i.e., `go-hass-agent reset`).
func Reset(logger *slog.Logger, givenOptions ...Option) error {
	// Establish run-time options.
	setup(givenOptions...)
	// Setup context.
	ctx, cancelFunc := newCtx(logger)
	defer cancelFunc()

	// Load the preferences so we know what we need to reset.
	if err := preferences.Load(); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return fmt.Errorf("%w: %w", ErrAgentStart, err)
	}

	// If MQTT is enabled, reset any saved MQTT config.
	if preferences.MQTTEnabled() {
		if err := resetMQTTWorkers(ctx); err != nil {
			logging.FromContext(ctx).Error("Problems occurred resetting MQTT configuration.", slog.Any("error", err))
		}
	}

	return nil
}
