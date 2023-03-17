package agent

import (
	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/ui"
)

type Agent struct {
	ui fyne.App
}

func NewAgent() *Agent {
	return &Agent{
		ui: ui.NewUI(),
	}
}
