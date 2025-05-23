// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"log/slog"

	slogctx "github.com/veqryn/slog-context"

	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/ui"
	"github.com/joshuar/go-hass-agent/internal/workers"
)

// APIs holds various APIs that the app needs to use.
type APIs interface {
	Hass() *hass.Client
}

// appUI are the methods required for the agent to display its windows, tray
// and notifications.
type appUI interface {
	DisplayNotification(n ui.Notification)
	DisplayTrayIcon(ctx context.Context, cancelFunc context.CancelFunc)
	DisplayRegistrationWindow(ctx context.Context, prefs *preferences.Registration) chan bool
	Run(ctx context.Context)
}

type App struct {
	ui            appUI
	workerManager *workers.Manager
}

func New(ctx context.Context, appAPIs APIs, headless bool) *App {
	ctx = slogctx.NewCtx(ctx, slog.Default())
	app := &App{
		workerManager: workers.NewManager(ctx),
	}

	if !headless {
		app.ui = ui.NewFyneUI(ctx, appAPIs.Hass())
	}

	return app
}
