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
	s, _ := hassConfig.Get("entities")
	sensors := s.(map[string]map[string]interface{})
	var sensorsTable [][3]string
	for rowKey, rowValue := range sensors {
		var sensorRow [3]string
		sensorRow[0] = rowKey
		sensorRow[1] = fmt.Sprintf("Disabled: %v", rowValue["disabled"])
		sensorState := tracker.Get(rowKey)
		if sensorState != nil {
			sensorRow[2] = fmt.Sprintf("%v %s",
				sensorState.State(), sensorState.UnitOfMeasurement())
		} else {
			sensorRow[2] = ""
		}
		sensorsTable = append(sensorsTable, sensorRow)
	}
	t := widget.NewTable(
		func() (int, int) { return len(sensors), 3 },
		func() fyne.CanvasObject {
			return widget.NewLabel("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			label.SetText(sensorsTable[id.Row][id.Col])
		})
	// t.SetColumnWidth(0, 34)
	// t.SetColumnWidth(1, 102)
	// t.SetRowHeight(2, 50)
	w := agent.app.NewWindow(agent.setTitle("Sensors"))
	w.SetContent(t)
	w.FullScreen()
	w.Show()
}

func (agent *Agent) setTitle(s string) string {
	return translator.Translate("%s: %s", agent.Name, s)
}
