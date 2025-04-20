// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	slogctx "github.com/veqryn/slog-context"

	agentvalidator "github.com/joshuar/go-hass-agent/internal/components/validation"
	"github.com/joshuar/go-hass-agent/internal/hass/discovery"
	"github.com/joshuar/go-hass-agent/internal/models"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
)

var (
	//nolint:stylecheck
	//lint:ignore ST1005 these are not standard error messages
	ErrInvalidURL = errors.New(InvalidURLMsgString)
	//nolint:stylecheck
	//lint:ignore ST1005 these are not standard error messages
	ErrInvalidURI = errors.New(InvalidURIMsgString)
	//nolint:stylecheck
	//lint:ignore ST1005 these are not standard error messages
	ErrInvalidHostPort = errors.New(InvalidHostPortMsgString)
)

const (
	RegistrationInfoString = `To register the agent, please enter the relevant details for your Home Assistant
server (if not auto-detected) and long-lived access token.`

	InvalidURLMsgString      = `You need to specify a valid http(s)://host:port.`
	InvalidURIMsgString      = `You need to specify a valid scheme://host:port.`
	InvalidHostPortMsgString = `You need to specify a valid host:port combination.`

	MQTTServerInfoString   = "Format should be scheme://host:port Where 'scheme' is one of 'tcp' or 'ssl', 'host' is the ip-address (or hostname) and 'port' is the port on which the broker is accepting connections."
	MQTTUserInfoString     = "Optional username to authenticate with the broker."
	MQTTPasswordInfoString = "Optional password to authenticate with the broker."

	PrefsRestartMsgString = `Please restart the agent to use changed preferences.`
)

// Notification represents the methods for displaying a notification.
type Notification interface {
	GetTitle() string
	GetMessage() string
}

// Hass provides methods for retrieving data from Home Assistant that the UI
// needs to display.
type Hass interface {
	GetHAVersion() string
	GetSensorList() []models.UniqueID
	GetSensor(id models.UniqueID) (*models.Sensor, error)
}

// FyneUI contains the data and methods to manage the UI state.
type FyneUI struct {
	app    fyne.App
	logger *slog.Logger
	hass   Hass
}

// New FyneUI sets up the UI for the agent.
func NewFyneUI(ctx context.Context, hass Hass) *FyneUI {
	appUI := &FyneUI{
		app:  app.NewWithID(preferences.AppName),
		hass: hass,
	}
	appUI.app.SetIcon(&trayIcon{})

	return appUI
}

// Run is the "main loop" of the UI.
func (i *FyneUI) Run(ctx context.Context) {
	// Stop the UI if the agent done signal is received.
	go func() {
		<-ctx.Done()
		i.app.Quit()
	}()

	// Run the UI (blocking).
	i.app.Run()
}

// DisplayNotification will display a notification using Fyne.
func (i *FyneUI) DisplayNotification(notification Notification) {
	i.app.SendNotification(&fyne.Notification{
		Title:   notification.GetTitle(),
		Content: notification.GetMessage(),
	})
}

// DisplayTrayIcon displays an icon in the desktop tray with a menu for
// controlling the agent and showing other informational windows.
func (i *FyneUI) DisplayTrayIcon(ctx context.Context, cancelFunc context.CancelFunc) {
	if desk, ok := i.app.(desktop.App); ok {
		// About menu item.
		menuItemAbout := fyne.NewMenuItem("About",
			func() {
				i.aboutWindow().Show()
			})
		// Sensors menu item.
		menuItemSensors := fyne.NewMenuItem("Sensors",
			func() {
				i.sensorsWindow().Show()
			})
		// Preferences/Settings items.
		menuItemAppPrefs := fyne.NewMenuItem("App Settings",
			func() {
				i.agentSettingsWindow().Show()
			})
		menuItemFynePrefs := fyne.NewMenuItem("Fyne Settings",
			func() {
				i.fyneSettingsWindow().Show()
			})
		// Quit menu item.
		menuItemQuit := fyne.NewMenuItem("Quit", func() {
			i.logger.Debug("Qutting agent on user request.")
			cancelFunc()
		})
		menuItemQuit.IsQuit = true

		menu := fyne.NewMenu("",
			menuItemAbout,
			menuItemSensors,
			menuItemAppPrefs,
			menuItemFynePrefs,
			menuItemQuit)
		desk.SetSystemTrayMenu(menu)
	}

	go func() {
		<-ctx.Done()
		i.app.Quit()
	}()
}

