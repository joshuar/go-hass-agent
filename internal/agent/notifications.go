package agent

import (
	"context"

	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/joshuar/go-hass-agent/internal/logging"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func (agent *Agent) runNotificationsWorker() {
	ctx := context.Background()
	// WithTimeout(context.Background(), time.Minute)
	// defer cancel()

	url := agent.config.WebSocketURL
	accessToken := agent.config.token
	webhookID := agent.config.webhookID

	log.Debug().Msgf("Establishing websocket connection to %s", url)
	conn, _, err := websocket.Dial(ctx, url, nil)
	logging.CheckError(err)
	defer conn.Close(websocket.StatusInternalError, "error on websocket")

	response := &hass.WebsocketResponse{
		Success: true,
	}
	for {
		err = wsjson.Read(ctx, conn, &response)
		logging.CheckError(err)
		switch response.Type {
		case "auth_required":
			log.Debug().Msg("Requesting authorisation for websocket.")
			err = wsjson.Write(ctx, conn, struct {
				Type        string `json:"type"`
				AccessToken string `json:"access_token"`
			}{
				Type:        "auth",
				AccessToken: accessToken,
			})
			logging.CheckError(err)
		case "auth_ok":
			log.Debug().Msg("Registering app for push notifications.")
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
				log.Error().Msgf("Recieved error on websocket, %s: %s", response.Error.Code, response.Error.Message)
			}
		case "event":
			log.Debug().Msgf("Received notification event with message %s", response.Notification.Message)
		// err = wsjson.Write(ctx, conn, struct {
		// 	Type      string `json:"type"`
		// 	WebhookID string `json:"webhook_id"`
		// 	ConfirmID string `json:"confirm_id"`
		// 	ID        int    `json:"id"`
		// }{
		// 	Type:      "mobile_app/push_notification_confirm",
		// 	WebhookID: webhookID,
		// 	ConfirmID: response.Notification.ConfirmID,
		// 	ID:        response.ID + 1,
		// })
		// logging.CheckError(err)
		default:
			log.Debug().Msgf("Received unhandled response %v", response)
		}
	}
	// if res.Body.Read() != "auth_ok" {
	// 	log.Error().Msgf("Could not authenticate websocket: %v", res)
	// }

}
