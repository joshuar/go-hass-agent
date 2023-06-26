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
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/sensors"
	"github.com/joshuar/go-hass-agent/internal/translations"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var Version string

var translator *translations.Translator

var hassConfig *hass.HassConfig

var tracker *sensors.SensorTracker

const (
	Name              = "go-hass-agent"
	fyneAppID         = "com.github.joshuar.go-hass-agent"
)

type Agent struct {
	app           fyne.App
	tray          fyne.Window
	Name, Version string
}

func NewAgent(appID string) (context.Context, context.CancelFunc, *Agent) {
	a := &Agent{
		app:     newUI(appID),
		Name:    Name,
		Version: Version,
	}
	ctx, cancelfunc := context.WithCancel(context.Background())
	ctx = linux.SetupContext(ctx)
	a.SetupLogging()
	return ctx, cancelfunc, a
}

func Run(appID string) {
	translator = translations.NewTranslator()

	agentCtx, cancelFunc, agent := NewAgent(appID)
	log.Info().Msg("Started agent.")

	// Try to load the app config. If it is not valid, start a new registration
	// process. Keep trying until we successfully register with HA or the user
	// quits.
	var configOnce sync.Once
	go func() {
		configOnce.Do(func() {
			agent.CheckConfig(agentCtx, agent.requestRegistrationInfoUI)
		})
	}()

	// Wait for the config to load, then start the sensor tracker and
	// notifications worker
	var workerWg sync.WaitGroup
	defer workerWg.Done()
	go func() {
		configOnce.Do(func() {
			agent.CheckConfig(agentCtx, agent.requestRegistrationInfoUI)
		})

		appConfig := agent.LoadConfig()
		agentWorkerCtx := config.StoreInContext(agentCtx, appConfig)
		hassConfig = hass.NewHassConfig(agentWorkerCtx)
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			agent.runNotificationsWorker(agentWorkerCtx)
		}()
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			agent.runSensorTracker(agentWorkerCtx)
		}()
	}()

	// Handle interrupt/termination signals
	agent.handleSignals(cancelFunc, &workerWg)

	agent.setupSystemTray()
	agent.app.Run()
	cancelFunc()
	agent.stop()
}

func RunHeadless(appID string) {

	agentCtx, cancelFunc, agent := NewAgent(appID)
	log.Info().Msg("Started agent.")

	// Wait for the config to load, then start the sensor tracker and
	// notifications worker
	var workerWg sync.WaitGroup
	go func() {
		appConfig := agent.LoadConfig()
		agentWorkerCtx := config.StoreInContext(agentCtx, appConfig)
		hassConfig = hass.NewHassConfig(agentWorkerCtx)
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			agent.runSensorTracker(agentWorkerCtx)
		}()
	}()

	// Handle interrupt/termination signals
	agent.handleSignals(cancelFunc, &workerWg)

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

func (agent *Agent) CheckConfig(ctx context.Context, registrationFetcher func(context.Context) *hass.RegistrationHost) {
	config := agent.LoadConfig()
	for config.Validate() != nil {
		log.Warn().Msg("No suitable existing config found! Starting new registration process.")
		err := agent.runRegistrationWorker(ctx, registrationFetcher)
		if err != nil {
			log.Error().Err(err).
				Msg("Error trying to register.")
			agent.stop()
		} else {
			config = agent.LoadConfig()
		}
	}
}

func (agent *Agent) handleSignals(cancelFunc context.CancelFunc, wg *sync.WaitGroup) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancelFunc()
		wg.Wait()
		agent.stop()
		os.Exit(1)
	}()
}
