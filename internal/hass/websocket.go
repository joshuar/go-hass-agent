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
	"github.com/joshuar/go-hass-agent/internal/settings"
	"github.com/rs/zerolog/log"

	"github.com/lxzan/gws"
)

const PingInterval = time.Minute

type websocketMsg struct {
	Type           string `json:"type"`
	WebHookID      string `json:"webhook_id,omitempty"`
	AccessToken    string `json:"access_token,omitempty"`
	ID             uint64 `json:"id,omitempty"`
	SupportConfirm bool   `json:"support_confirm,omitempty"`
}

type websocketResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
	Type         string `json:"type"`
	HAVersion    string `json:"ha_version,omitempty"`
	Notification struct {
		Data      interface{} `json:"data,omitempty"`
		Message   string      `json:"message"`
		Title     string      `json:"title,omitempty"`
		ConfirmID string      `json:"confirm_id,omitempty"`
		Target    []string    `json:"target,omitempty"`
	} `json:"event,omitempty"`
	ID      uint64 `json:"id,omitempty"`
	Success bool   `json:"success,omitempty"`
}

func StartWebsocket(ctx context.Context, notifyCh chan fyne.Notification, doneCh chan struct{}) {
	conn, err := tryWebsocketConnect(ctx, notifyCh, doneCh)
	if err != nil {
		log.Error().Err(err).
			Msg("Could not connect to websocket.")
		return
	}
	log.Trace().Caller().Msg("Websocket connection established.")
	go conn.ReadLoop()
}

func tryWebsocketConnect(ctx context.Context, notifyCh chan fyne.Notification, doneCh chan struct{}) (*gws.Conn, error) {
	s, err := settings.FetchFromContext(ctx)
	if err != nil {
		return nil, err
	}
	url, err := s.GetValue(settings.WebsocketURL)
	if err != nil {
		return nil, err
	}

	ctxConnect, cancelConnect := context.WithTimeout(ctx, time.Minute)
	defer cancelConnect()

	var socket *gws.Conn

	retryFunc := func() error {
		socket, _, err = gws.NewClient(NewWebsocket(ctx, notifyCh, doneCh), &gws.ClientOption{
			Addr: url,
		})
		if err != nil {
			log.Error().Err(err).
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
	doneCh     chan struct{}
	cancelFunc context.CancelFunc
	token      string
	webhookID  string
	nextID     uint64
}

func NewWebsocket(ctx context.Context, notifyCh chan fyne.Notification, doneCh chan struct{}) *WebSocket {
	s, err := settings.FetchFromContext(ctx)
	if err != nil {
		return nil
	}

	token, err := s.GetValue(settings.Token)
	if err != nil {
		return nil
	}
	webhookID, err := s.GetValue(settings.WebhookID)
	if err != nil {
		return nil
	}

	wsCtx, wsCancel := context.WithCancel(ctx)
	ws := &WebSocket{
		ReadCh:     make(chan *webSocketData),
		WriteCh:    make(chan *webSocketData),
		token:      token,
		webhookID:  webhookID,
		cancelFunc: wsCancel,
		doneCh:     doneCh,
	}
	go ws.responseHandler(wsCtx, notifyCh)
	go ws.requestHandler(wsCtx)
	return ws
}

func (c *WebSocket) OnError(socket *gws.Conn, err error) {
	log.Error().Err(err).
		Msg("Error on websocket")
	c.cancelFunc()
	close(c.doneCh)
}

func (c *WebSocket) OnClose(socket *gws.Conn, err error) {
	if err != nil {
		log.Error().Err(err).
			Msg("Error closing websocket.")
	}
	c.cancelFunc()
	close(c.doneCh)
}

func (c *WebSocket) OnPong(socket *gws.Conn, payload []byte) {
	log.Trace().Caller().Msg("Recieved pong on websocket")
}

func (c *WebSocket) OnOpen(socket *gws.Conn) {
	log.Trace().Caller().Msg("Websocket opened.")
	go func() {
		ticker := time.NewTicker(PingInterval)
		for range ticker.C {
			log.Trace().Caller().
				Msg("Sending ping on websocket")
			if err := socket.SetDeadline(time.Now().Add(2 * PingInterval)); err != nil {
				log.Error().Err(err).
					Msg("Error setting deadline on websocket.")
				return
			}
			c.WriteCh <- &webSocketData{
				conn: socket,
				data: &websocketMsg{
					Type: "ping",
					ID:   atomic.LoadUint64(&c.nextID),
				},
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
	if err := json.Unmarshal(message.Bytes(), &response); err != nil {
		log.Error().Err(err).
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
			log.Trace().Caller().Msg("Stopping websocket response handler.")
			return
		case r := <-c.ReadCh:
			if r == nil {
				return
			}
			response, ok := r.data.(*websocketResponse)
			if !ok {
				log.Warn().Msg("Websocket response is not an expected format.")
				return
			}
			atomic.AddUint64(&c.nextID, 1)
			switch response.Type {
			case "auth_required":
				log.Trace().Caller().
					Msg("Requesting authorisation for websocket.")
				c.WriteCh <- &webSocketData{
					conn: r.conn,
					data: &websocketMsg{
						Type:        "auth",
						AccessToken: c.token,
					}}
			case "auth_ok":
				// spew.Dump(response)
				log.Trace().Caller().
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
						log.Warn().
							Msg("id_reuse error, attempting manual increment.")
						atomic.AddUint64(&c.nextID, 1)
					}
				}
			case "event":
				notifyCh <- *fyne.NewNotification(response.Notification.Title, response.Notification.Message)
			case "pong":
				b, err := json.Marshal(response)
				if err != nil {
					log.Error().Err(err).Msg("Unable to unmarshal.")
				}
				c.OnPong(r.conn, b)
			default:
				log.Warn().Caller().Msgf("Unhandled websocket response %v.", r)
			}
		}
	}
}

func (c *WebSocket) requestHandler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Trace().Caller().Msg("Stopping websocket request handler.")
			return
		case m := <-c.WriteCh:
			msg, err := json.Marshal(&m.data)
			if err != nil {
				log.Error().Err(err).
					Msg("Unable to marshal websocket message.")
			}
			err = m.conn.WriteMessage(gws.OpcodeText, msg)
			if err != nil {
				log.Error().Err(err).
					Msg("Unable to send websocket message.")
			}
		}
	}
}
