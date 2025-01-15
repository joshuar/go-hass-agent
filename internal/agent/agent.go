// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//go:generate go run github.com/matryer/moq -out agent_mocks_test.go . ui
package agent

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joshuar/go-hass-agent/internal/agent/agentsensor"
	fyneui "github.com/joshuar/go-hass-agent/internal/agent/ui/fyneUI"
	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
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
	ui     ui
	logger *slog.Logger
}

// newAgent creates the Agent struct.
func newAgent(ctx context.Context) *Agent {
	agent := &Agent{
		logger: logging.FromContext(ctx).WithGroup("agent"),
	}

	// If not running headless, set up the UI.
	if !preferences.Headless() {
		agent.ui = fyneui.NewFyneUI(ctx)
	}

	return agent
}

// Run is invoked when Go Hass Agent is run with the `run` command-line option
// (i.e., `go-hass-agent run`).
//
//nolint:funlen
func Run(ctx context.Context, dataCh chan any) error {
	var (
		wg      sync.WaitGroup
		regWait sync.WaitGroup
		err     error
	)

	defer close(dataCh)

	// Create struct.
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	agent := newAgent(ctx)

	regWait.Add(1)

	go func() {
		defer regWait.Done()
		// Check if the agent is registered. If not, start a registration flow.
		if err = checkRegistration(ctx, agent.ui); err != nil {
			agent.logger.Error("Error checking registration status.", slog.Any("error", err))
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

		// Set-up the worker context.
		workerCtx := setupWorkerCtx(ctx)

		// Initialize and gather OS sensor and event workers.
		sensorWorkers, eventWorkers := setupOSWorkers(workerCtx)
		// Initialize and add connection latency sensor worker.
		if worker := agentsensor.NewConnectionLatencySensorWorker(); worker != nil {
			sensorWorkers = append(sensorWorkers, worker)
		}
		// Initialize and add external IP address sensor worker.
		if worker := agentsensor.NewExternalIPUpdaterWorker(workerCtx); worker != nil {
			sensorWorkers = append(sensorWorkers, worker)
		}
		// Initialize and add external version sensor worker.
		if worker := agentsensor.NewVersionWorker(); worker != nil {
			sensorWorkers = append(sensorWorkers, worker)
		}

		// Initialize and add the script worker.
		scriptsWorkers, err := scripts.NewScriptsWorker(workerCtx)
		if err != nil {
			agent.logger.Warn("Could not init scripts workers.",
				slog.Any("error", err))
		} else {
			sensorWorkers = append(sensorWorkers, scriptsWorkers)
		}

		wg.Add(1)
		// Process sensor workers.
		go func() {
			defer wg.Done()
			processWorkers(workerCtx, dataCh, sensorWorkers...)
		}()

		wg.Add(1)
		// Process event workers.
		go func() {
			defer wg.Done()
			processWorkers(workerCtx, dataCh, eventWorkers...)
		}()

		wg.Add(1)
		// Process MQTT workers.
		go func() {
			defer wg.Done()
			processMQTTWorkers(workerCtx)
		}()

		wg.Add(1)
		// Listen for notifications from Home Assistant.
		go func() {
			defer wg.Done()
			runNotificationsWorker(workerCtx, agent.ui)
		}()
	}()

	// Do not run the UI loop if the agent is running in headless mode.
	if !preferences.Headless() {
		agent.ui.DisplayTrayIcon(ctx, cancelFunc)
		agent.ui.Run(ctx)
	}

	wg.Wait()

	return nil
}

// Register is run when Go Hass Agent is invoked with the `register`
// command-line option (i.e., `go-hass-agent register`). It will attempt to
// register Go Hass Agent with Home Assistant.
func Register(ctx context.Context) error {
	var wg sync.WaitGroup

	agent := newAgent(ctx)

	regCtx, cancelReg := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancelReg()
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := checkRegistration(regCtx, agent.ui); err != nil {
			agent.logger.Error("Error checking registration status", slog.Any("error", err))
		}

		cancelReg()
	}()

	if !preferences.Headless() {
		agent.ui.Run(regCtx)
	}

	wg.Wait()

	return nil
}

// Reset is invoked when Go Hass Agent is run with the `reset` command-line
// option (i.e., `go-hass-agent reset`).
func Reset(ctx context.Context) error {
	agent := newAgent(ctx)
	// If MQTT is enabled, reset any saved MQTT config.
	if preferences.MQTTEnabled() {
		if err := resetMQTTWorkers(ctx); err != nil {
			agent.logger.Error("Problems occurred resetting MQTT configuration.", slog.Any("error", err))
		}
	}
	// Agent reset complete.
	return nil
}
