package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	fyneAppID = "com.github.joshuar.go-hass-agent"
)

func NewUI() fyne.App {
	return app.NewWithID(fyneAppID)
}