// DisplayRegistrationWindow displays a UI to prompt the user for the details needed to
// complete registration. It will populate with any values that were already
// provided via the command-line.
func (i *FyneUI) DisplayRegistrationWindow(ctx context.Context, prefs *preferences.Registration) chan bool {
	userCancelled := make(chan bool)

	if i.app == nil {
		slogctx.FromCtx(ctx).Warn("No UI available.")
		close(userCancelled)

		return userCancelled
	}

	window := i.app.NewWindow("App Registration")

	var allFormItems []*widget.FormItem
	allFormItems = append(allFormItems, i.registrationFields(prefs)...)
	registrationForm := widget.NewForm(allFormItems...)
	registrationForm.OnSubmit = func() {
		window.Close()
		userCancelled <- false
		close(userCancelled)
	}
	registrationForm.OnCancel = func() {
		i.logger.Warn("Canceling registration on user request.")
		userCancelled <- true
		close(userCancelled)
		window.Close()
	}

	windowContents := container.NewVBox(
		widget.NewLabelWithStyle(RegistrationInfoString, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""),
		registrationForm,
	)

	// If we get a message on the (agent's) doneCh, close the window if open.
	go func() {
		<-ctx.Done()
		window.Close()
	}()

	window.SetContent(windowContents)
	window.Show()

	return userCancelled
}

// aboutWindow creates a window that will show some interesting information
// about the agent, such as version numbers.
func (i *FyneUI) aboutWindow() fyne.Window {
	var widgets []fyne.CanvasObject

	icon := canvas.NewImageFromResource(&trayIcon{})
	icon.FillMode = canvas.ImageFillOriginal

	widgets = append(widgets, icon,
		widget.NewLabelWithStyle("Go Hass Agent "+preferences.AppVersion(),
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true}))

	widgets = append(widgets,
		widget.NewLabelWithStyle("Home Assistant "+i.hass.GetHAVersion(),
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true}),
	)

	widgets = append(widgets,
		widget.NewLabelWithStyle("Tracking "+strconv.Itoa(len(i.hass.GetSensorList()))+" Entities",
			fyne.TextAlignCenter,
			fyne.TextStyle{Italic: true}),
	)

	linkWidgets := generateLinks()
	widgets = append(widgets,
		widget.NewLabel(""),
		container.NewGridWithColumns(len(linkWidgets), linkWidgets...),
	)
	windowContents := container.NewCenter(container.NewVBox(widgets...))
	window := i.app.NewWindow("About")
	window.SetContent(windowContents)

	return window
}

// fyneSettingsWindow creates a window that will show the Fyne settings for
// controlling the look and feel of other windows.
func (i *FyneUI) fyneSettingsWindow() fyne.Window {
	window := i.app.NewWindow("Fyne Preferences")
	window.SetContent(settings.NewSettings().LoadAppearanceScreen(window))

	return window
}

// agentSettingsWindow creates a window for changing settings related to the
// agent functionality. Most of these settings will be optional.
func (i *FyneUI) agentSettingsWindow() fyne.Window {
	var allFormItems []*widget.FormItem

	mqttPrefs := preferences.MQTT()
	// Generate a form of MQTT preferences.
	allFormItems = append(allFormItems, mqttConfigItems(mqttPrefs)...)

	window := i.app.NewWindow("App Preferences")
	settingsForm := widget.NewForm(allFormItems...)
	settingsForm.OnSubmit = func() {
		// Set the new MQTT preferences.
		err := preferences.Set(
			preferences.SetMQTTEnabled(mqttPrefs.MQTTEnabled),
			preferences.SetMQTTServer(mqttPrefs.MQTTServer),
			preferences.SetMQTTUser(mqttPrefs.MQTTUser),
			preferences.SetMQTTPassword(mqttPrefs.MQTTPassword),
		)
		// Save the new MQTT preferences to file.
		if err != nil {
			dialog.ShowError(err, window)
			i.logger.Error("Could note save preferences.", slog.Any("error", err))
		} else {
			dialog.ShowInformation("Saved", "MQTT Preferences have been saved. Restart agent to utilize them.", window)
			i.logger.Info("Saved MQTT preferences.")
		}
	}
	settingsForm.OnCancel = func() {
		window.Close()
	}
	settingsForm.SubmitText = "Save"
	window.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabelWithStyle(PrefsRestartMsgString, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		settingsForm,
	))

	return window
}

