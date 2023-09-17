// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ui

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-playground/validator/v10"
	"github.com/joshuar/go-hass-agent/assets/trayicon"
	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/translations"
	"github.com/rs/zerolog/log"
)

const (
	issueURL            = `https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D`
	featureRequestURL   = `https://github.com/joshuar/go-hass-agent/issues/new?assignees=joshuar&labels=&template=feature_request.md&title=`
	explainRegistration = `To register the agent, please enter the relevant details for your Home Assistant
server (if not auto-detected) and long-lived access token.`
)

type fyneUI struct {
	app        fyne.App
	mainWindow fyne.Window
}

func (ui *fyneUI) Run() {
	ui.app.Run()
}

func (ui *fyneUI) DisplayNotification(title, message string) {
	ui.app.SendNotification(&fyne.Notification{
		Title:   title,
		Content: message,
	})
}

func (ui *fyneUI) openURL(u string) {
	dest, _ := url.Parse(u)
	if err := ui.app.OpenURL(dest); err != nil {
		log.Warn().Err(err).
			Msgf("Unable to open url %s", dest.String())
	}
}

func NewFyneUI(agent Agent, headless bool) *fyneUI {
	ui := &fyneUI{
		app: app.NewWithID(agent.AppID()),
	}
	ui.app.SetIcon(&trayicon.TrayIcon{})
	if !headless {
		ui.mainWindow = ui.app.NewWindow(agent.AppName())
		ui.mainWindow.SetCloseIntercept(func() {
			ui.mainWindow.Hide()
		})
	}
	return ui
}

func (ui *fyneUI) DisplayTrayIcon(ctx context.Context, agent Agent) {
	if desk, ok := ui.app.(desktop.App); ok {
		t := translations.NewTranslator()
		menuItemQuit := fyne.NewMenuItem(t.Translate("Quit"), func() {
			agent.Stop()
		})
		menuItemQuit.IsQuit = true
		menu := fyne.NewMenu("Main",
			fyne.NewMenuItem(t.Translate("About"),
				func() {
					w := ui.aboutWindow(ctx, agent, t)
					if w != nil {
						w.Show()
					}
				}),
			fyne.NewMenuItem(t.Translate("Report Issue"),
				func() {
					ui.openURL(issueURL)
				}),
			fyne.NewMenuItem(t.Translate("Request Feature"),
				func() {
					ui.openURL(featureRequestURL)
				}),
			fyne.NewMenuItem(t.Translate("Fyne Settings"),
				func() {
					w := ui.fyneSettingsWindow(t)
					w.Show()
				}),
			fyne.NewMenuItem(t.Translate("App Settings"),
				func() {
					w := ui.agentSettingsWindow(agent, t)
					if w != nil {
						w.Show()
					}
				}),
			fyne.NewMenuItem(t.Translate("Sensors"),
				func() {
					w := ui.sensorsWindow(agent, t)
					if w != nil {
						w.Show()
					}
				}),
			menuItemQuit)
		desk.SetSystemTrayMenu(menu)
	}
}

// registrationWindow displays a UI to prompt the user for the details needed to
// complete registration. It will populate with any values that were already
// provided via the command-line.
func (ui *fyneUI) DisplayRegistrationWindow(ctx context.Context, done chan struct{}) {
	t := translations.NewTranslator()

	ui.mainWindow.SetTitle(t.Translate("App Registration"))

	var allFormItems []*widget.FormItem

	allFormItems = append(allFormItems, ui.serverConfigItems(ctx, t)...)
	registrationForm := widget.NewForm(allFormItems...)
	registrationForm.OnSubmit = func() {
		ui.mainWindow.Hide()
		close(done)
	}
	registrationForm.OnCancel = func() {
		log.Warn().Msg("Canceling registration.")
		close(done)
		ui.mainWindow.Close()
		ctx.Done()
	}

	ui.mainWindow.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(t.Translate(explainRegistration)),
		registrationForm,
	))

	ui.mainWindow.SetOnClosed(func() {
		log.Debug().Msg("Closed")
	})

	ui.mainWindow.Show()
}

// serverSelectionForm generates a fyne.CanvasObject consisting of a form for
// selecting a server to register the agent against
func (ui *fyneUI) serverConfigItems(ctx context.Context, t *translations.Translator) []*widget.FormItem {
	s := hass.FindServers(ctx)
	allServers, _ := s.Get()

	token := binding.BindPreferenceString(config.PrefToken, ui.app.Preferences())
	server := binding.BindPreferenceString(config.PrefHost, ui.app.Preferences())

	tokenSelect := widget.NewEntryWithData(token)
	tokenSelect.Validator = validation.NewRegexp("[A-Za-z0-9_\\.]+", "Invalid token format")

	autoServerSelect := widget.NewSelect(allServers, func(s string) {
		if err := server.Set(s); err != nil {
			log.Debug().Err(err).
				Msg("Could not set server pref to selected value.")
		}
	})

	manualServerEntry := widget.NewEntryWithData(server)
	manualServerEntry.Validator = HostValidator()
	manualServerEntry.Disable()
	manualServerSelect := widget.NewCheck("", func(b bool) {
		switch b {
		case true:
			manualServerEntry.Enable()
			autoServerSelect.Disable()
		case false:
			manualServerEntry.Disable()
			autoServerSelect.Enable()
		}
	})

	var items []*widget.FormItem

	items = append(items, widget.NewFormItem(t.Translate("Token"), tokenSelect),
		widget.NewFormItem(t.Translate("Auto-discovered Servers"), autoServerSelect),
		widget.NewFormItem(t.Translate("Use Custom Server?"), manualServerSelect),
		widget.NewFormItem(t.Translate("Manual Server Entry"), manualServerEntry))

	return items
}

