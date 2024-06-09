// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/lxzan/gws"
	"github.com/rs/zerolog/log"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

const (
	pingInterval = time.Minute
	connDeadline = 2 * pingInterval
)

const closeNormal = 1000

type websocketMsg struct {
	Type           string `json:"type"`
	WebHookID      string `json:"webhook_id,omitempty"`
	AccessToken    string `json:"access_token,omitempty"`
	ID             uint64 `json:"id,omitempty"`
	SupportConfirm bool   `json:"support_confirm,omitempty"`
}

func (m *websocketMsg) send(conn *gws.Conn) error {
	msg, err := json.Marshal(&m)
	if err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}

	err = conn.WriteMessage(gws.OpcodeText, msg)
	if err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}

	return nil
}

type websocketResponse struct {
	Result       any                   `json:"result,omitempty"`
	Error        APIError              `json:"error,omitempty"`
	Type         string                `json:"type"`
	HAVersion    string                `json:"ha_version,omitempty"`
	Notification WebsocketNotification `json:"event,omitempty"`
	ID           uint64                `json:"id,omitempty"`
	Success      bool                  `json:"success,omitempty"`
}

type WebsocketNotification struct {
	Data      any      `json:"data,omitempty"`
	Message   string   `json:"message"`
	Title     string   `json:"title,omitempty"`
	ConfirmID string   `json:"confirm_id,omitempty"`
	Target    []string `json:"target,omitempty"`
}

func (n *WebsocketNotification) GetTitle() string {
	return n.Title
}

func (n *WebsocketNotification) GetMessage() string {
	return n.Message
}

//nolint:exhaustruct
func StartWebsocket(ctx context.Context) chan *WebsocketNotification {
	var socket *gws.Conn

	notifyCh := make(chan *WebsocketNotification)

	prefs, err := preferences.Load()
	if err != nil {
		log.Error().Err(err).Msg("Could not load preferences.")
	}

	retryFunc := func() error {
		var resp *http.Response

		socket, resp, err = gws.NewClient(
			newWebsocket(prefs, notifyCh),
			&gws.ClientOption{Addr: prefs.WebsocketURL})
		if err != nil {
			return fmt.Errorf("could not connect to websocket: %w", err)
		}
		defer resp.Body.Close()

		return nil
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				if socket != nil {
					socket.WriteClose(closeNormal, nil)
				}
			default:
				err = backoff.Retry(retryFunc, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
				if err != nil {
					log.Error().Err(err).
						Msg("Could not connect to websocket.")

					return
				}

				log.Trace().Msg("Websocket connection established.")
				socket.ReadLoop()
			}
		}
	}()

	return notifyCh
}

type webSocket struct {
	notifyCh  chan *WebsocketNotification
	doneCh    chan struct{}
	token     string
	webhookID string
	nextID    uint64
}

func newWebsocket(prefs *preferences.Preferences, notifyCh chan *WebsocketNotification) *webSocket {
	websocket := &webSocket{
		notifyCh:  notifyCh,
		doneCh:    make(chan struct{}),
		token:     prefs.Token,
		webhookID: prefs.WebhookID,
		nextID:    0,
	}

	return websocket
}

//nolint:exhaustruct
func (c *webSocket) newAuthMsg() *websocketMsg {
	return &websocketMsg{
		Type:        "auth",
		AccessToken: c.token,
	}
}

//nolint:exhaustruct
func (c *webSocket) newRegistrationMsg() *websocketMsg {
	return &websocketMsg{
		Type:           "mobile_app/push_notification_channel",
		ID:             atomic.LoadUint64(&c.nextID),
		WebHookID:      c.webhookID,
		SupportConfirm: false,
	}
}

//nolint:exhaustruct
func (c *webSocket) newPingMsg() *websocketMsg {
	return &websocketMsg{
		Type: "ping",
		ID:   atomic.LoadUint64(&c.nextID),
	}
}

//revive:disable:unused-receiver
func (c *webSocket) OnError(_ *gws.Conn, err error) {
	log.Error().Err(err).
		Msg("Error on websocket")
}

func (c *webSocket) OnClose(_ *gws.Conn, err error) {
	if err.Error() != "" {
		log.Error().Err(err).Msg("Websocket connection closed with error.")
	}

	close(c.doneCh)
}

func (c *webSocket) OnPong(_ *gws.Conn, _ []byte) {
	log.Trace().Msg("Received pong on websocket")
}

func (c *webSocket) OnOpen(socket *gws.Conn) {
	log.Trace().Msg("Websocket opened.")

	go c.keepAlive(socket)
}

func (c *webSocket) OnPing(_ *gws.Conn, _ []byte) {
	log.Trace().Msg("Received ping on websocket.")
}

//nolint:cyclop,exhaustruct
func (c *webSocket) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()

	response := &websocketResponse{
		Success: true,
	}

	if err := json.Unmarshal(message.Bytes(), &response); err != nil {
		log.Error().Err(err).
			Msgf("Failed to unmarshall response %s.", message.Data.String())

		return
	}

	atomic.AddUint64(&c.nextID, 1)

	var reply *websocketMsg

	switch response.Type {
	case "event":
		c.notifyCh <- &response.Notification
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
	case "auth_required":
		log.Trace().Msg("Requesting authorisation for websocket.")

		reply = c.newAuthMsg()
	case "auth_ok":
		log.Trace().Msg("Registering app for push notifications.")

		reply = c.newRegistrationMsg()
	case "pong":
		b, err := json.Marshal(response)
		if err != nil {
			log.Error().Err(err).Msg("Unable to unmarshal.")
		}

		c.OnPong(socket, b)
	default:
		log.Warn().Msgf("Unhandled websocket response %v.", response.Type)
	}

	if reply != nil {
		err := reply.send(socket)
		if err != nil {
			log.Error().Err(err).
				Msg("Unable to send websocket message.")
		}
	}
}

func (c *webSocket) keepAlive(conn *gws.Conn) {
	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case <-c.doneCh:
			return
		case <-ticker.C:
			log.Trace().Str("runner", "websocket").
				Msg("Sending ping on websocket")

			if err := conn.SetDeadline(time.Now().Add(connDeadline)); err != nil {
				log.Error().Err(err).
					Msg("Error setting deadline on websocket.")

				return
			}

			msg := c.newPingMsg()

			if err := msg.send(conn); err != nil {
				log.Error().Err(err).
					Msg("Error sending ping.")
			}
		}
	}
}
