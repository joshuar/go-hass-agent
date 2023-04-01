package agent

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/jeandeaual/go-locale"
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
	config        AppConfig
	Name, Version string
	MsgPrinter    *message.Printer
}

func NewAgent() (*Agent, context.Context, context.CancelFunc) {
	agentCtx, cancel := context.WithCancel(context.Background())
	agent := &Agent{
		App:     newUI(),
		Name:    Name,
		Version: Version,
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
	return agent, agentCtx, cancel
}

func (agent *Agent) runWorkers(ctx context.Context, once *sync.Once) {
	once.Do(func() { agent.loadConfig(ctx) })
	go agent.runNotificationsWorker(ctx)
	go agent.runLocationWorker(ctx)
	go agent.runAppSensorWorker(ctx)
	go agent.runBatterySensorWorker(ctx)
}

func (agent *Agent) Exit() {
	log.Debug().Caller().Msg("Shutting down agent.")
	agent.Tray.Close()
	agent.App.Quit()
}
