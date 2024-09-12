// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/joshuar/go-hass-agent/internal/preferences"
	"github.com/joshuar/go-hass-agent/internal/scripts"
)

func newScriptsController(ctx context.Context) SensorController {
	appID := preferences.AppIDFromContext(ctx)
	scriptPath := filepath.Join(xdg.ConfigHome, appID, "scripts")

	scriptController, err := scripts.NewScriptsController(ctx, scriptPath)
	if err != nil {
		logging.FromContext(ctx).Error("Could not set up scripts controller.", slog.Any("error", err))

		return nil
	}

	return scriptController
}
