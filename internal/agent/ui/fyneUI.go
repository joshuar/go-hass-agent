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
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-playground/validator/v10"
	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/translations"
	"github.com/rs/zerolog/log"
)

const (
	explainRegistration = `To register the agent, please enter the relevant details for your Home Assistant
server (if not auto-detected) and long-lived access token.`
)

type fyneUI struct {
	app        fyne.App
	mainWindow fyne.Window
	text       *translations.Translator
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
	if dest, err := url.Parse(strings.TrimSpace(u)); err != nil {
		log.Warn().Err(err).
			Msgf("Unable parse url %s", u)
	} else {
		if err := ui.app.OpenURL(dest); err != nil {
			log.Warn().Err(err).
				Msgf("Unable to open url %s", dest.String())
		}
	}
}

func NewFyneUI(agent Agent, headless bool) *fyneUI {
	ui := &fyneUI{
		app:  app.NewWithID(agent.AppID()),
		text: translations.NewTranslator(),
	}
	ui.app.SetIcon(&TrayIcon{})
	if !headless {
		ui.mainWindow = ui.app.NewWindow(agent.AppName())
		ui.mainWindow.SetCloseIntercept(func() {
			ui.mainWindow.Hide()
		})
	}
	return ui
}

// DisplayTrayIcon displays an icon in the desktop tray with a menu for
// controlling the agent and showing other informational windows.
func (ui *fyneUI) DisplayTrayIcon(ctx context.Context, agent Agent) {
	if desk, ok := ui.app.(desktop.App); ok {
		menuItemQuit := fyne.NewMenuItem(ui.text.Translate("Quit"), func() {
			agent.Stop()
		})
		menuItemQuit.IsQuit = true
		menu := fyne.NewMenu("Main",
			fyne.NewMenuItem(ui.text.Translate("About"),
				func() {
					w := ui.aboutWindow(ctx, agent, ui.text)
					if w != nil {
						w.Show()
					}
				}),
			fyne.NewMenuItem(ui.text.Translate("Report Issue"),
				func() {
					ui.openURL(issueURL)
				}),
			fyne.NewMenuItem(ui.text.Translate("Request Feature"),
				func() {
					ui.openURL(featureRequestURL)
				}),
			fyne.NewMenuItem(ui.text.Translate("Fyne Settings"),
				func() {
					w := ui.fyneSettingsWindow(ui.text)
					w.Show()
				}),
			fyne.NewMenuItem(ui.text.Translate("App Settings"),
				func() {
					w := ui.agentSettingsWindow(agent, ui.text)
					if w != nil {
						w.Show()
					}
				}),
			fyne.NewMenuItem(ui.text.Translate("Sensors"),
				func() {
					w := ui.sensorsWindow(agent, ui.text)
					if w != nil {
						w.Show()
					}
				}),
			menuItemQuit)
		desk.SetSystemTrayMenu(menu)
	}
}

// DisplayRegistrationWindow displays a UI to prompt the user for the details needed to
// complete registration. It will populate with any values that were already
// provided via the command-line.
func (ui *fyneUI) DisplayRegistrationWindow(ctx context.Context, agent Agent, done chan struct{}) {
	ui.mainWindow.SetTitle(ui.text.Translate("App Registration"))

	var allFormItems []*widget.FormItem

	allFormItems = append(allFormItems, ui.serverConfigItems(ctx, agent, ui.text)...)
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
		widget.NewLabel(ui.text.Translate(explainRegistration)),
		registrationForm,
	))

	ui.mainWindow.SetOnClosed(func() {
		log.Debug().Msg("Closed")
	})

	ui.mainWindow.Show()
}

