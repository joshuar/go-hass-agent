// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/components/logging"
	"github.com/joshuar/go-hass-agent/internal/components/preferences"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
)

// runNotificationsWorker will run a goroutine that is listening for
// notification messages from Home Assistant on a websocket connection. Any
// received notifications will be dipslayed on the device running the agent.
func runNotificationsWorker(ctx context.Context, headless bool, agentUI ui) {
	// Don't run if agent is running headless.
	if headless {
		return
	}

	websocket := api.NewWebsocket(ctx,
		preferences.WebsocketURL(),
		preferences.WebhookID(),
		preferences.Token())

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Connect the websocket.
			notifyCh, err := websocket.Connect(ctx)
			if err != nil {
				logging.FromContext(ctx).Warn("Failed to connect to websocket.",
					slog.Any("error", err))

				return
			}

			// Start listening on the websocket
			go func() {
				websocket.Listen()
			}()

			// Display any notifications received.
			for notification := range notifyCh {
				agentUI.DisplayNotification(&notification)
			}
		}
	}
}
