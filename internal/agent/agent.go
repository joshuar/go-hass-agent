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
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/translations"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var Version string

var translator *translations.Translator

var debugAppID = ""

const (
	Name              = "go-hass-agent"
	fyneAppID         = "com.github.joshuar.go-hass-agent"
	issueURL          = "https://github.com/joshuar/autocorrector/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D"
	featureRequestURL = "https://github.com/joshuar/autocorrector/issues/new?assignees=&labels=&template=feature_request.md&title="
)

type Agent struct {
	app           fyne.App
	tray          fyne.Window
	Name, Version string
}

func NewAgent() *Agent {
	return &Agent{
		app:     newUI(),
		Name:    Name,
		Version: Version,
	}
}

func Run(id string) {
	if id != "" {
		debugAppID = id
	}
	agentCtx, cancelfunc := context.WithCancel(context.Background())
	agentCtx = device.SetupContext(agentCtx)
	log.Info().Msg("Starting agent.")
	agent := NewAgent()

	translator = translations.NewTranslator()

	agent.SetupLogging()

	// Try to load the app config. If it is not valid, start a new registration
	// process. Keep trying until we successfully register with HA or the user
	// quits.
	var configWg sync.WaitGroup
	configWg.Add(1)
	go func() {
		defer configWg.Done()
		agent.Load(agentCtx, agent.requestRegistrationInfoUI)
	}()

	// Wait for the config to load, then start the sensor tracker and
	// notifications worker
	var workerWg sync.WaitGroup
	defer workerWg.Done()
	go func() {
		configWg.Wait()
		appConfig := agent.loadAppConfig()
		ctx := config.StoreConfigInContext(agentCtx, appConfig)
		workerWg.Add(1)
		go func() {
			agent.runNotificationsWorker(ctx)
		}()
		workerWg.Add(1)
		go func() {
			agent.runSensorTracker(ctx)
		}()
	}()

	// Handle interrupt/termination signals
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancelfunc()
		workerWg.Wait()
		os.Exit(1)
	}()

	agent.setupSystemTray()
	agent.app.Run()
	agent.stop()
}

func RunHeadless(id string) {
	if id != "" {
		debugAppID = id
	}
	agentCtx, cancelfunc := context.WithCancel(context.Background())
	agentCtx = device.SetupContext(agentCtx)
	log.Info().Msg("Starting agent.")
	agent := NewAgent()

	agent.SetupLogging()

	// Wait for the config to load, then start the sensor tracker and
	// notifications worker
	var workerWg sync.WaitGroup
	go func() {
		appConfig := agent.loadAppConfig()
		ctx := config.StoreConfigInContext(agentCtx, appConfig)
		workerWg.Add(1)
		go func() {
			agent.runNotificationsWorker(ctx)
		}()
		workerWg.Add(1)
		go func() {
			agent.runSensorTracker(ctx)
		}()
	}()

	// Handle interrupt/termination signals
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancelfunc()
		workerWg.Wait()
		os.Exit(1)
	}()
	<-agentCtx.Done()
	agent.stop()
}

func (agent *Agent) stop() {
	log.Info().Msg("Shutting down agent.")
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

func (agent *Agent) SetupLogging() {
	// If possible, create and log to a file as well as the console.
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

func (agent *Agent) Load(ctx context.Context, registrationFetcher func(context.Context) *hass.RegistrationHost) {
	appConfig := agent.loadAppConfig()
	for appConfig.Validate() != nil {
		log.Warn().Msg("No suitable existing config found! Starting new registration process")
		err := agent.runRegistrationWorker(ctx, registrationFetcher)
		if err != nil {
			log.Error().Err(err).
				Msgf("Error trying to register: %v. Exiting.")
			agent.stop()
		}
		appConfig = agent.loadAppConfig()
	}
}
