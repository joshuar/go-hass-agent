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

func (agent *Agent) runNotificationsWorker(ctx context.Context, options Options) {
	if options.Headless {
		log.Warn().Msg("Will not send notifications as there is no supported windowing environment.")
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
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Stopping websocket.")
				return
			default:
				api.StartWebsocket(ctx, notifyCh)
			}
		}
	}()

	wg.Wait()
}
