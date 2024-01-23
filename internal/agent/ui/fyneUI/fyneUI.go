// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package fyneui

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
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/agent/config"
	"github.com/joshuar/go-hass-agent/internal/agent/ui"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/translations"
)

const (
	explainRegistration = `To register the agent, please enter the relevant details for your Home Assistant
server (if not auto-detected) and long-lived access token.`
	restartNote = `Please restart the agent to use changed settings.`
)

type fyneUI struct {
	app  fyne.App
	text *translations.Translator
}

func (i *fyneUI) Run(doneCh chan struct{}) {
	if i.app == nil {
		log.Warn().Msg("No supported windowing environment. Will not run UI elements.")
		return
	}
	log.Trace().Msg("Starting Fyne UI loop.")
	go func() {
		<-doneCh
		i.app.Quit()
	}()
	i.app.Run()
}

func (i *fyneUI) DisplayNotification(title, message string) {
	if i.app == nil {
		return
	}
	i.app.SendNotification(&fyne.Notification{
		Title:   title,
		Content: message,
	})
}

// Translate takes the input string and outputs a translated string for the
// language under which the agent is running.
func (i *fyneUI) Translate(text string) string {
	return i.text.Translate(text)
}

func NewFyneUI(id string) *fyneUI {
	i := &fyneUI{
		app:  app.NewWithID(id),
		text: translations.NewTranslator(),
	}
	i.app.SetIcon(&ui.TrayIcon{})
	return i
}

// DisplayTrayIcon displays an icon in the desktop tray with a menu for
// controlling the agent and showing other informational windows.
func (i *fyneUI) DisplayTrayIcon(a ui.Agent, cfg config.Config, t ui.SensorTracker) {
	if desk, ok := i.app.(desktop.App); ok {
		// About menu item.
		menuItemAbout := fyne.NewMenuItem(i.Translate("About"),
			func() {
				i.aboutWindow().Show()
			})
		// Sensors menu item.
		menuItemSensors := fyne.NewMenuItem(i.Translate("Sensors"),
			func() {
				i.sensorsWindow(a, t).Show()
			})

		// Settings menu and submenu items.
		settingsMenu := fyne.NewMenuItem(i.Translate("Settings"), nil)
		settingsMenu.ChildMenu = fyne.NewMenu("",
			fyne.NewMenuItem(i.Translate("App"),
				func() {
					i.agentSettingsWindow(cfg).Show()
				}),
			fyne.NewMenuItem(i.text.Translate("Fyne"),
				func() {
					i.fyneSettingsWindow().Show()
				}),
		)
		// Quit menu item.
		menuItemQuit := fyne.NewMenuItem(i.Translate("Quit"), func() {
			a.Stop()
		})
		menuItemQuit.IsQuit = true

		menu := fyne.NewMenu("",
			menuItemAbout,
			menuItemSensors,
			settingsMenu,
			menuItemQuit)
		desk.SetSystemTrayMenu(menu)
	}
}

// DisplayRegistrationWindow displays a UI to prompt the user for the details needed to
// complete registration. It will populate with any values that were already
// provided via the command-line.
func (i *fyneUI) DisplayRegistrationWindow(ctx context.Context, server, token *string, done chan struct{}) {
	w := i.app.NewWindow(i.Translate("App Registration"))

	var allFormItems []*widget.FormItem

	allFormItems = append(allFormItems, i.registrationFields(ctx, server, token)...)
	registrationForm := widget.NewForm(allFormItems...)
	registrationForm.OnSubmit = func() {
		w.Close()
		close(done)
	}
	registrationForm.OnCancel = func() {
		log.Warn().Msg("Canceling registration.")
		close(done)
		w.Close()
		ctx.Done()
	}

	w.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabel(i.Translate(explainRegistration)),
		registrationForm,
	))
	log.Debug().Msg("Asking user for registration details.")
	w.Show()
}

// aboutWindow creates a window that will show some interesting information
// about the agent, such as version numbers.
func (i *fyneUI) aboutWindow() fyne.Window {
	c := container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle("Go Hass Agent "+config.AppVersion, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""), // balance the header on the tutorial screen we leave blank on this content
		container.NewHBox(
			widget.NewHyperlink("website", parseURL(ui.AppURL)),
			widget.NewLabel("-"),
			widget.NewHyperlink("request feature", parseURL(ui.FeatureRequestURL)),
			widget.NewLabel("-"),
			widget.NewHyperlink("report issue", parseURL(ui.IssueURL)),
		),
	))

	w := i.app.NewWindow(i.Translate("About"))
	w.SetContent(c)
	return w
}

// fyneSettingsWindow creates a window that will show the Fyne settings for
// controlling the look and feel of other windows.
func (i *fyneUI) fyneSettingsWindow() fyne.Window {
	w := i.app.NewWindow(i.Translate("Fyne Settings"))
	w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
	return w
}

