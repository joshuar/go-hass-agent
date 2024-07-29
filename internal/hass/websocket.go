// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/lxzan/gws"

	"github.com/joshuar/go-hass-agent/internal/logging"
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

//nolint:exhaustruct,govet
func StartWebsocket(ctx context.Context) (chan *WebsocketNotification, error) {
	var socket *gws.Conn

	notifyCh := make(chan *WebsocketNotification)

	url, err := preferences.ContextGetWebsocketURL(ctx)
	if err != nil {
		close(notifyCh)

		return notifyCh, fmt.Errorf("unable to start websocket connection: %w", err)
	}

	retryFunc := func() error {
		var resp *http.Response

		ws, err := newWebsocket(ctx, notifyCh)
		if err != nil {
			return fmt.Errorf("could not connect to websocket: %w", err)
		}

		socket, resp, err = gws.NewClient(ws, &gws.ClientOption{Addr: url})
		if err != nil {
			return fmt.Errorf("could not connect to websocket: %w", err)
		}
		defer resp.Body.Close()

		return nil
	}

	go func() {
		defer close(notifyCh)

		for {
			select {
			case <-ctx.Done():
				if socket != nil {
					socket.WriteClose(closeNormal, nil)
				}
			default:
				err = backoff.Retry(retryFunc, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
				if err != nil {
					logging.FromContext(ctx).Error("Could not connect to websocket.", "error", err.Error())

					return
				}

				logging.FromContext(ctx).Log(ctx, logging.LevelTrace, "Websocket connection established.")
				socket.ReadLoop()
			}
		}
	}()

	return notifyCh, nil
}

type webSocket struct {
	notifyCh  chan *WebsocketNotification
	doneCh    chan struct{}
	logger    *slog.Logger
	token     string
	webhookID string
	nextID    uint64
}

func newWebsocket(ctx context.Context, notifyCh chan *WebsocketNotification) (*webSocket, error) {
	token, err := preferences.ContextGetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create websocket: %w", err)
	}

	webhookid, err := preferences.ContextGetWebhookID(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create websocket: %w", err)
	}

	websocket := &webSocket{
		notifyCh:  notifyCh,
		doneCh:    make(chan struct{}),
		token:     token,
		webhookID: webhookid,
		nextID:    0,
		logger:    logging.FromContext(ctx),
	}

	return websocket, nil
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
	c.logger.Error("Error on websocket.", "error", err.Error())
}

func (c *webSocket) OnClose(_ *gws.Conn, err error) {
	if err != nil {
		c.logger.Warn("Websocket connection closed with error.", "error", err.Error())
	}

	close(c.doneCh)
}

func (c *webSocket) OnPong(_ *gws.Conn, _ []byte) {}

func (c *webSocket) OnOpen(socket *gws.Conn) {
	c.logger.Debug("Websocket opened.")

	go c.keepAlive(socket)
}

func (c *webSocket) OnPing(_ *gws.Conn, _ []byte) {}

//nolint:cyclop,exhaustruct
func (c *webSocket) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()

	response := &websocketResponse{
		Success: true,
	}

	if err := json.Unmarshal(message.Bytes(), &response); err != nil {
		c.logger.Error("Failed to unmarshal response.", slog.Any("error", err), slog.Any("raw_response", message.Data.Bytes()))

		return
	}

	atomic.AddUint64(&c.nextID, 1)

	var reply *websocketMsg

	switch response.Type {
	case "event":
		c.notifyCh <- &response.Notification
	case "result":
		if !response.Success {
			c.logger.Error("Received error on websocket.", "code", response.Error.Code, "error", response.Error.Message)

			if response.Error.Code == "id_reuse" {
				c.logger.Warn("Detected message ID reuse, attempting manual increment.")
				atomic.AddUint64(&c.nextID, 1)
			}
		}
	case "auth_required":
		c.logger.Debug("Requesting authorisation for websocket.")

		reply = c.newAuthMsg()
	case "auth_ok":
		c.logger.Debug("Registering app for push notifications.")

		reply = c.newRegistrationMsg()
	case "pong":
		b, err := json.Marshal(response)
		if err != nil {
			c.logger.Error("Unable to unmarshal pong response.", "error", err.Error())
		}

		c.OnPong(socket, b)
	default:
		c.logger.Warn("Unhandled websocket response type.", "type", response.Type)
	}

	if reply != nil {
		err := reply.send(socket)
		if err != nil {
			c.logger.Error("Unable to send websocket message.", "error", err.Error())
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
			if err := conn.SetDeadline(time.Now().Add(connDeadline)); err != nil {
				c.logger.Error("Could not set deadline on websocket.", "error", err.Error())

				return
			}

			msg := c.newPingMsg()

			if err := msg.send(conn); err != nil {
				c.logger.Error("Could not send ping message.", "error", err.Error())
			}
		}
	}
}
