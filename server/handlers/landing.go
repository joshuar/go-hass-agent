// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package handlers

import (
	"net/http"

	"github.com/justinas/alice"

	"github.com/joshuar/go-hass-agent/agent"
)

func Landing(agent *agent.Agent) http.HandlerFunc {
	return alice.New(
		routeLogger,
	).ThenFunc(func(res http.ResponseWriter, req *http.Request) {
		if agent.IsRegistered() {
			res.WriteHeader(http.StatusNotImplemented)
		} else {
			http.Redirect(res, req, "/register", http.StatusTemporaryRedirect)
		}
	}).ServeHTTP
}
