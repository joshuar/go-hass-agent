// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/internal/translations"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var Version string

var translator *translations.Translator
var sensors *tracker.SensorTracker

const (
	Name = "go-hass-agent"
)

// Agent holds the data and structure representing an instance of the agent.
// This includes the data structure for the UI elements and tray and some
// strings such as app name and version.
type Agent struct {
	app        fyne.App
	mainWindow fyne.Window
	Config     AgentConfig
	done       chan struct{}
	Name       string
	Version    string
}

// AgentOptions holds options taken from the command-line that was used to
// invoke go-hass-agent that are relevant for agent functionality.
type AgentOptions struct {
	ID                 string
	Headless, Register bool
}

func newAgent(appID string, headless bool) *Agent {
	a := &Agent{
		app:     newUI(appID),
		Name:    Name,
		Version: Version,
		done:    make(chan struct{}),
		Config:  config.NewFyneConfig(),
	}
	if !headless {
		a.mainWindow = a.app.NewWindow(Name)
		a.mainWindow.SetCloseIntercept(func() {
			a.mainWindow.Hide()
		})
	}
	return a
}

// Run is the "main loop" of the agent. It sets up the agent, loads the config
// then spawns a sensor tracker and the workers to gather sensor data and
// publish it to Home Assistant
func Run(options AgentOptions) {
	translator = translations.NewTranslator()
	agent := newAgent(options.ID, options.Headless)
	defer close(agent.done)

	agentCtx, cancelFunc := context.WithCancel(context.Background())
	agent.setupLogging(agentCtx)

	registrationDone := make(chan struct{})
	go agent.registrationProcess(agentCtx, "", "", options.Register, options.Headless, registrationDone)

	var workerWg sync.WaitGroup
	trackerCh := make(chan *tracker.SensorTracker)
	go func() {
		<-registrationDone
		if err := UpgradeConfig(agent.Config); err != nil {
			log.Warn().Err(err).Msg("Could not upgrade config.")
		}
		if err := ValidateConfig(agent.Config); err != nil {
			log.Fatal().Err(err).Msg("Invalid config. Cannot start.")
		}
		// Start all the sensor workers as appropriate
		workerWg.Add(1)
		go func() {
			sensors = <-trackerCh
			if sensors == nil {
				log.Fatal().Msg("Could not start sensor tracker.")
			}
		}()
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			agent.runNotificationsWorker(agentCtx, options)
		}()
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			tracker.RunSensorTracker(agentCtx, agent.Config, trackerCh)
		}()
	}()
	agent.handleSignals(cancelFunc)
	agent.handleShutdown(agentCtx)

	// If we are not running in headless mode, show a tray icon
	if !options.Headless {
		agent.setupSystemTray(agentCtx)
		agent.app.Run()
	}
	workerWg.Wait()
	<-agentCtx.Done()
}

// Register runs a registration flow. It either prompts the user for needed
// information or parses what is already provided. It will send a registration
// request to Home Assistant and handles the response. It will handle either a
// UI or non-UI registration flow.
func Register(options AgentOptions, server, token string) {
	translator = translations.NewTranslator()
	agent := newAgent(options.ID, options.Headless)
	defer close(agent.done)

	agentCtx, cancelFunc := context.WithCancel(context.Background())

	// Don't proceed unless the agent is registered and forced is not set
	// if agent.IsRegistered() && !options.Register {
	// 	log.Warn().Msg("Agent is already registered and forced option not specified, not performing registration.")
	// 	cancelFunc()
	// 	return
	// }

	registrationDone := make(chan struct{})
	go agent.registrationProcess(agentCtx, server, token, options.Register, options.Headless, registrationDone)

	agent.handleSignals(cancelFunc)
	agent.handleShutdown(agentCtx)
	if !options.Headless {
		agent.setupSystemTray(agentCtx)
		agent.app.Run()
	}

	<-registrationDone
	log.Info().Msg("Device registered with Home Assistant.")
}

func ShowVersion(options AgentOptions) {
	agent := newAgent(options.ID, true)
	log.Info().Msgf("%s: %s", agent.Name, agent.Version)
}

func ShowInfo(options AgentOptions) {
	agent := newAgent(options.ID, true)
	var info strings.Builder
	var deviceName, deviceID string
	if err := agent.Config.Get(config.PrefDeviceName, &deviceName); err == nil && deviceName != "" {
		fmt.Fprintf(&info, "Device Name %s.", deviceName)
	}
	if err := agent.Config.Get(config.PrefDeviceID, &deviceID); err == nil && deviceID != "" {
		fmt.Fprintf(&info, "Device ID %s.", deviceID)
	}
	log.Info().Msg(info.String())
}

// setupLogging will attempt to create and then write logging to a file. If it
// cannot do this, logging will only be available on stdout
func (agent *Agent) setupLogging(ctx context.Context) {
	logFile, err := agent.Config.StoragePath("go-hass-app.log")
	if err != nil {
		log.Error().Err(err).
			Msg("Unable to create a log file. Will only write logs to stdout.")
	} else {
		logWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			log.Error().Err(err).
				Msg("Unable to open log file for writing.")
		} else {
			consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
			multiWriter := zerolog.MultiLevelWriter(consoleWriter, logWriter)
			log.Logger = log.Output(multiWriter)
			go func() {
				<-ctx.Done()
				logWriter.Close()
			}()
		}
	}
}

// handleSignals will handle Ctrl-C of the agent
func (agent *Agent) handleSignals(cancelFunc context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Debug().Msg("Ctrl-C pressed.")
		cancelFunc()
	}()
}

// handleShutdown will handle context cancellation of the agent
func (agent *Agent) handleShutdown(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Context canceled.")
				os.Exit(1)
			case <-agent.done:
				log.Debug().Msg("Agent done.")
				os.Exit(0)
			}
		}
	}()
}
