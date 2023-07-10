// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"

	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

func (agent *Agent) runNotificationsWorker(ctx context.Context) {
	doneCh := make(chan struct{})
	notifyCh := make(chan fyne.Notification)

	hass.StartWebsocket(ctx, notifyCh, doneCh)
	for {
		select {
		case <-doneCh:
			doneCh = make(chan struct{})
			hass.StartWebsocket(ctx, notifyCh, doneCh)
		case <-ctx.Done():
			log.Debug().Msg("Stopping notification handler.")
			return
		case n := <-notifyCh:
			agent.app.SendNotification(&fyne.Notification{
				Title:   n.Title,
				Content: n.Content,
			})
		}
	}
}
