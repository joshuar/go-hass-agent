// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"sync"

	"github.com/joshuar/go-hass-agent/internal/hass/api"
	"github.com/rs/zerolog/log"
)

func (agent *Agent) runNotificationsWorker(ctx context.Context, options AgentOptions) {
	if options.Headless {
		log.Warn().Msg("Running headless, will not register for receiving notifications.")
		return
	}

	notifyCh := make(chan [2]string)
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
				agent.ui.DisplayNotification(n[0], n[1])
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
