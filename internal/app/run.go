// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"log/slog"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/joshuar/go-hass-agent/internal/mqtt"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

func Run(ctx context.Context, headless bool, appAPIs APIs) error {
	var (
		wg      sync.WaitGroup
		regWait sync.WaitGroup
		err     error
	)

	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	// Add device-specific values to context.
	ctx = device.SetupCtx(ctx)

	app := New(ctx, appAPIs, headless)

	regWait.Add(1)

	go func() {
		defer regWait.Done()
		// Check if the agent is registered. If not, start a registration flow.
		if err = checkRegistration(ctx, app, headless); err != nil {
			app.logger.Error("Error checking registration status.", slog.Any("error", err))
			cancelFunc()
		}
	}()
	// Run entity workers.
	wg.Add(1)
	go app.runEntityWorkers(ctx, appAPIs, &wg, &regWait)
	// Run mqtt workers.
	wg.Add(1)
	go app.runMQTTWorkers(ctx, &wg, &regWait)
	// Run notifications worker.
	if !headless {
		wg.Add(1)
		go app.runNotificationsWorker(ctx, &wg, &regWait)
	}

	// Do not run the UI loop if the agent is running in headless mode.
	if !headless {
		app.ui.DisplayTrayIcon(ctx, cancelFunc)
		app.ui.Run(ctx)
	}

	wg.Wait()

	return nil
}

// runEntityWorkers will start all entity workers and pass their data on to the
// hass client.
func (a *App) runEntityWorkers(ctx context.Context, appAPIs APIs, wg *sync.WaitGroup, regwg *sync.WaitGroup) {
	defer wg.Done()
	// Wait until registration check completes.
	regwg.Wait()
	// If the agent is not registered, bail.
	if !preferences.Registered() {
		return
	}
	// Create entity workers.
	var entityWorkers []workers.EntityWorker
	// Add device-based entity workers.
	entityWorkers = append(entityWorkers, device.CreateDeviceEntityWorkers(ctx)...)
	// Add os-based entity workers.
	entityWorkers = append(entityWorkers, device.CreateOSEntityWorkers(device.SetupCtx(ctx))...)
	// Start all entity workers.
	entityCh := a.workerManager.StartEntityWorkers(ctx, entityWorkers...)

	go func() {
		defer a.workerManager.StopAllWorkers()
		<-ctx.Done()
	}()

	// Get hass client to handle entity workers.
	appAPIs.Hass().EntityHandler(ctx, entityCh)
}

// runMQTTWorkers will start all MQTT workers and pass their data to the MQTT client.
func (a *App) runMQTTWorkers(ctx context.Context, wg *sync.WaitGroup, regwg *sync.WaitGroup) {
	defer wg.Done()
	// Wait until registration check completes.
	regwg.Wait()
	// If the agent is not registered and MQTT is not enabled, bail.
	if !(preferences.Registered() && preferences.MQTTEnabled()) {
		return
	}
	// Create MQTT workers.
	var mqttWorkers []workers.MQTTWorker
	// Add device-based MQTT workers.
	mqttWorkers = append(mqttWorkers, device.CreateDeviceMQTTWorkers(ctx)...)
	// Add os-based MQTT workers.
	mqttWorkers = append(mqttWorkers, device.CreateOSMQTTWorkers(ctx))
	// Start all MQTT workers.
	data := a.workerManager.StartMQTTWorkers(ctx, mqttWorkers...)
	if err := mqtt.Start(ctx, data); err != nil {
		a.logger.Warn("Unable to start MQTT client.",
			slog.Any("error", err),
		)
	}
}

// runNotificationsWorker will run a goroutine that is listening for
// notification messages from Home Assistant on a websocket connection. Any
// received notifications will be dipslayed on the device running the agent.
func (a *App) runNotificationsWorker(ctx context.Context, wg *sync.WaitGroup, regwg *sync.WaitGroup) {
	defer wg.Done()
	// Wait until registration check completes.
	regwg.Wait()
	// If the agent is not registered, bail.
	if !preferences.Registered() {
		return
	}

	websocket := api.NewWebsocket(ctx,
		preferences.WebsocketURL(),
		preferences.WebhookID(),
		preferences.Token())

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Connect the websocket.
			notifyCh, err := websocket.Connect(ctx)
			if err != nil {
				logging.FromContext(ctx).Warn("Failed to connect to websocket.",
					slog.Any("error", err))

				return
			}
			// Start listening on the websocket
			go func() {
				websocket.Listen()
			}()
			// Display any notifications received.
			for notification := range notifyCh {
				a.ui.DisplayNotification(&notification)
			}
		}
	}
}
