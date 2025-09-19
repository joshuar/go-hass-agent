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
	"strconv"
	"sync"
	"time"

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
)

const (
	serverConfigPrefix = "server"
)

// Agent represents the data and methods required for running the agent.
type Server struct {
	*http.Server

	agent  *agent.Agent
	static embed.FS
	Config *Config
}

// AgentConfig contains the agent configuration options.
type Config struct {
	// Host is the hostname to listen on.
	Host string `toml:"host"`
	// Port is the port to listen on.
	Port int `toml:"port"`
	// CertFile points to the file containing a server certificate.
	CertFile string `toml:"cert"`
	// KeyFile points to the file containing a server key.
	KeyFile      string        `toml:"key"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`
	IdleTimeout  time.Duration `toml:"idle_timeout"`
}

func New(static embed.FS, agent *agent.Agent) (*Server, error) {
	// Create server object with default config.
	server := &Server{
		Config: &Config{
			Host:         "localhost",
			Port:         7000,
			CertFile:     "",
			KeyFile:      "",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		static: static,
	}
	// Load the server config from file, overwriting default config.
	if err := config.Load(serverConfigPrefix, server.Config); err != nil {
		return server, fmt.Errorf("unable to load server config: %w", err)
	}

	// Set up routes.
	router := setupRoutes(static, agent)

	h2s := &http2.Server{}
	server.Server = &http.Server{
		Handler:      h2c.NewHandler(nosurf.New(router), h2s),
		Addr:         net.JoinHostPort(server.Config.Host, strconv.Itoa(server.Config.Port)),
		ReadTimeout:  server.Config.ReadTimeout,
		WriteTimeout: server.Config.WriteTimeout,
		IdleTimeout:  server.Config.IdleTimeout,
	}

	return server, nil
}

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

		err := s.Shutdown(context.Background())
		// Can't do much here except for logging any errors
		if err != nil {
			slogctx.FromCtx(ctx).Error("Error occurred when trying to shut down server.",
				slog.Any("error", err),
			)
		}
	}()

	// And we serve HTTP until the world ends.
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
		return fmt.Errorf("error shutting down server: %w", err)
	}

	return nil
}

func setupRoutes(static embed.FS, agent *agent.Agent) *chi.Mux {
	// Set up a new chi router.
	router := chi.NewRouter()
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

	// Routes.
	//
	// Static content.
	router.Group(func(r chi.Router) {
		r.Handle("/web/content/*", handlers.StaticFileServerHandler(http.FS(static)))
	})
	// // Error handling.
	// router.NotFound(handlers.NotFound())

	router.Get("/register", handlers.GetRegistration(agent))
	router.With(middlewares.RequireHTMX).Get("/register/discovery", handlers.RegistrationDiscovery())
	router.With(middlewares.RequireHTMX).Post("/register", handlers.ProcessRegistration(agent))
	// Front page.
	router.Get("/", handlers.Landing(agent))
	// // Access routes.
	// router.Get("/login", handlers.Login())
	// router.Group(func(r chi.Router) {
	// 	r.Use(
	// 		session.Manager.LoadAndSave,
	// 	)
	// 	r.Get("/logout", handlers.Logout())
	// })

	// // Authenticated routes.
	// router.Group(func(r chi.Router) {
	// 	r.Use(
	// 		middlewares.SetupHTMX,
	// 		middlewares.SetupElastic(),
	// 		session.Manager.LoadAndSave,
	// 		middlewares.RequireUserAuth(handler.DataAPI(), handler.AuthAPI()),
	// 	)
	// 	r.Get("/register", handler.Home())
	// })

	return router
}
