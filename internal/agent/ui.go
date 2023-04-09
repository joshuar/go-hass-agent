package agent

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/joshuar/go-hass-agent/assets/trayicon"
)

func newUI() fyne.App {
	a := app.NewWithID(fyneAppID)
	a.SetIcon(&trayicon.TrayIcon{})
	return a
}

func (agent *Agent) setupSystemTray() {
	// a.hassConfig = hass.GetConfig(a.config.RestAPIURL)
	agent.Tray = agent.App.NewWindow("System Tray")
	agent.Tray.SetMaster()
	if desk, ok := agent.App.(desktop.App); ok {
		menuItemAbout := fyne.NewMenuItem("About", func() {
			w := agent.App.NewWindow(agent.MsgPrinter.Sprintf("About %s", agent.Name))
			w.SetContent(container.New(layout.NewVBoxLayout(),
				widget.NewLabel(agent.MsgPrinter.Sprintf("App Version: %s", agent.Version)),
				// widget.NewLabel("Home Assistant Version: "+a.hassConfig.Version),
				widget.NewButton("Ok", func() {
					w.Close()
				}),
			))
			w.Show()
		})
		menu := fyne.NewMenu(agent.Name, menuItemAbout)
		desk.SetSystemTrayMenu(menu)
	}
	agent.Tray.Hide()
}
