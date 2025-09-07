// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/grandcat/zeroconf"
	"github.com/justinas/alice"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/agent"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/hass"
	"github.com/joshuar/go-hass-agent/logging"
	"github.com/joshuar/go-hass-agent/models"
	"github.com/joshuar/go-hass-agent/server/forms"
	"github.com/joshuar/go-hass-agent/web/templates"
)

func GetRegistration(agent *agent.Agent) http.HandlerFunc {
	return alice.New(
		routeLogger,
	).ThenFunc(func(res http.ResponseWriter, req *http.Request) {
		if agent.IsRegistered() {
			renderTemplate(templates.RegistrationResponse(models.NewInfoMessage("Already registered", "")), "Register - Go Hass Agent").ServeHTTP(res, req)
		} else {
			renderTemplate(templates.RegistrationForm(nil, nil), "Register - Go Hass Agent").ServeHTTP(res, req)
		}
	}).ServeHTTP
}

func RegistrationDiscovery() http.HandlerFunc {
	return alice.New(
		routeLogger,
	).ThenFunc(func(res http.ResponseWriter, req *http.Request) {
		serverList := []string{config.DefaultServer}

		resolver, err := zeroconf.NewResolver(nil)
		if err != nil {
			templ.Handler(templates.DiscoveredServers(serverList)).ServeHTTP(res, req)
			return
		}

		entries := make(chan *zeroconf.ServiceEntry)

		go func(results <-chan *zeroconf.ServiceEntry) {
			for entry := range results {
				var server string

				for _, t := range entry.Text {
					if value, found := strings.CutPrefix(t, "base_url="); found {
						server = value
					}
				}

				if server != "" {
					serverList = append(serverList, server)
				} else {
					slogctx.FromCtx(req.Context()).Log(req.Context(), logging.LevelTrace,
						"Found a server malformed server, will not use.", slog.String("server", entry.HostName))
				}
			}
		}(entries)

		slogctx.FromCtx(req.Context()).Info("Looking for Home Assistant servers on the local network...")

		searchCtx, searchCancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer searchCancel()

		err = resolver.Browse(searchCtx, "_home-assistant._tcp", "local.", entries)
		if err != nil {
			slogctx.FromCtx(req.Context()).Error("Could not search for Home Assistant servers.",
				slog.Any("error", err),
			)
		}

		<-searchCtx.Done()
		templ.Handler(templates.DiscoveredServers(serverList)).ServeHTTP(res, req)
	}).ServeHTTP
}

func ProcessRegistration(agent *agent.Agent) http.HandlerFunc {
	return alice.New(
		routeLogger,
	).ThenFunc(func(res http.ResponseWriter, req *http.Request) {
		request, valid, err := forms.DecodeForm[*hass.RegistrationRequest](req)
		if err != nil || !valid {
			renderTemplate(templates.RegistrationForm(models.NewErrorMessage("Invalid details.", err.Error()), request), "Register - Go Hass Agent").ServeHTTP(res, req)
			return
		}

		err = hass.Register(req.Context(), agent.Config.ID, request)
		if err != nil {
			templ.Handler(templates.RegistrationForm(
				models.NewErrorMessage("There was a problem trying to register this device:", err.Error()),
				request,
			)).ServeHTTP(res, req)
			return
		}

		agent.Register()

		templ.Handler(templates.RegistrationForm(
			models.NewSuccessMessage("Device registered!", ""),
			request,
		)).ServeHTTP(res, req)
	}).ServeHTTP
}
