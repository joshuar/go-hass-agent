// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// revive:disable:unused-receiver
//
//go:generate moq -out agent_mocks_test.go . ui SensorController MQTTController
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

	fyneui "github.com/joshuar/go-hass-agent/internal/agent/ui/fyneUI"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

var ErrInvalidPreferences = errors.New("invalid agent preferences")

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
//revive:disable:function-length
func Run(ctx context.Context) error {
	var (
		wg      sync.WaitGroup
		regWait sync.WaitGroup
		prefs   *preferences.Preferences
		err     error
	)

	agent := &Agent{}

	// If running headless, do not set up the UI.
	if !Headless(ctx) {
		agent.ui = fyneui.NewFyneUI(ctx)
	}

	// Load the preferences from file. Ignore the case where there are no
	// existing preferences.
	prefs, err = preferences.Load(ctx)
	if err != nil && !errors.Is(err, preferences.ErrNoPreferences) {
		return fmt.Errorf("could not load preferences: %w", err)
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
		if err = checkRegistration(runCtx, agent.ui, prefs.GetDeviceInfo(), prefs); err != nil {
			logging.FromContext(ctx).Error("Error checking registration status.", slog.Any("error", err))
			cancelRun()
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		regWait.Wait()

		// If the agent is not registered, bail.
		if !prefs.AgentRegistered() {
			return
		}

		var (
			sensorControllers []SensorController
			mqttControllers   []MQTTController
		)
		// Setup and sort all controllers by type.
		for _, c := range agent.setupControllers(runCtx, prefs) {
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
			runSensorWorkers(runCtx, prefs, sensorControllers...)
		}()

		if len(mqttControllers) > 0 {
			wg.Add(1)
			// Run workers for any MQTT controllers.
			go func() {
				defer wg.Done()
				runMQTTWorkers(runCtx, prefs.GetMQTTPreferences(), mqttControllers...)
			}()
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
		wg    sync.WaitGroup
		prefs *preferences.Preferences
		err   error
	)

	agent := &Agent{}
	// If running headless, do not set up the UI.
	if !Headless(ctx) {
		agent.ui = fyneui.NewFyneUI(ctx)
	}

	prefs, err = preferences.Load(ctx)
	if err != nil && !errors.Is(err, preferences.ErrNoPreferences) {
		return fmt.Errorf("could not load preferences: %w", err)
	}

	regCtx, cancelReg := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancelReg()
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := checkRegistration(regCtx, agent.ui, prefs.GetDeviceInfo(), prefs); err != nil {
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
	prefs, err := preferences.Load(ctx)
	if err != nil && !errors.Is(err, preferences.ErrNoPreferences) {
		return fmt.Errorf("could not load preferences: %w", err)
	}

	if prefs.IsMQTTEnabled() {
		if err := resetMQTTControllers(ctx, prefs.GenerateMQTTDevice(ctx), prefs.GetMQTTPreferences()); err != nil {
			logging.FromContext(ctx).Error("Problems occurred resetting MQTT configuration.", slog.Any("error", err))
		}
	}

	return nil
}
