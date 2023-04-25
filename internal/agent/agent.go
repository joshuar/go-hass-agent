// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	_ "embed"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/sensors"
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
	Name      = "go-hass-agent"
	fyneAppID = "com.github.joshuar.go-hass-agent"
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
	ctx, cancelfunc := context.WithCancel(context.Background())
	ctx = device.SetupContext(ctx)
	log.Info().Msg("Starting agent.")
	agent := NewAgent()

	translator = translations.NewTranslator()

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

	var wg sync.WaitGroup

	// Try to load the app config. If it is not valid, start a new registration
	// process. Keep trying until we successfully register with HA or the user
	// quits.
	wg.Add(1)
	go func() {
		defer wg.Done()
		appConfig := agent.loadAppConfig()
		for appConfig.Validate() != nil {
			log.Warn().Msg("No suitable existing config found! Starting new registration process")
			err := agent.runRegistrationWorker(ctx)
			if err != nil {
				log.Error().Err(err).
					Msgf("Error trying to register: %v. Exiting.")
				agent.stop()
			}
			appConfig = agent.loadAppConfig()
		}
	}()

	agent.setupSystemTray()

	go agent.tracker(ctx, &wg)

	agent.app.Run()
	cancelfunc()
	agent.stop()
}

func (agent *Agent) stop() {
	log.Info().Msg("Shutting down agent.")
	agent.tray.Close()
}

func (agent *Agent) getStorageURI() fyne.URI {
	return agent.app.Storage().RootURI()
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

// tracker should be run in a goroutine and is responsible for creating,
// tracking and updating HA with all sensors provided from the platform/device.
func (agent *Agent) tracker(agentCtx context.Context, configWG *sync.WaitGroup) {
	configWG.Wait()

	appConfig := agent.loadAppConfig()
	ctx := config.NewContext(agentCtx, appConfig)
	registryPath, err := agent.extraStoragePath("sensorRegistry")
	if err != nil {
		log.Debug().Err(err).
			Msg("Unable to store registry on disk, trying in-memory store.")
	}
	sensorTracker := sensors.NewSensorTracker(ctx, registryPath)
	updateCh := make(chan interface{})

	go agent.runNotificationsWorker(ctx)

	// goroutine to listen for sensor updates. Sensors are tracked in a map to
	// handle registration and disabling/enabling. Updates are sent to Home
	// Assistant.
	go func() {
		for {
			select {
			case data := <-updateCh:
				switch data := data.(type) {
				case hass.SensorUpdate:
					go sensorTracker.Update(ctx, data)
				case hass.LocationUpdate:
					l := hass.MarshalLocationUpdate(data)
					go hass.APIRequest(ctx, l)
				default:
					log.Debug().Caller().
						Msgf("Got unexpected status update %v", data)
				}
			case <-ctx.Done():
				log.Debug().Caller().
					Msg("Stopping sensor tracking.")
				return
			}
		}
	}()

	sensorTracker.StartWorkers(ctx, updateCh)
}
