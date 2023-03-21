package agent

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

const (
	fyneAppID = "com.github.joshuar.go-hass-agent"
)

func NewUI() fyne.App {
	a := app.NewWithID(fyneAppID)
	a.SetIcon(theme.FyneLogo())
	return a
}
