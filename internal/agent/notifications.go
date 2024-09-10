// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"log/slog"

	"github.com/joshuar/go-hass-agent/internal/hass"
)

// runNotificationsWorker will run a goroutine that is listening for
// notification messages from Home Assistant on a websocket connection. Any
// received notifications will be dipslayed on the device running the agent.
func (agent *Agent) runNotificationsWorker(ctx context.Context) {
	// Don't run if agent is running headless.
	if agent.headless {
		return
	}

	websocket := hass.NewWebsocket(ctx,
		agent.prefs.WebsocketURL(),
		agent.prefs.WebhookID(),
		agent.prefs.Token())

	for {
		select {
		case <-ctx.Done():
			agent.logger.Debug("Stopping notifications worker.")

			return
		default:
			// Connect the websocket.
			notifyCh, err := websocket.Connect(ctx)
			if err != nil {
				agent.logger.Warn("Failed to connect to websocket.", slog.Any("error", err))

				return
			}

			// Start listening on the websocket
			go func() {
				websocket.Listen()
			}()

			// Display any notifications received.
			for notification := range notifyCh {
				agent.ui.DisplayNotification(&notification)
			}
		}
	}
}
