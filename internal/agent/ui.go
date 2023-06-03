// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xlayout "fyne.io/x/fyne/layout"
	"github.com/joshuar/go-hass-agent/assets/trayicon"
	"github.com/rs/zerolog/log"
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
		w := agent.app.NewWindow(agent.setTitle("Fyne Settings"))
		w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		w.Show()
	}
	agent.tray = agent.app.NewWindow("go-hass-agent")
	agent.tray.SetMaster()
	if desk, ok := agent.app.(desktop.App); ok {
		menuItemAbout := fyne.NewMenuItem("About", func() {
			deviceName, deviceID := agent.DeviceDetails()
			hassVersion, _ := hassConfig.Get("version")
			w := agent.app.NewWindow(agent.setTitle("About"))
			w.SetContent(container.New(layout.NewVBoxLayout(),
				widget.NewLabel(translator.Translate(
					"App Version: %s\tHA Version: %s", agent.Version, hassVersion)),
				widget.NewLabel(translator.Translate(
					"Device Name: "+deviceName)),
				widget.NewLabel(translator.Translate(
					"Device ID: "+deviceID)),
				widget.NewButton(translator.Translate("Ok"), func() {
					w.Close()
				}),
			))
			w.Show()
		})
		menuItemIssue := fyne.
			NewMenuItem(translator.Translate("Report Issue"),
				func() {
					url, _ := url.Parse(issueURL)
					agent.app.OpenURL(url)
				})
		menuItemFeatureRequest := fyne.
			NewMenuItem(translator.Translate("Request Feature"),
				func() {
					url, _ := url.Parse(featureRequestURL)
					agent.app.OpenURL(url)
				})
		menuItemSettings := fyne.
			NewMenuItem(translator.Translate("Settings"), openSettings)
		menuItemSensors := fyne.
			NewMenuItem(translator.Translate("Sensors"), agent.makeSensorTable)
		menu := fyne.NewMenu(agent.Name,
			menuItemAbout,
			menuItemIssue,
			menuItemFeatureRequest,
			menuItemSettings,
			menuItemSensors)
		desk.SetSystemTrayMenu(menu)
	}
	agent.tray.Hide()
}

func (agent *Agent) makeSensorTable() {
	s, err := hassConfig.Get("entities")
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Could not get entities from config.")
		return
	}
	sensors := s.(map[string]map[string]interface{})
	var sensorsTable []fyne.CanvasObject
	for rowKey, rowValue := range sensors {
		var sensorRow []fyne.CanvasObject
		sensorRow = append(sensorRow, widget.NewLabel(rowKey))
		sensorState := tracker.Get(rowKey)
		if sensorState != nil {
			sensorRow = append(sensorRow, widget.NewLabel(fmt.Sprintf("%v %s",
				sensorState.State(), sensorState.Units())))
		} else {
			sensorRow = append(sensorRow, widget.NewLabel(""))
		}
		if rowValue["disabled"].(bool) {
			sensorRow = append(sensorRow, widget.NewLabel("Disabled"))
		} else {
			sensorRow = append(sensorRow, widget.NewLabel(""))
		}
		tableRow := container.New(layout.NewGridLayout(3), sensorRow...)
		sensorsTable = append(sensorsTable,
			xlayout.Responsive(tableRow))
	}
	table := xlayout.NewResponsiveLayout(sensorsTable...)
	layout := container.NewVScroll(table)
	w := agent.app.NewWindow(agent.setTitle("Sensors"))
	w.SetContent(layout)
	w.Resize(fyne.NewSize(480, 640))
	w.Show()
}

func (agent *Agent) setTitle(s string) string {
	return translator.Translate("%s: %s", agent.Name, s)
}
