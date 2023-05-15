// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"github.com/cenkalti/backoff/v4"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/rs/zerolog/log"

	"github.com/lxzan/gws"
)

const PingInterval = time.Minute

type websocketMsg struct {
	Type           string `json:"type"`
	ID             uint64 `json:"id,omitempty"`
	WebHookID      string `json:"webhook_id,omitempty"`
	SupportConfirm bool   `json:"support_confirm,omitempty"`
	AccessToken    string `json:"access_token,omitempty"`
}

type websocketResponse struct {
	Type    string `json:"type"`
	Success bool   `json:"success,omitempty"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
	ID           uint64      `json:"id,omitempty"`
	Result       interface{} `json:"result,omitempty"`
	HAVersion    string      `json:"ha_version,omitempty"`
	Notification struct {
		Message   string      `json:"message"`
		Title     string      `json:"title,omitempty"`
		Target    []string    `json:"target,omitempty"`
		Data      interface{} `json:"data,omitempty"`
		ConfirmID string      `json:"confirm_id,omitempty"`
	} `json:"event,omitempty"`
}

func StartWebsocket(ctx context.Context, notifyCh chan fyne.Notification, doneCh chan struct{}) {
	conn, err := tryWebsocketConnect(ctx, notifyCh, doneCh)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to websocket.")
		return
	}
	log.Debug().Caller().Msg("Websocket connection established.")
	go conn.ReadLoop()
}

func tryWebsocketConnect(ctx context.Context, notifyCh chan fyne.Notification, doneCh chan struct{}) (*gws.Conn, error) {
	agentConfig, err := config.FetchConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}
	url := agentConfig.WebSocketURL

	ctxConnect, cancelConnect := context.WithTimeout(ctx, time.Minute)
	defer cancelConnect()

	var socket *gws.Conn

	retryFunc := func() error {
		socket, _, err = gws.NewClient(NewWebsocket(ctx, notifyCh, doneCh), &gws.ClientOption{
			Addr: url,
		})
		if err != nil {
			log.Debug().Err(err).Caller().
				Msg("Could not connect to websocket.")
			return err
		}
		return nil
	}
	err = backoff.Retry(retryFunc, backoff.WithContext(backoff.NewExponentialBackOff(), ctxConnect))
	if err != nil {
		cancelConnect()
		return nil, err
	}
	return socket, nil
}

type webSocketData struct {
	conn *gws.Conn
	data interface{}
}

type WebSocket struct {
	ReadCh     chan *webSocketData
	WriteCh    chan *webSocketData
	token      string
	webhookID  string
	doneCh     chan struct{}
	cancelFunc context.CancelFunc
	nextID     uint64
}

func NewWebsocket(ctx context.Context, notifyCh chan fyne.Notification, doneCh chan struct{}) *WebSocket {
	agentConfig, err := config.FetchConfigFromContext(ctx)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Could not retrieve valid config from context.")
		return nil
	}
	wsCtx, wsCancel := context.WithCancel(ctx)
	ws := &WebSocket{
		ReadCh:     make(chan *webSocketData),
		WriteCh:    make(chan *webSocketData),
		token:      agentConfig.Token,
		webhookID:  agentConfig.WebhookID,
		cancelFunc: wsCancel,
		doneCh:     doneCh,
	}
	go ws.responseHandler(wsCtx, notifyCh)
	go ws.requestHandler(wsCtx)
	return ws
}

func (c *WebSocket) OnError(socket *gws.Conn, err error) {
	log.Debug().Caller().Err(err).
		Msg("Error on websocket")
	c.cancelFunc()
	close(c.doneCh)
}

func (c *WebSocket) OnClose(socket *gws.Conn, code uint16, reason []byte) {
	log.Debug().Caller().
		Msgf("onclose: code=%d, payload=%s\n", code, string(reason))
	c.cancelFunc()
	close(c.doneCh)
}

func (c *WebSocket) OnPong(socket *gws.Conn, payload []byte) {
	log.Debug().Caller().Msg("Recieved pong on websocket")
}

func (c *WebSocket) OnOpen(socket *gws.Conn) {
	log.Debug().Caller().Msg("Websocket opened.")
	go func() {
		ticker := time.NewTicker(PingInterval)
		done := make(chan bool)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				log.Debug().Caller().
					Msg("Sending ping on websocket")
				err := socket.SetDeadline(time.Now().Add(2 * PingInterval))
				if err != nil {
					log.Debug().Err(err).
						Msg("Error setting deadline on websocket.")
				}
				c.WriteCh <- &webSocketData{
					conn: socket,
					data: &websocketMsg{
						Type: "ping",
						ID:   atomic.LoadUint64(&c.nextID),
					},
				}
			}
		}
	}()
}

func (c *WebSocket) OnPing(socket *gws.Conn, payload []byte) {
}

func (c *WebSocket) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()
	response := &websocketResponse{
		Success: true,
	}
	err := json.Unmarshal(message.Bytes(), &response)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msgf("Failed to unmarshall response %s.", message.Data.String())
		return
	}
	c.ReadCh <- &webSocketData{
		conn: socket,
		data: response,
	}
}

func (c *WebSocket) responseHandler(ctx context.Context, notifyCh chan fyne.Notification) {
	for {
		select {
		case <-ctx.Done():
			log.Debug().Caller().Msg("Stopping websocket response handler.")
			return
		case r := <-c.ReadCh:
			if r == nil {
				return
			}
			response := r.data.(*websocketResponse)
			atomic.AddUint64(&c.nextID, 1)
			switch response.Type {
			case "auth_required":
				log.Debug().Caller().
					Msg("Requesting authorisation for websocket.")
				c.WriteCh <- &webSocketData{
					conn: r.conn,
					data: &websocketMsg{
						Type:        "auth",
						AccessToken: c.token,
					}}
			case "auth_ok":
				// spew.Dump(response)
				log.Debug().Caller().
					Msg("Registering app for push notifications.")
				c.WriteCh <- &webSocketData{
					conn: r.conn,
					data: &websocketMsg{
						Type:           "mobile_app/push_notification_channel",
						ID:             atomic.LoadUint64(&c.nextID),
						WebHookID:      c.webhookID,
						SupportConfirm: false,
					}}
			case "result":
				if !response.Success {
					log.Error().
						Msgf("Recieved error on websocket, %s: %s.", response.Error.Code, response.Error.Message)
					if response.Error.Code == "id_reuse" {
						log.Debug().Caller().
							Msg("id_reuse error, attempting manual increment.")
						atomic.AddUint64(&c.nextID, 1)
					}
				}
			case "event":
				notifyCh <- *fyne.NewNotification(response.Notification.Title, response.Notification.Message)
			case "pong":
				b, err := json.Marshal(response)
				if err != nil {
					log.Debug().Err(err).Msg("Unable to unmarshal.")
				}
				c.OnPong(r.conn, b)
			default:
				log.Debug().Caller().Msgf("Received unhandled response %v", r)
			}
		}
	}
}

func (c *WebSocket) requestHandler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug().Caller().Msg("Stopping websocket request handler.")
			return
		case m := <-c.WriteCh:
			msg, err := json.Marshal(&m.data)
			if err != nil {
				log.Debug().Caller().Err(err).
					Msg("Unable to marshal message.")
			}
			err = m.conn.WriteMessage(gws.OpcodeText, msg)
			if err != nil {
				log.Debug().Caller().Err(err).
					Msg("Unable to send message.")
			}
		}
	}
}
