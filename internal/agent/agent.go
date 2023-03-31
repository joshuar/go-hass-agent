package agent

import (
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
	done          chan bool
}

func NewAgent() *Agent {
	agent := &Agent{
		App:     newUI(),
		Name:    Name,
		Version: Version,
		done:    make(chan bool),
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

	go agent.runWorkers(&once)
	return agent
}

func (agent *Agent) runWorkers(once *sync.Once) {
	once.Do(func() { agent.loadConfig() })
	go agent.runNotificationsWorker()
	go agent.runLocationWorker()
	go agent.runAppSensorWorker()
	go agent.runBatterySensorWorker()
}

func (agent *Agent) Exit() {
	log.Debug().Caller().Msg("Shutting down agent.")
	close(agent.done)
	agent.App.Quit()
}
