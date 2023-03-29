package agent

import (
	"context"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/carlmjohnson/requests"
	"github.com/jeandeaual/go-locale"
	"github.com/joshuar/go-hass-agent/assets/trayicon"
	"github.com/joshuar/go-hass-agent/internal/hass"
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

func newUI() fyne.App {
	a := app.NewWithID(fyneAppID)
	a.SetIcon(&trayicon.TrayIcon{})
	return a
}

func NewAgent() *Agent {
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

	go agent.runWorkers(&once)
	return agent
}

// func (a *Agent) GetHassConfig(configLoaded chan bool) {
// 	<-configLoaded
// 	a.hassConfig = hass.GetConfig(a.config.RestAPIURL)
// }

func (a *Agent) setupSystemTray() {
	// a.hassConfig = hass.GetConfig(a.config.RestAPIURL)
	a.Tray = a.App.NewWindow("System Tray")
	a.Tray.SetMaster()
	if desk, ok := a.App.(desktop.App); ok {
		log.Debug().Caller().
			Msg("Config loaded successfully. Creating tray icon.")

		menuItemAbout := fyne.NewMenuItem("About", func() {
			w := a.App.NewWindow(a.MsgPrinter.Sprintf("About ", a.Name))
			w.SetContent(container.New(layout.NewVBoxLayout(),
				widget.NewLabel(a.MsgPrinter.Sprintf("App Version: %s", a.Version)),
				// widget.NewLabel("Home Assistant Version: "+a.hassConfig.Version),
				widget.NewButton("Ok", func() {
					w.Close()
				}),
			))
			w.Show()
		})
		menu := fyne.NewMenu(a.Name, menuItemAbout)
		desk.SetSystemTrayMenu(menu)
	}
	a.Tray.Hide()
}

func (agent *Agent) runWorkers(once *sync.Once) {
	once.Do(func() { agent.LoadConfig() })
	go agent.runNotificationsWorker()
	go agent.runLocationWorker()
	go agent.runActiveAppSensor()
}

func (agent *Agent) Exit() {
	log.Debug().Caller().Msg("Shutting down agent.")
}

func (agent *Agent) PostRequest(ctx context.Context, request interface{}) interface{} {

	requestCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	// defer wg.Done()
	reqJson, err := hass.MarshalJSON(request.(hass.Request))
	if err != nil {
		log.Error().Msgf("Unable to format request: %v", err)
		return nil
	} else {
		var res interface{}
		err = requests.
			URL(agent.config.APIURL).
			BodyBytes(reqJson).
			ToJSON(&res).
			Fetch(requestCtx)
		if err != nil {
			log.Error().Msgf("Unable to send request: %v", err)
			return nil
		} else {
			return res
		}
	}
}
