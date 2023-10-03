// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/joshuar/go-hass-agent/internal/agent/config"
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

func StartWebsocket(ctx context.Context, settings Agent, notifyCh chan [2]string, doneCh chan struct{}) {
	var websocketURL string
	if err := settings.GetConfig(config.PrefWebsocketURL, &websocketURL); err != nil {
		log.Warn().Err(err).Msg("Could not retrieve websocket URL from config.")
		return
	}
	var socket *gws.Conn
	var err error

	retryFunc := func() error {
		var resp *http.Response
		socket, resp, err = gws.NewClient(
			newWebsocket(ctx, settings, notifyCh, doneCh),
			&gws.ClientOption{Addr: websocketURL})
		if err != nil {
			log.Error().Err(err).
				Msg("Could not connect to websocket.")
			return err
		}
		defer resp.Body.Close()
		return nil
	}
	err = backoff.Retry(retryFunc, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
	if err != nil {
		log.Error().Err(err).
			Msg("Could not connect to websocket.")
		return
	}
	log.Trace().Caller().Msg("Websocket connection established.")
	go socket.ReadLoop()
}

type webSocketData struct {
	conn *gws.Conn
	data interface{}
}

type WebSocket struct {
	ReadCh    chan *webSocketData
	WriteCh   chan *webSocketData
	doneCh    chan struct{}
	token     string
	webhookID string
	nextID    uint64
}

func newWebsocket(ctx context.Context, settings Agent, notifyCh chan [2]string, doneCh chan struct{}) *WebSocket {
	var token, webhookID string
	if err := settings.GetConfig(config.PrefToken, &token); err != nil {
		log.Warn().Err(err).Msg("Could not retrieve token from config.")
		return nil
	}
	if err := settings.GetConfig(config.PrefWebhookID, &webhookID); err != nil {
		log.Warn().Err(err).Msg("Could not retrieve webhookID from config.")
		return nil
	}

	ws := &WebSocket{
		ReadCh:    make(chan *webSocketData),
		WriteCh:   make(chan *webSocketData),
		token:     token,
		webhookID: webhookID,
		doneCh:    doneCh,
	}
	go ws.responseHandler(ctx, notifyCh)
	go ws.requestHandler(ctx)
	return ws
}

func (c *WebSocket) OnError(socket *gws.Conn, err error) {
	log.Error().Err(err).
		Msg("Error on websocket")
	c.doneCh <- struct{}{}
}

func (c *WebSocket) OnClose(socket *gws.Conn, err error) {
	log.Debug().Err(err).Msg("Websocket connection closed.")
	c.doneCh <- struct{}{}
	close(c.doneCh)
}

func (c *WebSocket) OnPong(socket *gws.Conn, payload []byte) {
	log.Trace().Caller().Msg("Received pong on websocket")
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

func (c *WebSocket) responseHandler(ctx context.Context, notifyCh chan [2]string) {
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
						Msgf("Received error on websocket, %s: %s.", response.Error.Code, response.Error.Message)
					if response.Error.Code == "id_reuse" {
						log.Warn().
							Msg("id_reuse error, attempting manual increment.")
						atomic.AddUint64(&c.nextID, 1)
					}
				}
			case "event":
				notifyCh <- [2]string{response.Notification.Title, response.Notification.Message}
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
