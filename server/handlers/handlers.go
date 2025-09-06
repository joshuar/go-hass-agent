// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogctx "github.com/veqryn/slog-context"
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
