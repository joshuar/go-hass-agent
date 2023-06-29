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
	"fyne.io/fyne/v2/widget"
	"github.com/joshuar/go-hass-agent/assets/trayicon"
	"github.com/rs/zerolog/log"
)

const (
	issueURL          = "https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D"
	featureRequestURL = "https://github.com/joshuar/go-hass-agent/issues/new?assignees=&labels=&template=feature_request.md&title="
)

func newUI(appID string) fyne.App {
	a := app.NewWithID(appID)
	a.SetIcon(&trayicon.TrayIcon{})
	return a
}

func (agent *Agent) setupSystemTray() {
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
			NewMenuItem(translator.Translate("Settings"), agent.settingsWindow)
		menuItemSensors := fyne.
			NewMenuItem(translator.Translate("Sensors"), agent.sensorsWindow)
		menu := fyne.NewMenu(agent.Name,
			menuItemAbout,
			menuItemIssue,
			menuItemFeatureRequest,
			menuItemSettings,
			menuItemSensors)
		desk.SetSystemTrayMenu(menu)
	}
}

func (agent *Agent) settingsWindow() {
	w := agent.app.NewWindow(agent.setTitle("Fyne Settings"))
	w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
	w.Show()
}

func (agent *Agent) sensorsWindow() {
	var tableData [][]string
	var entityNames []string
	s, err := hassConfig.Get("entities")
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Could not get entities from config.")
		return
	}
	for k := range s.(map[string]map[string]interface{}) {
		if state := tracker.Get(k); state != nil {
			entityNames = append(entityNames, k)
			tableData = append(tableData,
				[]string{
					k,
					fmt.Sprintf("%v %s",
						state.State(), state.Units()),
				})
		}
	}

	longestName := longestString(entityNames)

	list := widget.NewTable(
		func() (int, int) {
			return len(tableData), len(tableData[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel(longestName)
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(tableData[i.Row][i.Col])
		})
	w := agent.app.NewWindow(agent.setTitle("Sensors"))
	w.SetContent(list)
	w.Resize(fyne.NewSize(480, 640))
	w.Show()
}

func (agent *Agent) setTitle(s string) string {
	return translator.Translate("%s: %s", agent.Name, s)
}

func longestString(a []string) string {
	var l string
	if len(a) > 0 {
		l = a[0]
		a = a[1:]
	}
	for _, s := range a {
		if len(l) <= len(s) {
			// if len(l) < len(s) {
			// 	l = l[:0]
			// }
			l = s
		}
	}
	return l
}
