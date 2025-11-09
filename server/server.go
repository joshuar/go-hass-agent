// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	slogchi "github.com/samber/slog-chi"
	slogctx "github.com/veqryn/slog-context"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/joshuar/go-hass-agent/agent"
	"github.com/joshuar/go-hass-agent/config"
	"github.com/joshuar/go-hass-agent/server/handlers"
	"github.com/joshuar/go-hass-agent/server/middlewares"
	"github.com/joshuar/go-hass-agent/validation"
)

const (
	serverConfigPrefix = "server"
)

// Server holds data for implementing the web server component of the agent.
type Server struct {
	*http.Server

	static embed.FS
	Config *Config
}

// New creates a new server component for the agent.
func New(static embed.FS, agent *agent.Agent, options ...configOption) (*Server, error) {
	// Create server object with default config.
	server := &Server{
		static: static,
		Config: NewConfig(),
	}
	// Load the server config from file, overwriting default config.
	err := config.Load(serverConfigPrefix, server.Config)
	if err != nil {
		return server, fmt.Errorf("create web server: load config: %w", err)
	}
	// Overwrite config with any options passed on command-line.
	for option := range slices.Values(options) {
		option(server.Config)
	}

	err = validation.Validate.Struct(server.Config)
	if err != nil {
		return nil, fmt.Errorf("create web server: load config: %w", err)
	}

	// Set up routes.
	router := setupRoutes(static, agent)
	// Set up server object.
	h2s := &http2.Server{}
	server.Server = &http.Server{
		Handler:      h2c.NewHandler(nosurf.New(router), h2s),
		Addr:         net.JoinHostPort(server.Config.Host, server.Config.Port),
		ReadTimeout:  server.Config.ReadTimeout,
		WriteTimeout: server.Config.WriteTimeout,
		IdleTimeout:  server.Config.IdleTimeout,
	}

	return server, nil
}

// Start starts the web server component of the agent.
func (s *Server) Start(ctx context.Context) error {
	slogctx.FromCtx(ctx).Debug("Starting server...",
		slog.String("address", s.Addr))

	var wg sync.WaitGroup

	wg.Add(1)
	// Listen for shutdown events and process them.
	go func() {
		wg.Done()

		stop := make(chan os.Signal, 1)

		signal.Notify(stop, os.Interrupt)
		<-stop

		err := s.Shutdown(ctx)
		// Can't do much here except for logging any errors
		if err != nil {
			slogctx.FromCtx(ctx).Error("Error occurred when trying to shut down server.",
				slog.Any("error", err),
			)
		}
	}()

	// And we serve HTTP until the world ends.
	go func() {
		var err error
		if s.Config.CertFile != "" && s.Config.KeyFile != "" {
			err = s.ListenAndServeTLS(s.Config.CertFile, s.Config.KeyFile)
		} else {
			err = s.ListenAndServe()
		}
		if errors.Is(err, http.ErrServerClosed) { // graceful shutdown
			slogctx.FromCtx(ctx).Debug("Shutting down server...")
			wg.Wait()
		} else if err != nil {
			slogctx.FromCtx(ctx).Debug("Error shutting down server.",
				slog.Any("error", err))
		}
	}()

	return nil
}

func setupRoutes(static embed.FS, agent *agent.Agent) *chi.Mux {
	// Set up a new chi router.
	router := chi.NewRouter()
	// Health check endpoints (for GCP).
	router.Use(middleware.Heartbeat("/health-check"))
	// Middleware stack.
	router.Use(
		middleware.RequestID,
		middleware.Recoverer,
		slogchi.NewWithConfig(slog.Default(), slogchi.Config{
			ClientErrorLevel: slog.LevelWarn,
			ServerErrorLevel: slog.LevelError,
			WithRequestID:    true,
		}),
		middlewares.SetupHTMX,

		// middlewares.SetupCORS(config.Environment()),
		// middlewares.CSP(server.ServerConfig.CSP),
		// middlewares.Etag,
		middleware.StripSlashes,
		middlewares.SaveCSRFToken,
		middleware.NoCache,
	)
	// User endpoints.
	//
	// Static content.
	router.Group(func(r chi.Router) {
		r.Handle("/web/content/*", handlers.StaticFileServerHandler(http.FS(static)))
	})
	// // Error handling.
	// router.NotFound(handlers.NotFound())
	// Landing page.
	router.Get("/", handlers.Landing(agent))
	// Registration.
	router.Get("/register", handlers.GetRegistration(agent))
	router.With(middlewares.RequireHTMX).Get("/register/discovery", handlers.RegistrationDiscovery())
	router.With(middlewares.RequireHTMX).Post("/register", handlers.ProcessRegistration(agent))
	// Preferences.
	router.Get("/preferences", handlers.ShowPreferences())
	router.With(middlewares.RequireHTMX).Post("/preferences/mqtt", handlers.SaveMQTTPreferences())

	return router
}
