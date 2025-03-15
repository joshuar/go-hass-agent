// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass"
	fyneui "github.com/joshuar/go-hass-agent/internal/ui/fyneUI"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

// APIs holds various APIs that the app needs to use.
type APIs interface {
	Hass() *hass.Client
}

// appUI are the methods required for the agent to display its windows, tray
// and notifications.
type appUI interface {
	DisplayNotification(n fyneui.Notification)
	DisplayTrayIcon(ctx context.Context, cancelFunc context.CancelFunc)
	DisplayRegistrationWindow(ctx context.Context, prefs *preferences.Registration) chan bool
	Run(ctx context.Context)
}

type App struct {
	ui            appUI
	logger        *slog.Logger
	workerManager *workers.Manager
}

func New(ctx context.Context, appAPIs APIs, headless bool) *App {
	app := &App{
		logger:        logging.FromContext(ctx).WithGroup("app"),
		workerManager: workers.NewManager(ctx),
	}

	if !headless {
		app.ui = fyneui.NewFyneUI(ctx, appAPIs.Hass())
	}

	return app
}