func (ui *fyneUI) aboutWindow(ctx context.Context, agent Agent, t *translations.Translator) fyne.Window {
	var widgets []fyne.CanvasObject
	if hassConfig, err := hass.GetHassConfig(ctx, agent); err != nil {
		widgets = append(widgets, widget.NewLabel(t.Translate(
			"App Version: %s", agent.AppVersion())))
	} else {
		haVersion := hassConfig.GetVersion()
		widgets = append(widgets, widget.NewLabel(t.Translate(
			"App Version: %s\tHA Version: %s", agent.AppVersion(), haVersion)))
	}
	var deviceName, deviceID string
	if err := agent.GetConfig(config.PrefDeviceName, &deviceName); err == nil && deviceName != "" {
		widgets = append(widgets,
			widget.NewLabel(t.Translate("Device Name: "+deviceName)))
	}
	if err := agent.GetConfig(config.PrefDeviceID, &deviceID); err == nil && deviceID != "" {
		widgets = append(widgets,
			widget.NewLabel(t.Translate("Device ID: "+deviceID)))
	}
	w := ui.app.NewWindow(t.Translate("About"))
	cnt := container.New(layout.NewVBoxLayout(), widgets...)
	cnt.Add(widget.NewButton(t.Translate("Ok"), func() { w.Close() }))
	w.SetContent(cnt)
	return w
}

func (ui *fyneUI) fyneSettingsWindow(t *translations.Translator) fyne.Window {
	w := ui.app.NewWindow(t.Translate("Fyne Settings"))
	w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
	return w
}

func (ui *fyneUI) sensorsWindow(s Agent, t *translations.Translator) fyne.Window {
	sensors := s.SensorList()
	if sensors == nil {
		return nil
	}

	getValue := func(n string) string {
		if v, err := s.SensorValue(n); err == nil {
			return fmt.Sprintf("%v %s", v.State(), v.Units())
		}
		return ""
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
				label.SetText(getValue(sensors[i.Row]))
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
	// TODO: this is clunky. better way would be use Fyne bindings to sensor values
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
					}, widget.NewLabel(getValue(v)))
				}
				sensorsTable.Refresh()
			}
		}
	}()
	w := ui.app.NewWindow(t.Translate("Sensors"))
	w.SetContent(sensorsTable)
	w.Resize(fyne.NewSize(480, 640))
	w.SetOnClosed(func() {
		close(doneCh)
	})
	return w
}

func (ui *fyneUI) agentSettingsWindow(agent Agent, t *translations.Translator) fyne.Window {
	var allFormItems []*widget.FormItem
	allFormItems = append(allFormItems, ui.mqttConfigItems(agent, t)...)

	w := ui.app.NewWindow(t.Translate("App Settings"))
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
func (ui *fyneUI) mqttConfigItems(agent Agent, t *translations.Translator) []*widget.FormItem {
	mqttServer := binding.BindPreferenceString(config.PrefMQTTServer, ui.app.Preferences())
	mqttServerEntry := widget.NewEntryWithData(mqttServer)
	mqttServerEntry.Validator = HostValidator()
	mqttServerEntry.Disable()

	mqttTopic := binding.BindPreferenceString(config.PrefMQTTTopic, ui.app.Preferences())
	mqttTopicEntry := widget.NewEntryWithData(mqttTopic)
	mqttTopicEntry.Disable()

	mqttUser := binding.BindPreferenceString(config.PrefMQTTUser, ui.app.Preferences())
	mqttUserEntry := widget.NewEntryWithData(mqttUser)
	mqttUserEntry.Disable()

	mqttEnabled := widget.NewCheck("", func(b bool) {
		switch b {
		case true:
			mqttServerEntry.Enable()
			mqttTopicEntry.Enable()
			if err := agent.SetConfig("UseMQTT", true); err != nil {
				log.Warn().Err(err).Msg("Could not enable MQTT.")
			}
		case false:
			mqttServerEntry.Disable()
			mqttTopicEntry.Disable()
			if err := agent.SetConfig("UseMQTT", false); err != nil {
				log.Warn().Err(err).Msg("Could not disable MQTT.")
			}
		}
	})

	var items []*widget.FormItem

	items = append(items, widget.NewFormItem(t.Translate("Use MQTT?"), mqttEnabled),
		widget.NewFormItem(t.Translate("MQTT Server"), mqttServerEntry),
		widget.NewFormItem(t.Translate("MQTT Topic"), mqttTopicEntry))

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

// hostValidator is a custom fyne validator that will validate a string is a
// valid hostname:port combination
func HostValidator() fyne.StringValidator {
	v := validator.New()
	return func(text string) error {
		if v.Var(text, "http_url") != nil {
			return errors.New("you need to specify a valid url")
		}
		if _, err := url.Parse(text); err != nil {
			return errors.New("url is invalid")
		}
		return nil
	}
}
