// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	fyneui "github.com/joshuar/go-hass-agent/internal/agent/ui/fyneUI"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/preferences"
)

// Agent holds the data and structure representing an instance of the agent.
// This includes the data structure for the UI elements and tray and some
// strings such as app name and version.
type Agent struct {
	ui      UI
	done    chan struct{}
	Options *Options
}

// Options holds options taken from the command-line that was used to
// invoke go-hass-agent that are relevant for agent functionality.
type Options struct {
	ID, Server, Token       string
	Headless, ForceRegister bool
}

func New(o *Options) *Agent {
	a := &Agent{
		done:    make(chan struct{}),
		Options: o,
	}
	if !a.Options.Headless {
		a.ui = fyneui.NewFyneUI(a.Options.ID)
	}
	return a
}

// Run is the "main loop" of the agent. It sets up the agent, loads the config
// then spawns a sensor tracker and the workers to gather sensor data and
// publish it to Home Assistant.
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

		var err error
		prefs, err := preferences.Load()
		if err != nil {
			log.Fatal().Err(err).Msg("Could not load preferences.")
		}
		ctx, cancelFunc := setupContext(prefs)
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
		if prefs.MQTTEnabled {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runMQTTWorker(runnerCtx)
			}()
		}
		// Listen for notifications from Home Assistant.
		if !agent.IsHeadless() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				agent.runNotificationsWorker(runnerCtx)
			}()
		}
	}()

	agent.handleSignals()

	if !agent.IsHeadless() {
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

	if !agent.IsHeadless() {
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

// IsHeadless returns a bool indicating whether the agent is running in
// "headless" mode (i.e., without a GUI) or not.
func (agent *Agent) IsHeadless() bool {
	return agent.Options.Headless
}

// AppID returns the "application ID". Currently, this ID is just used to
// indicate whether the agent is running in debug mode or not.
func (agent *Agent) AppID() string {
	return agent.Options.ID
}

// Stop will close the agent's done channel which indicates to any goroutines it
// is time to clean up and exit.
func (agent *Agent) Stop() {
	log.Debug().Msg("Stopping agent.")
	close(agent.done)
}

// setupContext embeds the config object in a context which allows access to it
// from any functions that inherit this context. This is used early in the agent
// start up to ensure all subsequent functionality can access config details as
// needed.
func setupContext(prefs *preferences.Preferences) (context.Context, context.CancelFunc) {
	baseCtx, cancelFunc := context.WithCancel(context.Background())
	agentCtx := hass.ContextSetURL(baseCtx, prefs.RestAPIURL)
	agentCtx = hass.ContextSetClient(agentCtx, resty.New())
	return agentCtx, cancelFunc
}
