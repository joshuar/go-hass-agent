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
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/joshuar/go-hass-agent/assets/trayicon"
	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

const (
	issueURL          = `https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D`
	featureRequestURL = `https://github.com/joshuar/go-hass-agent/issues/new?assignees=&labels=&template=feature_request.md&title=`
)

func newUI(appID string) fyne.App {
	a := app.NewWithID(appID)
	a.SetIcon(&trayicon.TrayIcon{})
	return a
}

func (agent *Agent) setupSystemTray(ctx context.Context) {
	if desk, ok := agent.app.(desktop.App); ok {
		menuItemQuit := fyne.NewMenuItem(translator.Translate("Quit"), func() {
			close(agent.done)
		})
		menuItemQuit.IsQuit = true
		menu := fyne.NewMenu(agent.Name,
			fyne.NewMenuItem(translator.Translate("About"),
				func() {
					w := agent.aboutWindow(ctx)
					if w != nil {
						w.Show()
					}
				}),
			fyne.NewMenuItem(translator.Translate("Report Issue"),
				func() {
					dest, _ := url.Parse(issueURL)
					if err := agent.app.OpenURL(dest); err != nil {
						log.Warn().Err(err).
							Msgf("Unable to open url %s", dest.String())
					}
				}),
			fyne.NewMenuItem(translator.Translate("Request Feature"),
				func() {
					dest, _ := url.Parse(featureRequestURL)
					if err := agent.app.OpenURL(dest); err != nil {
						log.Warn().Err(err).
							Msgf("Unable to open url %s", dest.String())
					}
				}),
			fyne.NewMenuItem(translator.Translate("Fyne Settings"),
				func() {
					w := agent.fyneSettingsWindow()
					w.Show()
				}),
			fyne.NewMenuItem(translator.Translate("App Settings"),
				func() {
					w := agent.agentSettingsWindow()
					if w != nil {
						w.Show()
					}
				}),
			fyne.NewMenuItem(translator.Translate("Sensors"),
				func() {
					w := agent.sensorsWindow(ctx)
					if w != nil {
						w.Show()
					}
				}),
			menuItemQuit)
		desk.SetSystemTrayMenu(menu)
	}
}

func (agent *Agent) aboutWindow(ctx context.Context) fyne.Window {
	var widgets []fyne.CanvasObject
	if hassConfig, err := hass.GetHassConfig(ctx, agent.Config); err != nil {
		widgets = append(widgets, widget.NewLabel(translator.Translate(
			"App Version: %s", agent.Version)))
	} else {
		haVersion := hassConfig.GetVersion()
		widgets = append(widgets, widget.NewLabel(translator.Translate(
			"App Version: %s\tHA Version: %s", agent.Version, haVersion)))
	}
	var deviceName, deviceID string
	if err := agent.Config.Get(config.PrefDeviceName, &deviceName); err == nil && deviceName != "" {
		widgets = append(widgets,
			widget.NewLabel(translator.Translate("Device Name: "+deviceName)))
	}
	if err := agent.Config.Get(config.PrefDeviceID, &deviceID); err == nil && deviceID != "" {
		widgets = append(widgets,
			widget.NewLabel(translator.Translate("Device ID: "+deviceID)))
	}
	w := agent.app.NewWindow(translator.Translate("About"))
	c := container.New(layout.NewVBoxLayout(), widgets...)
	c.Add(widget.NewButton(translator.Translate("Ok"), func() { w.Close() }))
	w.SetContent(c)
	return w
}

func (agent *Agent) fyneSettingsWindow() fyne.Window {
	w := agent.app.NewWindow(translator.Translate("Fyne Settings"))
	w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
	return w
}

func (agent *Agent) sensorsWindow(ctx context.Context) fyne.Window {
	sensors := getSensorsAsList(ctx, agent.Config)
	if sensors == nil {
		return nil
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
	return w
}

func (agent *Agent) agentSettingsWindow() fyne.Window {
	var allFormItems []*widget.FormItem
	allFormItems = append(allFormItems, agent.mqttConfigItems()...)

	w := agent.app.NewWindow(translator.Translate("App Settings"))
	settingsForm := widget.NewForm(allFormItems...)
	w.SetContent(container.New(layout.NewVBoxLayout(),
		settingsForm,
		widget.NewLabel("Changes will be saved automatically."),
	))
	w.SetOnClosed(func() {
		log.Debug().Msg("Closed")
	})
	return w
}

// mqttConfigForm returns a fyne.CanvasObject consisting of a form for
// configuring the agent to use an MQTT for pub/sub functionality
func (agent *Agent) mqttConfigItems() []*widget.FormItem {
	mqttServer := binding.BindPreferenceString(config.PrefMQTTServer, agent.app.Preferences())
	mqttServerEntry := widget.NewEntryWithData(mqttServer)
	mqttServerEntry.Validator = hostValidator()
	mqttServerEntry.Disable()

	mqttTopic := binding.BindPreferenceString(config.PrefMQTTTopic, agent.app.Preferences())
	mqttTopicEntry := widget.NewEntryWithData(mqttTopic)
	mqttTopicEntry.Disable()

	mqttUser := binding.BindPreferenceString(config.PrefMQTTUser, agent.app.Preferences())
	mqttUserEntry := widget.NewEntryWithData(mqttUser)
	mqttUserEntry.Disable()

	mqttEnabled := widget.NewCheck("", func(b bool) {
		switch b {
		case true:
			mqttServerEntry.Enable()
			mqttTopicEntry.Enable()
			if err := agent.Config.Set("UseMQTT", true); err != nil {
				log.Warn().Err(err).Msg("Could not enable MQTT.")
			}
		case false:
			mqttServerEntry.Disable()
			mqttTopicEntry.Disable()
			if err := agent.Config.Set("UseMQTT", false); err != nil {
				log.Warn().Err(err).Msg("Could not disable MQTT.")
			}
		}
	})

	var items []*widget.FormItem

	items = append(items, widget.NewFormItem(translator.Translate("Use MQTT?"), mqttEnabled),
		widget.NewFormItem(translator.Translate("MQTT Server"), mqttServerEntry),
		widget.NewFormItem(translator.Translate("MQTT Topic"), mqttTopicEntry))

	return items
}

func longestString(a []string) string {
	var l string
	if len(a) > 0 {
		l = a[0]
		a = a[1:]
	}
	for _, s := range a {
		if len(l) <= len(s) {
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
