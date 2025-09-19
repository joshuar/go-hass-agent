// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package handlers

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/web/templates"
)

func StaticFileServerHandler(fs http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check, if the requested file is existing.
		_, err := fs.Open(r.URL.Path)
		if err != nil {
			// If file is not found, return HTTP 404 error.
			http.NotFound(w, r)
			return
		}
		// File is found, return to standard http.FileServer.
		http.FileServer(fs).ServeHTTP(w, r)
	})
}

// routeLogger decorates the logger in the request context with routing information.
func routeLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := slogctx.With(req.Context(),
			slog.String("route", chi.RouteContext(req.Context()).RoutePattern()),
			slog.String("method", req.Method),
		)
		ctx = slogctx.With(ctx, slog.Group("req", slog.String("id", middleware.GetReqID(ctx))))
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// renderPage will render the given template as a full page. It handles htmx and non-htmx requests, rendering the
// appropriate full or partial HTML response as appropriate.
func renderPage(template templ.Component, title string) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if template == nil {
			// If there is no response, return 204: No Content.
			res.WriteHeader(http.StatusNoContent)
			return
		}
		// Write the response template.
		if IsHTMX(req) {
			if IsHistoryRestoreRequest(req) {
				templ.Handler(templates.Page(title, template)).ServeHTTP(res, req)
				return
			} else if title != "" {
				// Update the page title if set.
				template = templ.Join(template, templates.SetPageTitle(title))
			}
			template = templ.Join(template, templates.UpdateCSRFToken())
			target := templates.FragmentKey(req.Header.Get("HX-Target"))
			if target == "" {
				target = templates.FragmentContent
			}
			templ.Handler(template, templ.WithFragments(target)).ServeHTTP(res, req)
		} else {
			template = templates.Page(title, template)
			err := template.Render(req.Context(), res)
			if err != nil {
				slogctx.FromCtx(req.Context()).Error("Failed to render page template.", slog.Any("error", err))
				http.Error(res, "Failed to render page template.", http.StatusInternalServerError)
				return
			}
		}
	})
}

// renderPartial will render the given template, optionally updating the page title if one is given.
func renderPartial(template templ.Component) http.Handler {
	return templ.Handler(templ.Join(template, templates.UpdateCSRFToken()))
}

func IsHTMX(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}

func IsHistoryRestoreRequest(req *http.Request) bool {
	return req.Header.Get("HX-History-Restore-Request") == "true"
}
