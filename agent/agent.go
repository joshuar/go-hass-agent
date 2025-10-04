// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

// Package agent defines the core functionality for running the agent.
package agent

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gen2brain/beeep"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent/workers"
	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/hass"
	"github.com/joshuar/go-hass-agent/hass/api"
)

//go:embed assets/icon.png
var icon []byte

var registered chan struct{}

const (
	ConfigPrefix = "agent"
)

// Agent represents the data and methods required for running the agent.
type Agent struct {
	Config *Config
}

// Config contains the agent configuration options.
type Config struct {
	Registered bool `toml:"registered"`
}

// New sets up a new agent instance.
func New() (*Agent, error) {
	registered = make(chan struct{})

	agent := &Agent{
		Config: &Config{
			Registered: false,
		},
	}
	// Load the server config.
	if err := config.Load(ConfigPrefix, agent.Config); err != nil {
		return agent, fmt.Errorf("unable to load agent config: %w", err)
	}
	// Pick up legacy (pre v14) registered value in config and rewrite into new location.
	if config.Exists("registered") {
		registered, err := config.Get[bool]("registered")
		if err != nil {
			return agent, fmt.Errorf("unable to load agent config: %w", err)
		}
		err = config.Set(map[string]any{"agent.registered": registered})
		if err != nil {
			return agent, fmt.Errorf("unable to load agent config: %w", err)
		}
	}
	// Check for registration and flag as appropriate.
	if agent.IsRegistered() {
		close(registered)
	}
	// Return the controller object.
	return agent, nil
}

// IsRegistered returns a boolean indicating whether the agent has been registered.
func (a *Agent) IsRegistered() bool {
	return a.Config.Registered
}

// Register will mark the registration status of the agent as registered.
func (a *Agent) Register() {
	a.Config.Registered = true
	err := config.Save(ConfigPrefix, a.Config)
	if err != nil {
		slog.Error("Unable to save registration status to config.",
			slog.Any("error", err))
		return
	}
	close(registered)
}

// Reset will undo registration status of the agent.
func (a *Agent) Reset() {
	a.Config.Registered = false
	err := config.Save(ConfigPrefix, a.Config)
	if err != nil {
		slog.Error("Unable to save registration status to config.",
			slog.Any("error", err))
		return
	}
	registered = make(chan struct{})
}

// Run is the main loop of the agent. It will configure and run all sensor workers and process and send the data to Home
// Assistant. Run blocks and won't perform any actions until the registration status of the agent is true.
func (a *Agent) Run(ctx context.Context) error {
	ctx = workers.SetupCtx(ctx)
	for {
		select {
		case <-registered:
			slog.Debug("Agent is registered.")
			hassClient, err := hass.NewClient(ctx, a)
			if err != nil {
				return fmt.Errorf("unable to run agent: %w", err)
			}
			manager := workers.NewManager(ctx)
			var wg sync.WaitGroup
			// Entity/Event workers.
			wg.Go(func() {
				// Create entity workers.
				var entityWorkers []workers.EntityWorker
				// Add device-based entity workers.
				entityWorkers = append(entityWorkers, CreateDeviceEntityWorkers(ctx, hassClient.RestAPIURL())...)
				// Add os-based entity workers.
				entityWorkers = append(entityWorkers, CreateOSEntityWorkers(ctx)...)
				// Start all entity workers.
				entityCh := manager.StartEntityWorkers(ctx, entityWorkers...)

				go func() {
					defer manager.StopAllWorkers(ctx)
					<-ctx.Done()
				}()

				// Get hass client to handle entity workers.
				hassClient.EntityHandler(ctx, entityCh)
			})
			// MQTT workers.
			wg.Go(func() {
				// Don't continue if MQTT isn't configured.
				if !config.Exists("mqtt") {
					slog.Debug("Not configuring MQTT functionality, not configured.")
					return
				}
				// Get MQTT status.
				enabled, err := config.Get[bool]("mqtt.enabled")
				if err != nil {
					slogctx.FromCtx(ctx).Warn("Unable to start device MQTT workers.",
						slog.Any("error", err))
					return
				}
				// Don't continue if MQTT is explicitly disabled.
				if !enabled {
					slog.Debug("Not starting MQTT workers, MQTT functionality explicitly disabled.")
					return
				}
				// Create MQTT workers.
				var mqttWorkers []workers.MQTTWorker
				// Add device-based MQTT workers.
				deviceMQTTworkers, err := CreateDeviceMQTTWorkers(ctx)
				if err != nil {
					slogctx.FromCtx(ctx).Warn("Unable to start device MQTT workers.",
						slog.Any("error", err))
				}
				mqttWorkers = append(mqttWorkers, deviceMQTTworkers...)
				// Add os-based MQTT workers.
				osMQTTworkers, err := CreateOSMQTTWorkers(ctx)
				if err != nil {
					slogctx.FromCtx(ctx).Warn("Unable to start OS MQTT workers.",
						slog.Any("error", err))
				}
				mqttWorkers = append(mqttWorkers, osMQTTworkers)
				// Start all MQTT workers.
				data := manager.StartMQTTWorkers(ctx, mqttWorkers...)
				if err := mqtt.Start(ctx, data); err != nil {
					slogctx.FromCtx(ctx).Warn("Unable to start MQTT.",
						slog.Any("error", err))
				}
			})
			// Run notification worker.
			wg.Go(func() {
				beeep.AppName = config.AppName
				websocket, err := api.NewWebsocket(ctx)
				if err != nil {
					slogctx.FromCtx(ctx).Warn("Unable to listen for notifications.",
						slog.Any("error", err))
					return
				}
				for {
					select {
					case <-ctx.Done():
						return
					default:
						// Connect the websocket.
						notifyCh, err := websocket.Connect(ctx)
						if err != nil {
							slogctx.FromCtx(ctx).Warn("Failed to connect to websocket.",
								slog.Any("error", err))

							return
						}
						// Start listening on the websocket
						go func() {
							websocket.Listen()
						}()
						// Display any notifications received.
						for notification := range notifyCh {
							err := beeep.Notify(notification.Title, notification.Message, icon)
							if err != nil {
								slogctx.FromCtx(ctx).Warn("Unable to send notification.",
									slog.Any("error", err))
							}
						}
					}
				}
			})
			wg.Wait()
		case <-ctx.Done():
			slog.Debug("Stopping agent.")
			return nil
		}
	}
}