// aboutWindow creates a window that will show some interesting information
// about the agent, such as version numbers.
func (ui *fyneUI) aboutWindow(ctx context.Context, agent Agent, t *translations.Translator) fyne.Window {
	var widgets []fyne.CanvasObject
	if hassConfig, err := hass.GetHassConfig(ctx); err != nil {
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

// fyneSettingsWindow creates a window that will show the Fyne settings for
// controlling the look and feel of other windows.
func (ui *fyneUI) fyneSettingsWindow(t *translations.Translator) fyne.Window {
	w := ui.app.NewWindow(t.Translate("Fyne Settings"))
	w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
	return w
}

// sensorsWindow creates a window that displays all of the sensors and their
// values that are currently tracked by the agent. Values are updated
// continuously.
func (ui *fyneUI) sensorsWindow(s Agent, t *translations.Translator) fyne.Window {
	sensors := s.SensorList()
	if sensors == nil {
		return nil
	}

	getValue := func(n string) string {
		if v, err := s.SensorValue(n); err == nil {
			var b strings.Builder
			fmt.Fprintf(&b, "%v", v.State())
			if v.Units() != "" {
				fmt.Fprintf(&b, " %s", v.Units())
			}
			return b.String()
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

// agentSettingsWindow creates a window for changing settings related to the
// agent functionality. Most of these settings will be optional.
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

// serverConfigItems generates a list of form item widgets for selecting a
// server to register the agent against
func (ui *fyneUI) serverConfigItems(ctx context.Context, agent Agent, t *translations.Translator) []*widget.FormItem {
	allServers := hass.FindServers(ctx)

	tokenEntry := configEntry(agent, config.PrefToken, "ASecretLongLivedToken", false)
	tokenEntry.Validator = validation.NewRegexp("[A-Za-z0-9_\\.]+", "Invalid token format")

	serverEntry := configEntry(agent, config.PrefHost, allServers[0], false)
	serverEntry.Validator = httpValidator()
	serverEntry.Disable()

	autoServerSelect := widget.NewSelect(allServers, func(s string) {
		serverEntry.SetText(s)
	})

	manualServerEntry := serverEntry
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

	items = append(items, widget.NewFormItem(t.Translate("Token"), tokenEntry),
		widget.NewFormItem(t.Translate("Auto-discovered Servers"), autoServerSelect),
		widget.NewFormItem(t.Translate("Use Custom Server?"), manualServerSelect),
		widget.NewFormItem(t.Translate("Manual Server Entry"), manualServerEntry))

	return items
}

// mqttConfigItems generates a list of for item widgets for configuring the
// agent to use an MQTT for pub/sub functionality
func (ui *fyneUI) mqttConfigItems(agent Agent, t *translations.Translator) []*widget.FormItem {
	serverEntry := configEntry(agent, config.PrefMQTTServer, "localhost:1883", false)
	serverEntry.Validator = hostPortValidator()
	serverEntry.Disable()

	topicEntry := configEntry(agent, config.PrefMQTTTopic, "homeassistant", false)
	topicEntry.Disable()

	userEntry := configEntry(agent, config.PrefMQTTUser, "", false)
	userEntry.Disable()
	passwordEntry := configEntry(agent, config.PrefMQTTPassword, "", true)
	passwordEntry.Disable()

	mqttEnabled := configCheck(agent, config.PrefMQTTEnabled, func(b bool) {
		switch b {
		case true:
			serverEntry.Enable()
			topicEntry.Enable()
			userEntry.Enable()
			passwordEntry.Enable()
			if err := agent.SetConfig("UseMQTT", true); err != nil {
				log.Warn().Err(err).Msg("Could not enable MQTT.")
			}
		case false:
			serverEntry.Disable()
			topicEntry.Disable()
			userEntry.Disable()
			passwordEntry.Disable()
			if err := agent.SetConfig("UseMQTT", false); err != nil {
				log.Warn().Err(err).Msg("Could not disable MQTT.")
			}
		}
	})

	var items []*widget.FormItem

	items = append(items, widget.NewFormItem(t.Translate("Use MQTT?"), mqttEnabled),
		widget.NewFormItem(t.Translate("MQTT Server"), serverEntry),
		widget.NewFormItem(t.Translate("MQTT Topic"), topicEntry),
		widget.NewFormItem(t.Translate("MQTT User"), userEntry),
		widget.NewFormItem(t.Translate("MQTT Password"), passwordEntry),
	)

	return items
}

// configEntry creates a form entry widget that is tied to the given config
// value of the given agent. When the value of the entry widget changes, the
// corresponding config value will be updated.
func configEntry(agent Agent, name, placeholder string, secret bool) *widget.Entry {
	var entry *widget.Entry
	if secret {
		entry = widget.NewPasswordEntry()
	} else {
		entry = widget.NewEntry()
	}
	entry.OnChanged = func(s string) {
		if err := agent.SetConfig(name, s); err != nil {
			log.Warn().Err(err).Msgf("Could not set config entry %s.", name)
		}
	}
	if err := agent.GetConfig(name, &entry.Text); err != nil {
		log.Warn().Err(err).Msgf("Could not get value of config entry %s. Using placeholder.", name)
		entry.SetText(placeholder)
	}
	return entry
}

// configCheck creates a form checkbox widget that is tied to the given config
// value of the given agent. When the value of the entry widget changes, the
// corresponding config value will be updated.
func configCheck(agent Agent, name string, checkFn func(bool)) *widget.Check {
	entry := widget.NewCheck("", checkFn)
	if err := agent.GetConfig(name, &entry.Checked); err != nil {
		log.Warn().Err(err).Msgf("Could not get value of config entry %s. Using placeholder.", name)
		entry.SetChecked(false)
	}
	return entry
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

// httpValidator is a custom fyne validator that will validate a string is a
// valid http/https URL
func httpValidator() fyne.StringValidator {
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

// hostPortValidator is a custom fyne validator that will validate a string is a
// valid hostname:port combination
func hostPortValidator() fyne.StringValidator {
	v := validator.New()
	return func(text string) error {
		if v.Var(text, "hostname_port") != nil {
			return errors.New("you need to specify a valid host:port combination")
		}
		if _, err := url.Parse(text); err != nil {
			return errors.New("string is invalid")
		}
		return nil
	}
}
