// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/rs/zerolog/log"
)

func (agent *Agent) runNotificationsWorker(ctx context.Context, options AgentOptions) {
	if options.Headless {
		log.Warn().Msg("Running headless, will not register for recieving notifications.")
		return
	}

	notifyCh := make(chan fyne.Notification)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopping notification handler.")
				return
			case n := <-notifyCh:
				agent.UI.DisplayNotification(n.Title, n.Content)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		restartCh := make(chan struct{})
		api.StartWebsocket(ctx, agent, notifyCh, restartCh)
		for range restartCh {
			log.Debug().Msg("Restarting websocket connection.")
			api.StartWebsocket(ctx, agent, notifyCh, restartCh)
		}
	}()

	wg.Wait()
}
