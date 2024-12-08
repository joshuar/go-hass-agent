// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver
//
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

// CtxOption is a functional parameter that will add a value to the agent
// context.
type CtxOption func(context.Context) context.Context

// LoadCtx will "load" a context.Context with the given options (i.e. add values
// to it to be used by the agent).
func LoadCtx(ctx context.Context, options ...CtxOption) context.Context {
	for _, option := range options {
		ctx = option(ctx) //nolint:fatcontext
	}

	return ctx
}

// SetHeadless sets the headless flag in the context.
func SetHeadless(value bool) CtxOption {
	return func(ctx context.Context) context.Context {
		ctx = addToContext(ctx, headlessCtxKey, value)
		return ctx
	}
}

// SetRegistrationInfo sets registration details in the context to be used for
// registering the agent.
func SetRegistrationInfo(server, token string, ignoreURLs bool) CtxOption {
	return func(ctx context.Context) context.Context {
		ctx = addToContext(ctx, serverCtxKey, server)
		ctx = addToContext(ctx, tokenCtxKey, token)
		ctx = addToContext(ctx, ignoreURLsCtxKey, ignoreURLs)

		return ctx
	}
}

// ForceRegister sets the forceregister flag in the context.
func SetForceRegister(value bool) CtxOption {
	return func(ctx context.Context) context.Context {
		ctx = addToContext(ctx, forceRegisterCtxKey, value)
		return ctx
	}
}

// Run is the "main loop" of the agent. It sets up the agent, loads the config
// then spawns a sensor tracker and the workers to gather sensor data and
// publish it to Home Assistant.
//
//nolint:funlen
//revive:disable:function-length
func Run(ctx context.Context) error {
	var (
		wg      sync.WaitGroup
		regWait sync.WaitGroup
		err     error
	)

	agent := &Agent{}

	// If running headless, do not set up the UI.
	if !Headless(ctx) {
		agent.ui = fyneui.NewFyneUI(ctx)
	}

	// Load the preferences from file. Ignore the case where there are no
	// existing preferences.
	if err = preferences.Load(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return fmt.Errorf("%w: %w", ErrAgentStart, err)
	}

	// Set up a context for running the agent and tie its lifetime to the
	// typical process termination signals.
	runCtx, cancelRun := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancelRun()
	}()

	regWait.Add(1)

	go func() {
		defer regWait.Done()
		// Check if the agent is registered. If not, start a registration flow.
		if err = checkRegistration(runCtx, agent.ui); err != nil {
			logging.FromContext(ctx).Error("Error checking registration status.", slog.Any("error", err))
			cancelRun()
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
		sensorWorkers, eventWorkers := setupOSWorkers(runCtx)
		// Initialize and add connection latency sensor worker.
		sensorWorkers = append(sensorWorkers, agentsensor.NewConnectionLatencySensorWorker())
		// Initialize and add external IP address sensor worker.
		sensorWorkers = append(sensorWorkers, agentsensor.NewExternalIPUpdaterWorker(runCtx))
		// Initialize and add external version sensor worker.
		sensorWorkers = append(sensorWorkers, agentsensor.NewVersionWorker())

		// Initialize and add the script worker.
		scriptsWorkers, err := scripts.NewScriptsWorker(runCtx)
		if err != nil {
			logging.FromContext(runCtx).Warn("Could not init scripts workers.", slog.Any("error", err))
		} else {
			sensorWorkers = append(sensorWorkers, scriptsWorkers)
		}

		wg.Add(1)
		// Process sensor workers.
		go func() {
			defer wg.Done()
			processWorkers(runCtx, client, sensorWorkers...)
		}()

		wg.Add(1)
		// Process event workers.
		go func() {
			defer wg.Done()
			processWorkers(runCtx, client, eventWorkers...)
		}()

		// If MQTT is enabled, init MQTT workers and process them.
		if preferences.MQTTEnabled() {
			if mqttPrefs, err := preferences.GetMQTTPreferences(); err != nil {
				logging.FromContext(runCtx).Warn("Could not init mqtt workers.",
					slog.Any("error", err))
			} else {
				mqttWorkers := setupMQTT(runCtx)

				wg.Add(1)

				go func() {
					defer wg.Done()
					processMQTTWorkers(runCtx, mqttPrefs, mqttWorkers...)
				}()
			}
		}

		wg.Add(1)
		// Listen for notifications from Home Assistant.
		go func() {
			defer wg.Done()
			runNotificationsWorker(runCtx, agent.ui)
		}()
	}()

	// Do not run the UI loop if the agent is running in headless mode.
	if !Headless(ctx) {
		agent.ui.DisplayTrayIcon(runCtx, cancelRun)
		agent.ui.Run(runCtx)
	}

	wg.Wait()

	return nil
}

func Register(ctx context.Context) error {
	var (
		wg  sync.WaitGroup
		err error
	)

	agent := &Agent{}
	// If running headless, do not set up the UI.
	if !Headless(ctx) {
		agent.ui = fyneui.NewFyneUI(ctx)
	}

	if err = preferences.Load(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
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

	if !Headless(ctx) {
		agent.ui.Run(regCtx)
	}

	wg.Wait()

	return nil
}

// Reset will remove any agent related files and configuration.
func Reset(ctx context.Context) error {
	if err := preferences.Load(ctx); err != nil && !errors.Is(err, preferences.ErrLoadPreferences) {
		return fmt.Errorf("%w: %w", ErrAgentStart, err)
	}

	if preferences.MQTTEnabled() {
		if err := resetMQTTWorkers(ctx); err != nil {
			logging.FromContext(ctx).Error("Problems occurred resetting MQTT configuration.", slog.Any("error", err))
		}
	}

	return nil
}