// sensorsWindow creates a window that displays all of the sensors and their
// values that are currently tracked by the agent. Values are updated
// continuously.
//
//nolint:gocognit
func (i *FyneUI) sensorsWindow() fyne.Window {
	sensors := i.hass.GetSensorList()
	if sensors == nil {
		return nil
	}

	getValue := func(n string) string {
		if details, err := i.hass.GetSensor(n); err == nil {
			var valueStr strings.Builder

			fmt.Fprintf(&valueStr, "%v", details.State)

			if details.UnitOfMeasurement != nil {
				fmt.Fprintf(&valueStr, " %s", *details.UnitOfMeasurement)
			}

			return valueStr.String()
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
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label, ok := obj.(*widget.Label)
			if !ok {
				return
			}

			switch id.Col {
			case 0:
				label.SetText(sensors[id.Row])
			case 1:
				label.SetText(getValue(sensors[id.Row]))
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
	// ?: this is clunky. better way would be use Fyne bindings to sensor values
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

	window := i.app.NewWindow("Sensors")
	window.SetContent(sensorsTable)
	window.Resize(fyne.NewSize(480, 640))
	window.SetOnClosed(func() {
		close(doneCh)
	})

	return window
}

// registrationFields generates a list of form item widgets for selecting a
// server to register the agent against.
func (i *FyneUI) registrationFields(prefs *preferences.Registration) []*widget.FormItem {
	searchCtx, searchCancelFunc := context.WithCancel(context.TODO())
	defer searchCancelFunc()

	var allServers []string

	foundServers, err := discovery.FindServers(searchCtx)
	if err != nil {
		i.logger.Warn("Errors occurred discovering Home Assistant servers.", slog.Any("error", err))
	}

	allServers = append(allServers, foundServers...)

	tokenEntry := configEntry(&prefs.Token, false)
	tokenEntry.Validator = validation.NewRegexp("[A-Za-z0-9_\\.]+", "Invalid token format")

	serverEntry := configEntry(&prefs.Server, false)
	serverEntry.Validator = httpValidator()
	serverEntry.Disable()

	autoServerSelect := widget.NewSelect(allServers, func(s string) {
		serverEntry.SetText(s)
	})
	autoServerSelect.SetSelectedIndex(0)

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

	ignoreURLsSelect := widget.NewCheck("", func(b bool) {
		switch b {
		case true:
			prefs.IgnoreHassURLs = true
		case false:
			prefs.IgnoreHassURLs = false
		}
	})

	return []*widget.FormItem{
		{
			Text:     "Token",
			HintText: "The long-lived access token generated in Home Assistant.",
			Widget:   tokenEntry,
		},
		{
			Text:     "Auto-discovered Servers",
			HintText: "These are the Home Assistant servers that were detected on the local network.",
			Widget:   autoServerSelect,
		},
		{
			Text:     "Use Custom Server?",
			HintText: "Select this option to enter a server manually below.", Widget: manualServerSelect,
		},
		{Text: "Manual Server Entry", Widget: manualServerEntry},
		{
			Text:     "Ignore returned URLs?",
			HintText: "Override Home Assistant and use server chosen (above) for API access.",
			Widget:   ignoreURLsSelect,
		},
	}
}

// mqttConfigItems generates a list of for item widgets for configuring the
// agent to use an MQTT for pub/sub functionality.
func mqttConfigItems(prefs *preferences.MQTTPreferences) []*widget.FormItem {
	serverEntry := configEntry(&prefs.MQTTServer, false)
	serverEntry.Validator = uriValidator()
	serverEntry.Disable()
	serverFormItem := widget.NewFormItem("MQTT Server", serverEntry)
	serverFormItem.HintText = MQTTServerInfoString

	userEntry := configEntry(&prefs.MQTTUser, false)
	userEntry.Disable()
	userFormItem := widget.NewFormItem("MQTT User", userEntry)
	userFormItem.HintText = MQTTUserInfoString

	passwordEntry := configEntry(&prefs.MQTTPassword, true)
	passwordEntry.Disable()
	passwordFormItem := widget.NewFormItem("MQTT Password", passwordEntry)
	passwordFormItem.HintText = MQTTPasswordInfoString

	mqttEnabled := configCheck(&prefs.MQTTEnabled, func(b bool) {
		switch b {
		case true:
			serverEntry.Enable()
			userEntry.Enable()
			passwordEntry.Enable()

			prefs.MQTTEnabled = true
		case false:
			serverEntry.Disable()
			userEntry.Disable()
			passwordEntry.Disable()

			prefs.MQTTEnabled = false
		}
	})

	var items []*widget.FormItem

	items = append(items, widget.NewFormItem("Use MQTT?", mqttEnabled),
		serverFormItem,
		userFormItem,
		passwordFormItem,
	)

	return items
}

// configEntry creates a form entry widget that is tied to the given config
// value of the given agent. When the value of the entry widget changes, the
// corresponding config value will be updated.
//
//revive:disable:flag-parameter
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
func longestString(stringList []string) string {
	var longString string
	if len(stringList) > 0 {
		longString = stringList[0]
		stringList = stringList[1:]
	}

	for _, s := range stringList {
		if len(longString) <= len(s) {
			longString = s
		}
	}

	return longString
}

// httpValidator is a custom fyne validator that will validate a string is a
// valid http/https URL.
func httpValidator() fyne.StringValidator {
	return func(text string) error {
		if agentvalidator.Validate.Var(text, "http_url") != nil {
			return ErrInvalidURL
		}

		if _, err := url.Parse(text); err != nil {
			return ErrInvalidURL
		}

		return nil
	}
}

// uriValidator is a custom fyne validator that will validate a string is a
// valid http/https URL.
func uriValidator() fyne.StringValidator {
	return func(text string) error {
		if agentvalidator.Validate.Var(text, "uri") != nil {
			return ErrInvalidURI
		}

		if _, err := url.Parse(text); err != nil {
			return ErrInvalidURI
		}

		return nil
	}
}

// hostPortValidator is a custom fyne validator that will validate a string is a
// valid hostname:port combination.
//
//nolint:err113
//lint:ignore U1000 keeping this around just in case
func hostPortValidator(msg string) fyne.StringValidator {
	var errMsg error
	if msg != "" {
		errMsg = errors.New(msg)
	} else {
		errMsg = ErrInvalidHostPort
	}

	return func(text string) error {
		if agentvalidator.Validate.Var(text, "hostname_port") != nil {
			return errMsg
		}
		// if _, err := url.Parse(text); err != nil {
		// 	return errors.New("string is invalid")
		// }
		return nil
	}
}

// parseURL takes a URL as a string and parses it as a url.URL.
func parseURL(u string) (*url.URL, error) {
	dest, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL: %w", err)
	}

	return dest, nil
}

func generateLinks() []fyne.CanvasObject {
	var (
		link *url.URL
		err  error
	)

	widgets := make([]fyne.CanvasObject, 0, 3) //nolint:mnd

	link, err = parseURL(preferences.AppURL)
	if err != nil {
		slog.Warn("Unable to parse app URL.", slog.Any("error", err))
	} else {
		widgets = append(widgets, widget.NewHyperlink("website", link))
	}

	link, err = parseURL(preferences.FeatureRequestURL)
	if err != nil {
		slog.Warn("Unable to parse feature request URL.", slog.Any("error", err))
	} else {
		widgets = append(widgets, widget.NewHyperlink("request feature", link))
	}

	link, err = parseURL(preferences.IssueURL)
	if err != nil {
		slog.Warn("Unable to parse issues URL.", slog.Any("error", err))
	} else {
		widgets = append(widgets, widget.NewHyperlink("report issue", link))
	}

	return widgets
}