// agentSettingsWindow creates a window for changing settings related to the
// agent functionality. Most of these settings will be optional.
func (i *fyneUI) agentSettingsWindow(cfg config.Config) fyne.Window {
	var allFormItems []*widget.FormItem

	// MQTT settings
	mqttSettings := config.LoadMQTTPrefs(cfg)
	allFormItems = append(allFormItems, i.mqttConfigItems(mqttSettings)...)

	w := i.app.NewWindow(i.Translate("App Settings"))
	settingsForm := widget.NewForm(allFormItems...)
	settingsForm.OnSubmit = func() {
		mqttSettings.Save(cfg)
		log.Debug().Msg("Saved settings.")
	}
	settingsForm.OnCancel = func() {
		w.Close()
		log.Debug().Msg("No settings saved.")
	}
	settingsForm.SubmitText = i.Translate("Save")
	w.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabelWithStyle(i.Translate(restartNote), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		settingsForm,
	))
	return w
}

// sensorsWindow creates a window that displays all of the sensors and their
// values that are currently tracked by the agent. Values are updated
// continuously.
func (i *fyneUI) sensorsWindow(a ui.Agent, t ui.SensorTracker) fyne.Window {
	sensors := t.SensorList()
	if sensors == nil {
		return nil
	}

	getValue := func(n string) string {
		if v, err := t.Get(n); err == nil {
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
	w := i.app.NewWindow(i.Translate("Sensors"))
	w.SetContent(sensorsTable)
	w.Resize(fyne.NewSize(480, 640))
	w.SetOnClosed(func() {
		close(doneCh)
	})
	return w
}

// registrationFields generates a list of form item widgets for selecting a
// server to register the agent against.
func (i *fyneUI) registrationFields(ctx context.Context, server, token *string) []*widget.FormItem {
	allServers := hass.FindServers(ctx)

	if *token == "" {
		*token = "ASecretLongLivedToken"
	}
	tokenEntry := configEntry(token, false)
	tokenEntry.Validator = validation.NewRegexp("[A-Za-z0-9_\\.]+", "Invalid token format")

	if *server == "" {
		*server = allServers[0]
	}
	serverEntry := configEntry(server, false)
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

	items = append(items, widget.NewFormItem(i.Translate("Token"), tokenEntry),
		widget.NewFormItem(i.Translate("Auto-discovered Servers"), autoServerSelect),
		widget.NewFormItem(i.Translate("Use Custom Server?"), manualServerSelect),
		widget.NewFormItem(i.Translate("Manual Server Entry"), manualServerEntry))

	return items
}

// mqttConfigItems generates a list of for item widgets for configuring the
// agent to use an MQTT for pub/sub functionality.
func (i *fyneUI) mqttConfigItems(m *config.MQTTPrefs) []*widget.FormItem {
	serverEntry := configEntry(&m.Server, false)
	serverEntry.Validator = httpValidator()
	serverEntry.Disable()

	userEntry := configEntry(&m.User, false)
	userEntry.Disable()

	passwordEntry := configEntry(&m.Password, true)
	passwordEntry.Disable()

	mqttEnabled := configCheck(&m.Enabled, func(b bool) {
		switch b {
		case true:
			serverEntry.Enable()
			userEntry.Enable()
			passwordEntry.Enable()
			m.Enabled = true
		case false:
			serverEntry.Disable()
			userEntry.Disable()
			passwordEntry.Disable()
			m.Enabled = false
		}
	})

	var items []*widget.FormItem

	items = append(items, widget.NewFormItem(i.Translate("Use MQTT?"), mqttEnabled),
		widget.NewFormItem(i.Translate("MQTT Server"), serverEntry),
		widget.NewFormItem(i.Translate("MQTT User"), userEntry),
		widget.NewFormItem(i.Translate("MQTT Password"), passwordEntry),
	)

	return items
}

// configEntry creates a form entry widget that is tied to the given config
// value of the given agent. When the value of the entry widget changes, the
// corresponding config value will be updated.
func configEntry(value *string, secret bool) *widget.Entry {
	boundEntry := binding.BindString(value)
	entryWidget := widget.NewEntryWithData(boundEntry)
	if secret {
		entryWidget.Password = true
	}
	return entryWidget
}

// configCheck creates a form checkbox widget that is tied to the given config
// value of the given agent. When the value of the entry widget changes, the
// corresponding config value will be updated.
func configCheck(value *bool, checkFn func(bool)) *widget.Check {
	entry := widget.NewCheck("", checkFn)
	entry.SetChecked(*value)
	return entry
}

// longestString returns the longest string of a slice of strings. This can be
// used as a placeholder in Fyne containers to ensure there is enough space to
// display any of the strings in the slice.
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
// valid http/https URL.
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
// valid hostname:port combination.
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

// parseURL takes a URL as a string and parses it as a url.URL.
func parseURL(u string) *url.URL {
	dest, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		log.Warn().Err(err).
			Msgf("Unable parse url %s", u)
	}
	return dest
}
