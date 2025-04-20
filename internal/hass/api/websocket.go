// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package api

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
)

const (
	pingInterval = time.Minute
	connDeadline = 2 * pingInterval

	closeNormal = 1000
)

type webSocketRequest struct {
	Type           string `json:"type"`
	WebHookID      string `json:"webhook_id,omitempty"`
	AccessToken    string `json:"access_token,omitempty"`
	ID             uint64 `json:"id,omitempty"`
	SupportConfirm bool   `json:"support_confirm,omitempty"`
}

func (m *webSocketRequest) send(conn *gws.Conn) error {
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

type Websocket struct {
	NotifyCh  chan WebsocketNotification
	logger    *slog.Logger
	socket    *gws.Conn
	token     string
	webhookID string
	url       string
	nextID    uint64
}

func (c *Websocket) newAuthMsg() *webSocketRequest {
	return &webSocketRequest{
		Type:        "auth",
		AccessToken: c.token,
	}
}

func (c *Websocket) newRegistrationMsg() *webSocketRequest {
	return &webSocketRequest{
		Type:           "mobile_app/push_notification_channel",
		ID:             atomic.LoadUint64(&c.nextID),
		WebHookID:      c.webhookID,
		SupportConfirm: false,
	}
}

func (c *Websocket) newPingMsg() *webSocketRequest {
	return &webSocketRequest{
		Type: "ping",
		ID:   atomic.LoadUint64(&c.nextID),
	}
}

//revive:disable:unused-receiver
func (c *Websocket) OnError(_ *gws.Conn, err error) {
	c.logger.Error("Error on websocket.", slog.Any("error", err))
}

func (c *Websocket) OnClose(_ *gws.Conn, err error) {
	if err != nil && err.Error() != "gws: close normal" {
		c.logger.Warn("Websocket connection closed with error.", slog.Any("error", err))
	}
}

func (c *Websocket) OnPong(_ *gws.Conn, _ []byte) {}

func (c *Websocket) OnOpen(socket *gws.Conn) {
	go c.keepAlive(socket)
}

func (c *Websocket) OnPing(_ *gws.Conn, _ []byte) {}

//nolint:cyclop
func (c *Websocket) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()

	response := &websocketResponse{
		Success: true,
	}

	if err := json.Unmarshal(message.Bytes(), &response); err != nil {
		c.logger.Error("Failed to unmarshal response.",
			slog.Any("error", err),
			slog.Any("raw_response", message.Data.Bytes()))

		return
	}

	atomic.AddUint64(&c.nextID, 1)

	var reply *webSocketRequest

	switch response.Type {
	case "event":
		c.NotifyCh <- response.Notification
	case "result":
		if !response.Success {
			c.logger.Error("Received error on websocket.",
				slog.Any("error", response.Error))

			if *response.Error.Code == "id_reuse" {
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
			c.logger.Error("Unable to unmarshal pong response.", slog.Any("error", err))
		}

		c.OnPong(socket, b)
	default:
		c.logger.Warn("Unhandled websocket response type.", slog.String("type", response.Type))
	}

	if reply != nil {
		err := reply.send(socket)
		if err != nil {
			c.logger.Error("Unable to send websocket message.", slog.Any("error", err))
		}
	}
}

func (c *Websocket) keepAlive(conn *gws.Conn) {
	ticker := time.NewTicker(pingInterval)

	for range ticker.C {
		if err := conn.SetDeadline(time.Now().Add(connDeadline)); err != nil {
			c.logger.Error("Could not set deadline on websocket.", slog.Any("error", err))

			return
		}

		msg := c.newPingMsg()

		if err := msg.send(conn); err != nil {
			c.logger.Error("Could not send ping message.", slog.Any("error", err))
		}
	}
}

// NewWebsocket creates a new websocket object using the given websocket url,
// webhookid and token.
func NewWebsocket(ctx context.Context, url, webhookID, token string) *Websocket {
	return &Websocket{
		token:     token,
		webhookID: webhookID,
		url:       url,
	}
}

// Connect establishes a connection on the websocket. It implements an
// exponential backoff method for retries on the event of connection failures.
// Once a connection is established, it sets up the notification channel for
// receiving notifications from Home Assistant.
func (c *Websocket) Connect(ctx context.Context) (chan WebsocketNotification, error) {
	var (
		err    error
		socket *gws.Conn
	)

	retryFunc := func() error {
		var resp *http.Response

		socket, resp, err = gws.NewClient(c, &gws.ClientOption{Addr: c.url})
		if err != nil {
			return fmt.Errorf("could not establish connection: %w", err)
		}
		defer resp.Body.Close()

		return nil
	}

	err = backoff.Retry(retryFunc, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	c.socket = socket
	c.NotifyCh = make(chan WebsocketNotification)

	go func() {
		<-ctx.Done()
		c.socket.WriteClose(closeNormal, []byte(`normal close`))
	}()

	return c.NotifyCh, nil
}

// Listen will listen for notifications from Home Assistant and pass these
// through the created channel, for the agent to consume. If the socket is
// closed, the channel is also closed.
func (c *Websocket) Listen() {
	defer close(c.NotifyCh)
	c.logger.Debug("Listening on websocket.")
	c.socket.ReadLoop()
}
