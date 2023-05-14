// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

func (agent *Agent) runNotificationsWorker(ctx context.Context) {
	agentConfig, err := config.FetchConfigFromContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Could not retrieve valid config from context.")
		return
	}

	go func() {
		for n := range agentConfig.NotifyCh {
			agent.app.SendNotification(&fyne.Notification{
				Title:   n.Title,
				Content: n.Content,
			})
		}
	}()

	doneCh := make(chan struct{})

	hass.StartWebsocket(ctx, doneCh)
	for {
		select {
		case <-doneCh:
			doneCh = make(chan struct{})
			hass.StartWebsocket(ctx, doneCh)
		}
	}
}
