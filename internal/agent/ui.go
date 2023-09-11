// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

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
	issueURL          = `https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D`
	featureRequestURL = `https://github.com/joshuar/go-hass-agent/issues/new?assignees=&labels=&template=feature_request.md&title=`
	firstRunText      = `Welcome to go-hass-agent. As this is the first run of the agent, a window will be displayed 
for you to enter registration details. Please enter the required details, 
click Submit and the agent should start running.`
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

func (agent *Agent) showFirstRunWindow(ctx context.Context) {
	w := agent.app.NewWindow(translator.Translate("First Run"))
	w.SetContent(container.NewVBox(
		widget.NewLabel(translator.Translate(firstRunText)),
		widget.NewButton(translator.Translate("Ok"), func() { w.Hide() })))
	w.CenterOnScreen()
	w.Show()
	w.SetMaster()
}

func (agent *Agent) aboutWindow(ctx context.Context) {
	deviceName, deviceID := agent.DeviceDetails()
	hassConfig, err := hass.GetHassConfig(ctx, agent.Config)
	if err != nil {
		log.Warn().Err(err).
			Msg("Unable to version of Home Assistant.")
		return
	}
	haVersion := hassConfig.GetVersion()
	w := agent.app.NewWindow(translator.Translate("About"))
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
	agent.mainWindow.SetTitle(translator.Translate("Fyne Settings"))
	agent.mainWindow.SetContent(settings.NewSettings().LoadAppearanceScreen(agent.mainWindow))
	agent.mainWindow.Show()
}

func (agent *Agent) sensorsWindow(ctx context.Context) {
	sensors := getSensorsAsList(ctx, agent.Config)
	if sensors == nil {
		return
	}
	sensorsTable := widget.NewTableWithHeaders(
		func() (int, int) {
			return len(sensors), 2
		},
		func() fyne.CanvasObject {
			return widget.NewLabel(longestString(sensors))
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label, ok := o.(*widget.Label)
			if !ok {
				return
			}
			switch i.Col {
			case 0:
				label.SetText(sensors[i.Row])
			case 1:
				label.SetText(getSensorValue(sensors[i.Row]))
			}
		})
	sensorsTable.ShowHeaderColumn = false
	sensorsTable.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabel("Header")
	}
	sensorsTable.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		label, ok := template.(*widget.Label)
		if !ok {
			return
		}
		if id.Row == -1 && id.Col == 0 {
			label.SetText("Sensor")
		}
		if id.Row == -1 && id.Col == 1 {
			label.SetText("Value")
		}
	}
	doneCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-doneCh:
				return
			case <-ticker.C:
				for i, v := range sensors {
					sensorsTable.UpdateCell(widget.TableCellID{
						Row: i,
						Col: 1,
					}, widget.NewLabel(getSensorValue(v)))
				}
				sensorsTable.Refresh()
			}
		}
	}()
	w := agent.app.NewWindow(translator.Translate("First Run"))
	w.SetTitle(translator.Translate("Sensors"))
	w.SetContent(sensorsTable)
	w.Resize(fyne.NewSize(480, 640))
	w.SetOnClosed(func() {
		close(doneCh)
	})
	w.Show()
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

func getSensorsAsList(ctx context.Context, cfg AgentConfig) []string {
	if sensors == nil {
		log.Warn().Msg("No sensors available.")
		return nil
	}
	hassConfig, err := hass.GetHassConfig(ctx, cfg)
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not get registered entities list from Home Assistant.")
		return nil
	}
	entities := hassConfig.GetRegisteredEntities()
	if entities == nil {
		log.Warn().
			Msg("No registered entities in Home Assistant.")
		return nil
	}
	sortedEntities := make([]string, 0, len(entities))
	for k := range entities {
		if s, err := sensors.Get(k); err == nil && s.State() != nil {
			sortedEntities = append(sortedEntities, k)
		}
	}
	sort.Strings(sortedEntities)
	return sortedEntities
}

func getSensorValue(sensor string) string {
	if s, err := sensors.Get(sensor); err == nil {
		return fmt.Sprintf("%v %s", s.State(), s.Units())
	}
	return ""
}
