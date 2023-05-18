// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/joshuar/go-hass-agent/assets/trayicon"
)

func newUI(appID string) fyne.App {
	var a fyne.App
	if appID != "" {
		a = app.NewWithID(appID)
		a.SetIcon(theme.FyneLogo())

	} else {
		a = app.NewWithID(fyneAppID)
		a.SetIcon(&trayicon.TrayIcon{})
	}
	return a
}

func (agent *Agent) setupSystemTray() {
	openSettings := func() {
		w := agent.app.NewWindow("Fyne Settings")
		w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		w.Resize(fyne.NewSize(480, 480))
		w.Show()
	}
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
		menuItemIssue := fyne.NewMenuItem("Report Issue", func() {
			url, _ := url.Parse(issueURL)
			agent.app.OpenURL(url)
		})
		menuItemFeatureRequest := fyne.NewMenuItem("Request Feature", func() {
			url, _ := url.Parse(featureRequestURL)
			agent.app.OpenURL(url)
		})
		menuItemSettings := fyne.NewMenuItem("Settings", openSettings)
		menu := fyne.NewMenu(agent.Name,
			menuItemAbout,
			menuItemIssue,
			menuItemFeatureRequest,
			menuItemSettings)
		desk.SetSystemTrayMenu(menu)
	}
	agent.tray.Hide()
}
