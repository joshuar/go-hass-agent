package agent

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	log "github.com/sirupsen/logrus"
)

const (
	fyneAppID = "com.github.joshuar.go-hass-agent"
)

func NewUI() fyne.App {
	a := app.NewWithID(fyneAppID)
	a.SetIcon(theme.FyneLogo())
	return a
}

func (a *Agent) StartTrayIcon() fyne.Window {
	if desk, ok := a.App.(desktop.App); ok {
		log.Debug("Creating tray icon")
		h := fyne.NewMenuItem("Hello", func() {})
		h.Icon = theme.HomeIcon()
		menu := fyne.NewMenu("Hello World", h)
		desk.SetSystemTrayMenu(menu)
	}
	return a.App.NewWindow(a.Name)
}
