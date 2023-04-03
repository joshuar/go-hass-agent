package agent

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/jeandeaual/go-locale"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	Name      = "go-hass-agent"
	Version   = "0.0.1"
	fyneAppID = "com.github.joshuar.go-hass-agent"
)

type Agent struct {
	App           fyne.App
	Tray          fyne.Window
	Name, Version string
	MsgPrinter    *message.Printer
	cancel        context.CancelFunc
}

func NewAgent(ctx context.Context) *Agent {
	agentCtx, cancel := context.WithCancel(ctx)
	agent := &Agent{
		App:     newUI(),
		Name:    Name,
		Version: Version,
		cancel:  cancel,
	}

	userLocales, err := locale.GetLocales()
	if err != nil {
		log.Warn().Msg("Could not find a suitable locale. Defaulting to English.")
		agent.MsgPrinter = message.NewPrinter(message.MatchLanguage(language.English.String()))
	} else {
		agent.MsgPrinter = message.NewPrinter(message.MatchLanguage(userLocales...))
		log.Debug().Caller().Msgf("Setting language to %v.", userLocales)
	}
	agent.setupSystemTray()

	var once sync.Once

	go agent.runWorkers(agentCtx, &once)
	return agent
}

func (agent *Agent) runWorkers(ctx context.Context, once *sync.Once) {
	appConfig := &config.AppConfig{}
	once.Do(func() { appConfig = agent.loadConfig(ctx) })
	workerCtx := config.NewContext(ctx, appConfig)
	go agent.runNotificationsWorker(workerCtx)
	go agent.runLocationWorker(workerCtx)
	go agent.runAppSensorWorker(workerCtx)
	go agent.runBatterySensorWorker(workerCtx)
}

func (agent *Agent) Exit() {
	log.Debug().Caller().Msg("Shutting down agent.")
	agent.cancel()
	agent.Tray.Close()
	agent.App.Quit()
}
