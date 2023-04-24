// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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
	agent.tray = agent.app.NewWindow("System Tray")
	agent.tray.SetMaster()
	if desk, ok := agent.app.(desktop.App); ok {
		menuItemAbout := fyne.NewMenuItem("About", func() {
			deviceName, deviceID := agent.DeviceDetails()
			w := agent.app.NewWindow(translator.Translate("About %s", agent.Name))
			w.SetContent(container.New(layout.NewVBoxLayout(),
				widget.NewLabel(translator.Translate(
					"App Version: %s", agent.Version)),
				widget.NewLabel(translator.Translate(
					"Device Name: "+deviceName)),
				widget.NewLabel(translator.Translate(
					"Device ID: "+deviceID)),
				widget.NewButton("Ok", func() {
					w.Close()
				}),
			))
			w.Show()
		})
		menu := fyne.NewMenu(agent.Name, menuItemAbout)
		desk.SetSystemTrayMenu(menu)
	}
	agent.tray.Hide()
}
