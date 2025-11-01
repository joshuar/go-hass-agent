// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/justinas/alice"

	"github.com/joshuar/go-hass-agent/agent/workers/mqtt"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/server/forms"
	"github.com/joshuar/go-hass-agent/web/templates"
)

// ShowPreferences handles showing a form for editing the agent preferences.
func ShowPreferences() http.HandlerFunc {
	return alice.New(
		routeLogger,
	).ThenFunc(func(res http.ResponseWriter, req *http.Request) {
		prefs := templates.NewPreferences()
		err := config.Load("mqtt", prefs.MQTT)
		if err != nil {
			template := templ.Join(
				templates.PreferencesForm(prefs),
				templates.Notification(models.NewErrorMessage("Error retrieving preferences.", err.Error())))
			renderPartial(template).ServeHTTP(res, req)
		}
		renderPage(templates.PreferencesForm(prefs), "Preferences - Go Hass Agent").ServeHTTP(res, req)
	}).ServeHTTP
}

// SavePreferences handles extracting the new preferences from the request and saving them to the configuration file.
func SaveMQTTPreferences() http.HandlerFunc {
	return alice.New(
		routeLogger,
	).ThenFunc(func(res http.ResponseWriter, req *http.Request) {
		prefs, valid, err := forms.DecodeForm[*mqtt.Config](req)
		if err != nil || !valid {
			template := templ.Join(
				templates.PreferencesForm(&templates.Preferences{MQTT: prefs}),
				templates.Notification(models.NewErrorMessage("Invalid details.", err.Error())))
			renderPartial(template).ServeHTTP(res, req)
			return
		}
		err = config.Save(mqtt.ConfigPrefix, prefs)
		if err != nil {
			template := templ.Join(
				templates.PreferencesForm(&templates.Preferences{MQTT: prefs}),
				templates.Notification(models.NewErrorMessage("Failed to save preferences.", err.Error())))
			renderPartial(template).ServeHTTP(res, req)
		}
		template := templ.Join(
			templates.PreferencesForm(&templates.Preferences{MQTT: prefs}),
			templates.Notification(
				models.NewSuccessMessage("Preferences saved.", "Remember to restart the agent to use the new settings.")))
		renderPartial(template).ServeHTTP(res, req)
	}).ServeHTTP
}
