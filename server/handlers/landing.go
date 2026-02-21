// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package handlers

import (
	"net/http"

	"github.com/justinas/alice"

	"github.com/joshuar/go-hass-agent/agent"
	"github.com/joshuar/go-hass-agent/hass"
	"github.com/joshuar/go-hass-agent/web/templates"
)

func Landing(agent *agent.Agent) http.HandlerFunc {
	return alice.New(
		routeLogger,
	).ThenFunc(func(res http.ResponseWriter, req *http.Request) {
		if agent.IsRegistered() {
			hassclient, err := hass.NewClient(req.Context(), agent)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
			renderPage(templates.Landing(agent, hassclient), "Go Hass Agent").ServeHTTP(res, req)
		} else {
			http.Redirect(res, req, "/register", http.StatusTemporaryRedirect)
		}
	}).ServeHTTP
}
