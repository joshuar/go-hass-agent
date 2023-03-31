package agent

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"github.com/cenkalti/backoff/v4"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func (agent *Agent) runNotificationsWorker() {

	url := agent.config.WebSocketURL

	reconnect := make(chan bool)
	defer close(reconnect)

	for {
		log.Debug().Caller().Msgf("Using %s for websocket connection for notification access.", url)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		var conn *websocket.Conn
		var err error
		retryFunc := func() error {
			conn, _, err = websocket.Dial(ctx, url, nil)
			if err != nil {
				log.Debug().Caller().Msgf("Unable to connect to websocket: %v", err)
				return err
			}
			return nil
		}
		err = backoff.Retry(retryFunc, backoff.NewExponentialBackOff())
		if err != nil {
			log.Debug().Caller().Msgf("Failed to connect to websocket: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")
		go agent.webSocketNotifications(conn, reconnect)
		<-reconnect
	}
}

func (agent *Agent) webSocketNotifications(conn *websocket.Conn, reconnect chan bool) {
	accessToken := agent.config.token
	webhookID := agent.config.webhookID

	for {
		ctx := context.Background()
		response := &hass.WebsocketResponse{
			Success: true,
		}
		err := wsjson.Read(ctx, conn, &response)
		if err != nil {
			log.Warn().Msg(err.Error())
			ctx.Done()
			reconnect <- true
			return
		} else {
			switch response.Type {
			case "auth_required":
				log.Debug().Caller().Msg("Requesting authorisation for websocket.")
				reqCtx, cancel := context.WithTimeout(ctx, time.Minute)
				defer cancel()
				err = wsjson.Write(reqCtx, conn, struct {
					Type        string `json:"type"`
					AccessToken string `json:"access_token"`
				}{
					Type:        "auth",
					AccessToken: accessToken,
				})
				logging.CheckError(err)
			case "auth_ok":
				log.Debug().Caller().Msg("Registering app for push notifications.")
				reqCtx, cancel := context.WithTimeout(ctx, time.Minute)
				defer cancel()
				err = wsjson.Write(reqCtx, conn, &struct {
					Type           string `json:"type"`
					ID             int    `json:"id"`
					WebHookID      string `json:"webhook_id"`
					SupportConfirm bool   `json:"support_confirm"`
				}{
					Type:           "mobile_app/push_notification_channel",
					ID:             1,
					WebHookID:      webhookID,
					SupportConfirm: false,
				})
				logging.CheckError(err)
			case "result":
				if !response.Success {
					log.Error().Msgf("Recieved error on websocket, %s: %s.", response.Error.Code, response.Error.Message)
					reconnect <- true
				}
			case "event":
				agent.App.SendNotification(&fyne.Notification{
					Title:   response.Notification.Title,
					Content: response.Notification.Message,
				})
			default:
				log.Debug().Caller().Msgf("Received unhandled response %v", response)
			}
		}
	}
}
