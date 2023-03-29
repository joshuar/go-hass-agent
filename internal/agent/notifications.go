package agent

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func (agent *Agent) runNotificationsWorker() {
	ctx := context.Background()

	url := agent.config.WebSocketURL
	accessToken := agent.config.token
	webhookID := agent.config.webhookID

	for {
		log.Debug().Caller().Msgf("Using %s for websocket connection for notification access.", url)
		conn, _, err := websocket.Dial(ctx, url, nil)
		if err != nil {
			log.Warn().Msgf("Could not connect websocket: %v. Will retry.", err)
			time.Sleep(time.Millisecond * time.Duration(2000))
		} else {
			defer conn.Close(websocket.StatusNormalClosure, "")
			for {
				response := &hass.WebsocketResponse{
					Success: true,
				}
				err = wsjson.Read(ctx, conn, &response)
				if err != nil {
					log.Warn().Msg(err.Error())
					conn.Close(websocket.StatusInternalError, "closing connection")
					return
				}
				switch response.Type {
				case "auth_required":
					log.Debug().Caller().Msg("Requesting authorisation for websocket.")
					err = wsjson.Write(ctx, conn, struct {
						Type        string `json:"type"`
						AccessToken string `json:"access_token"`
					}{
						Type:        "auth",
						AccessToken: accessToken,
					})
					logging.CheckError(err)
				case "auth_ok":
					log.Debug().Caller().Msg("Registering app for push notifications.")
					err = wsjson.Write(ctx, conn, &struct {
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
}
