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

	url := agent.config.WebSocketURL
	ctxNotifications, cancelNotifications := context.WithCancel(context.Background())

	// go agent.webSocketNotifications(conn, reconnect)

	for {
		select {
		case <-agent.done:
			log.Debug().Caller().Msg("Closing notifications worker.")
			cancelNotifications()
			return
		default:
			ws := hass.NewWebsocket(ctxNotifications, url)
			if ws == nil {
				log.Debug().Caller().
					Msgf("No websocket connection made.")
				cancelNotifications()
				return
			} else {
				data := make(chan hass.WebsocketResponse)
				defer close(data)
				go ws.Read(data)
				agent.handleNotifications(data, ws)
			}
		}
	}
}

func (agent *Agent) handleNotifications(response chan hass.WebsocketResponse, ws *hass.HassWebsocket) {
	accessToken := agent.config.token
	webhookID := agent.config.webhookID

	for {
		select {
		case <-agent.done:
			log.Debug().Caller().Msg("Stopping handling notifications.")
			return
		case r := <-response:
			switch r.Type {
			case "auth_required":
				log.Debug().Caller().Msg("Requesting authorisation for websocket.")
				err := ws.Write(struct {
					Type        string `json:"type"`
					AccessToken string `json:"access_token"`
				}{
					Type:        "auth",
					AccessToken: accessToken,
				})
				logging.CheckError(err)
			case "auth_ok":
				log.Debug().Caller().Msg("Registering app for push notifications.")
				err := ws.Write(struct {
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
				if !r.Success {
					log.Error().Msgf("Recieved error on websocket, %s: %s.", r.Error.Code, r.Error.Message)
					// reconnect <- true
				}
			case "event":
				agent.App.SendNotification(&fyne.Notification{
					Title:   r.Notification.Title,
					Content: r.Notification.Message,
				})
			default:
				log.Debug().Caller().Msgf("Received unhandled response %v", response)
			}
		}
	}
}

func (agent *Agent) webSocketNotifications(conn *websocket.Conn, reconnect chan bool) {
	accessToken := agent.config.token
	webhookID := agent.config.webhookID

	for {
		select {
		case <-agent.done:
			log.Debug().Caller().
				Msg("Stopping listening for notifications.")
			return
		default:
			ctx := context.Background()
			response := &hass.WebsocketResponse{
				Success: true,
			}
			err := wsjson.Read(ctx, conn, &response)
			if err != nil {
				log.Warn().Msg(err.Error())
				// ctx.Done()
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
}
