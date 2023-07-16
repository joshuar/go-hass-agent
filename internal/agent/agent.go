// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	_ "embed"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/tracker"
	"github.com/joshuar/go-hass-agent/internal/translations"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var Version string

var translator *translations.Translator

var sensorTracker *tracker.SensorTracker

const (
	Name = "go-hass-agent"
)

// Agent holds the data and structure representing an instance of the agent.
// This includes the data structure for the UI elements and tray and some
// strings such as app name and version.
type Agent struct {
	app           fyne.App
	mainWindow    fyne.Window
	done          chan struct{}
	Name, Version string
}

// AgentOptions holds options taken from the command-line that was used to
// invoke go-hass-agent that are relevant for agent functionality.
type AgentOptions struct {
	ID                 string
	Headless, Register bool
}

func newAgent(appID string) (context.Context, context.CancelFunc, *Agent) {
	a := &Agent{
		app:     newUI(appID),
		Name:    Name,
		Version: Version,
		done:    make(chan struct{}),
	}
	a.mainWindow = a.app.NewWindow(Name)
	ctx, cancelfunc := context.WithCancel(context.Background())
	ctx = linux.SetupContext(ctx)
	a.setupLogging()
	return ctx, cancelfunc, a
}

// Run is the "main loop" of the agent. It sets up the agent, loads the config
// then spawns a sensor tracker and the workers to gather sensor data and
// publish it to Home Assistant
func Run(options AgentOptions) {
	translator = translations.NewTranslator()
	agentCtx, cancelFunc, agent := newAgent(options.ID)
	defer close(agent.done)

	registrationDone := make(chan struct{})
	go agent.registrationProcess(agentCtx, "", "", options.Register, options.Headless, registrationDone)

	var workerWg sync.WaitGroup

	go func() {
		<-registrationDone
		// Load the config. If it is not valid, exit
		appConfig := agent.LoadConfig()
		if err := appConfig.Upgrade(); err != nil {
			log.Warn().Err(err).Msg("Could not upgrade config.")
		}
		if err := ValidateConfig(appConfig); err != nil {
			log.Fatal().Err(err).Msg("Invalid config. Cannot start.")
		}
		// Store relevant settings from appConfig in a new context for workers
		agentCtx = StoreSettings(agentCtx, appConfig)
		// Start all the sensor workers as appropriate
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			agent.runNotificationsWorker(agentCtx, options)
		}()
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			agent.runSensorTracker(agentCtx)
		}()
	}()
	agent.handleSignals(cancelFunc)
	agent.handleShutdown(agentCtx)

	// If we are not running in headless mode, show a tray icon
	if !options.Headless {
		agent.setupSystemTray(agentCtx)
		log.Debug().Msg("Starting main UI loop.")
		agent.app.Run()
		log.Debug().Msg("Finished UI loop.")
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
	agentCtx, cancelFunc, agent := newAgent(options.ID)
	defer close(agent.done)

	// Don't proceed unless the agent is registered and forced is not set
	if agent.IsRegistered() && !options.Register {
		log.Warn().Msg("Agent is already registered and forced option not specified, not performing registration.")
		return
	}

	registrationDone := make(chan struct{})
	go agent.registrationProcess(agentCtx, server, token, options.Register, options.Headless, registrationDone)

	agent.handleSignals(cancelFunc)
	agent.handleShutdown(agentCtx)
	if !options.Headless {
		agent.app.Run()
	}

	<-registrationDone
	log.Info().Msg("Device registered with Home Assistant.")
}

func ShowVersion(options AgentOptions) {
	_, _, agent := newAgent(options.ID)
	log.Info().Msgf("%s: %s", agent.Name, agent.Version)
}

func ShowInfo(options AgentOptions) {
	_, _, agent := newAgent(options.ID)
	deviceName, deviceID := agent.DeviceDetails()
	log.Info().Msgf("Device Name %s. Device ID %s.", deviceName, deviceID)

}

func (agent *Agent) DeviceDetails() (string, string) {
	return agent.app.Preferences().String("DeviceName"),
		agent.app.Preferences().String("DeviceID")
}

func (agent *Agent) IsRegistered() bool {
	return agent.app.Preferences().BoolWithFallback("Registered", false)
}

func (agent *Agent) SetRegistered(value bool) {
	agent.app.Preferences().SetBool("Registered", value)
}

func (agent *Agent) extraStoragePath(id string) (fyne.URI, error) {
	rootPath := agent.app.Storage().RootURI()
	extraPath, err := storage.Child(rootPath, id)
	if err != nil {
		return nil, err
	} else {
		return extraPath, nil
	}
}

// setupLogging will attempt to create and then write logging to a file. If it
// cannot do this, logging will only be available on stdout
func (agent *Agent) setupLogging() {
	logFile, err := agent.extraStoragePath("go-hass-app.log")
	if err != nil {
		log.Error().Err(err).
			Msg("Unable to create a log file. Will only write logs to stdout.")
	} else {
		logWriter, err := storage.Writer(logFile)
		if err != nil {
			log.Error().Err(err).
				Msg("Unable to open log file for writing.")
		} else {
			consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
			multiWriter := zerolog.MultiLevelWriter(consoleWriter, logWriter)
			log.Logger = log.Output(multiWriter)
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
				log.Debug().Msg("Context cancelled.")
				os.Exit(1)
			case <-agent.done:
				log.Debug().Msg("Agent done.")
				os.Exit(0)
			}
		}
	}()
}
