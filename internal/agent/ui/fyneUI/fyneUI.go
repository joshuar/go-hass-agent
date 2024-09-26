// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:comment-spacings
package fyneui

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
	"github.com/go-playground/validator/v10"

	"github.com/joshuar/go-hass-agent/internal/agent/ui"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/internal/translations"
)

var validate *validator.Validate

var (
	//nolint:stylecheck
	//lint:ignore ST1005 these are not standard error messages
	ErrInvalidURL = errors.New(ui.InvalidURLMsgString)
	//nolint:stylecheck
	//lint:ignore ST1005 these are not standard error messages
	ErrInvalidURI = errors.New(ui.InvalidURIMsgString)
	//nolint:stylecheck
	//lint:ignore ST1005 these are not standard error messages
	ErrInvalidHostPort = errors.New(ui.InvalidHostPortMsgString)
)

type Notification interface {
	GetTitle() string
	GetMessage() string
}

type FyneUI struct {
	app    fyne.App
	text   *translations.Translator
	logger *slog.Logger
}

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// New FyneUI sets up the UI for the agent.
func NewFyneUI(ctx context.Context) *FyneUI {
	appUI := &FyneUI{
		app:    app.NewWithID(preferences.AppName),
		text:   translations.NewTranslator(ctx),
		logger: logging.FromContext(ctx).With(slog.String("subsystem", "fyne")),
	}
	appUI.app.SetIcon(&ui.TrayIcon{})

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

// Translate takes the input string and outputs a translated string for the
// language under which the agent is running.
func (i *FyneUI) Translate(text string) string {
	return i.text.Translate(text)
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
		menuItemAbout := fyne.NewMenuItem(i.Translate("About"),
			func() {
				i.aboutWindow(ctx).Show()
			})
		// Sensors menu item.
		menuItemSensors := fyne.NewMenuItem(i.Translate("Sensors"),
			func() {
				i.sensorsWindow().Show()
			})
		// Preferences/Settings items.
		menuItemAppPrefs := fyne.NewMenuItem(i.Translate("App Settings"),
			func() {
				i.agentSettingsWindow(ctx).Show()
			})
		menuItemFynePrefs := fyne.NewMenuItem(i.text.Translate("Fyne Settings"),
			func() {
				i.fyneSettingsWindow().Show()
			})
		// Quit menu item.
		menuItemQuit := fyne.NewMenuItem(i.Translate("Quit"), func() {
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
		logging.FromContext(ctx).Warn("No UI available.")
		close(userCancelled)

		return userCancelled
	}

	window := i.app.NewWindow(i.Translate("App Registration"))

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
		widget.NewLabelWithStyle(i.Translate(ui.RegistrationInfoString), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
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
func (i *FyneUI) aboutWindow(ctx context.Context) fyne.Window {
	var widgets []fyne.CanvasObject

	prefs, err := preferences.Load(ctx)
	if err != nil {
		logging.FromContext(ctx).Error("Could not start sensor controller.", slog.Any("error", err))
		return nil
	}

	hassclient, err := hass.NewClient(ctx)
	if err != nil {
		logging.FromContext(ctx).Debug("Cannot create Home Assistant client.", slog.Any("error", err))
		return nil
	}

	hassclient.Endpoint(prefs.RestAPIURL(), hass.DefaultTimeout)

	icon := canvas.NewImageFromResource(&ui.TrayIcon{})
	icon.FillMode = canvas.ImageFillOriginal

	widgets = append(widgets, icon,
		widget.NewLabelWithStyle("Go Hass Agent "+preferences.AppVersion,
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true}))

	widgets = append(widgets,
		widget.NewLabelWithStyle("Home Assistant "+hassclient.HassVersion(ctx),
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Tracking "+strconv.Itoa(len(hass.SensorList()))+" Entities",
			fyne.TextAlignCenter,
			fyne.TextStyle{Italic: true}),
	)

	linkWidgets := generateLinks()
	widgets = append(widgets,
		widget.NewLabel(""),
		container.NewGridWithColumns(len(linkWidgets), linkWidgets...),
	)
	windowContents := container.NewCenter(container.NewVBox(widgets...))
	window := i.app.NewWindow(i.Translate("About"))
	window.SetContent(windowContents)

	return window
}

// fyneSettingsWindow creates a window that will show the Fyne settings for
// controlling the look and feel of other windows.
func (i *FyneUI) fyneSettingsWindow() fyne.Window {
	window := i.app.NewWindow(i.Translate("Fyne Preferences"))
	window.SetContent(settings.NewSettings().LoadAppearanceScreen(window))

	return window
}

// agentSettingsWindow creates a window for changing settings related to the
// agent functionality. Most of these settings will be optional.
func (i *FyneUI) agentSettingsWindow(ctx context.Context) fyne.Window {
	var allFormItems []*widget.FormItem

	prefs, err := preferences.Load(ctx)
	if err != nil {
		logging.FromContext(ctx).Error("Could not start sensor controller.", slog.Any("error", err))
		return nil
	}

	// Generate a form of MQTT preferences.
	allFormItems = append(allFormItems, i.mqttConfigItems(prefs.MQTT)...)

	window := i.app.NewWindow(i.Translate("App Preferences"))
	settingsForm := widget.NewForm(allFormItems...)
	settingsForm.OnSubmit = func() {
		// Save the new MQTT preferences to file.
		if err := prefs.Save(); err != nil {
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
	settingsForm.SubmitText = i.Translate("Save")
	window.SetContent(container.New(layout.NewVBoxLayout(),
		widget.NewLabelWithStyle(i.Translate(ui.PrefsRestartMsgString), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
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
	sensors := hass.SensorList()
	if sensors == nil {
		return nil
	}

	getValue := func(n string) string {
		if sensor, err := hass.GetSensor(n); err == nil {
			var valueStr strings.Builder

			fmt.Fprintf(&valueStr, "%v", sensor.State)

			if sensor.Units != "" {
				fmt.Fprintf(&valueStr, " %s", sensor.Units)
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

	window := i.app.NewWindow(i.Translate("Sensors"))
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

	foundServers, err := hass.FindServers(searchCtx)
	if err != nil {
		i.logger.Warn("Errors occurred discovering Home Assistant servers.", slog.Any("error", err))
	}

	allServers = append(allServers, prefs.Server)
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
			Text:     i.Translate("Token"),
			HintText: i.Translate("The long-lived access token generated in Home Assistant."),
			Widget:   tokenEntry,
		},
		{
			Text:     i.Translate("Auto-discovered Servers"),
			HintText: i.Translate("These are the Home Assistant servers that were detected on the local network."),
			Widget:   autoServerSelect,
		},
		{
			Text:     i.Translate("Use Custom Server?"),
			HintText: i.Translate("Select this option to enter a server manually below."), Widget: manualServerSelect,
		},
		{Text: i.Translate("Manual Server Entry"), Widget: manualServerEntry},
		{
			Text:     i.Translate("Ignore returned URLs?"),
			HintText: i.Translate("Override Home Assistant and use server chosen (above) for API access."),
			Widget:   ignoreURLsSelect,
		},
	}
}

// mqttConfigItems generates a list of for item widgets for configuring the
// agent to use an MQTT for pub/sub functionality.
func (i *FyneUI) mqttConfigItems(prefs *preferences.MQTT) []*widget.FormItem {
	serverEntry := configEntry(&prefs.MQTTServer, false)
	serverEntry.Validator = uriValidator()
	serverEntry.Disable()
	serverFormItem := widget.NewFormItem(i.Translate("MQTT Server"), serverEntry)
	serverFormItem.HintText = ui.MQTTServerInfoString

	userEntry := configEntry(&prefs.MQTTUser, false)
	userEntry.Disable()
	userFormItem := widget.NewFormItem(i.Translate("MQTT User"), userEntry)
	userFormItem.HintText = ui.MQTTUserInfoString

	passwordEntry := configEntry(&prefs.MQTTPassword, true)
	passwordEntry.Disable()
	passwordFormItem := widget.NewFormItem(i.Translate("MQTT Password"), passwordEntry)
	passwordFormItem.HintText = ui.MQTTPasswordInfoString

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

	items = append(items, widget.NewFormItem(i.Translate("Use MQTT?"), mqttEnabled),
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
	v := validator.New()

	return func(text string) error {
		if v.Var(text, "http_url") != nil {
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
		if validate.Var(text, "uri") != nil {
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

	v := validator.New()

	return func(text string) error {
		if v.Var(text, "hostname_port") != nil {
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
