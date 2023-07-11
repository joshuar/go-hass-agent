// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
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
	"github.com/joshuar/go-hass-agent/internal/hass"
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

func (agent *Agent) setupSystemTray(ctx context.Context) {
	if desk, ok := agent.app.(desktop.App); ok {
		log.Debug().Msg("Running in desktop mode. Setting tray menu.")
		menuItemAbout := fyne.NewMenuItem(translator.Translate("About"), func() {
			agent.aboutWindow(ctx)
		})
		menuItemIssue := fyne.
			NewMenuItem(translator.Translate("Report Issue"),
				func() {
					url, _ := url.Parse(issueURL)
					if err := agent.app.OpenURL(url); err != nil {
						log.Warn().Err(err).
							Msgf("Unable to open url %s", url.String())
					}
				})
		menuItemFeatureRequest := fyne.
			NewMenuItem(translator.Translate("Request Feature"),
				func() {
					url, _ := url.Parse(featureRequestURL)
					if err := agent.app.OpenURL(url); err != nil {
						log.Warn().Err(err).
							Msgf("Unable to open url %s", url.String())
					}
				})
		menuItemSettings := fyne.
			NewMenuItem(translator.Translate("Settings"), agent.settingsWindow)
		menuItemSensors := fyne.
			NewMenuItem(translator.Translate("Sensors"), func() {
				agent.sensorsWindow(ctx)
			})
		menuItemQuit := fyne.NewMenuItem(translator.Translate("Quit"), func() {
			close(agent.done)
		})
		menuItemQuit.IsQuit = true
		menu := fyne.NewMenu(agent.Name,
			menuItemAbout,
			menuItemIssue,
			menuItemFeatureRequest,
			menuItemSettings,
			menuItemSensors,
			menuItemQuit)
		desk.SetSystemTrayMenu(menu)
	}
}

func (agent *Agent) aboutWindow(ctx context.Context) {
	deviceName, deviceID := agent.DeviceDetails()
	haVersion, err := hass.GetVersion(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Unable to version of Home Assistant.")
		return
	}
	w := agent.app.NewWindow(agent.setTitle("About"))
	w.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(translator.Translate(
			"App Version: %s\tHA Version: %s", agent.Version, haVersion)),
		widget.NewLabel(translator.Translate(
			"Device Name: "+deviceName)),
		widget.NewLabel(translator.Translate(
			"Device ID: "+deviceID)),
		widget.NewButton(translator.Translate("Ok"), func() {
			w.Close()
		}),
	))
	w.Show()

}

func (agent *Agent) settingsWindow() {
	w := agent.app.NewWindow(agent.setTitle("Fyne Settings"))
	w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
	w.Show()
}

func (agent *Agent) sensorsWindow(ctx context.Context) {
	var tableData [][]string
	var entityNames []string
	entities, err := hass.GetRegisteredEntities(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not get registered entities list from Home Assistant.")
		return
	}
	if entities == nil {
		log.Warn().
			Msg("No registered entities in Home Assistant.")
		return
	}
	for k := range entities {
		if sensor, err := sensorTracker.Get(k); err == nil {
			entityNames = append(entityNames, k)
			tableData = append(tableData,
				[]string{
					k,
					fmt.Sprintf("%v %s", sensor.State(), sensor.Units()),
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
			label, ok := o.(*widget.Label)
			if !ok {
				return
			}
			label.SetText(tableData[i.Row][i.Col])
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
